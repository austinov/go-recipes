package cmetal

import (
	"fmt"
	"log"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/austinov/go-recipes/slack-bot/config"
	"github.com/austinov/go-recipes/slack-bot/dao"
	"github.com/austinov/go-recipes/slack-bot/loader"
)

type cmetalBand struct {
	Id   string
	Name string
}

type CMetalLoader struct {
	cfg     config.CMetalConfig
	bands   chan cmetalBand
	events  chan dao.Event
	done    chan struct{}
	breaker *Breaker
}

func New(cfg config.CMetalConfig) loader.Loader {
	breakTriggers := make([]Trigger, 3)
	breakTriggers[0] = Trigger{
		ErrorKind:  "HTTP",
		ErrorLimit: 10,
		Callback: func(kind string, err error) {
			log.Printf("Breaker triggered for %s: %#v\n", kind, err)
		},
	}
	breakTriggers[1] = Trigger{
		ErrorKind:  "APP",
		ErrorLimit: 1,
		Callback: func(kind string, err error) {
			log.Fatalf("Breaker triggered for %s: %#v\n", kind, err)
		},
	}
	breakTriggers[2] = Trigger{
		ErrorKind:  "PARSE",
		ErrorLimit: 1,
		Callback: func(kind string, err error) {
			log.Fatalf("Breaker triggered for %s: %#v\n", kind, err)
		},
	}
	return &CMetalLoader{
		cfg:     cfg,
		done:    make(chan struct{}),
		breaker: NewBreaker(breakTriggers),
	}
}

func (l *CMetalLoader) Start() error {
	if err := l.do(); err != nil {
		return err
	}
	if l.cfg.Frequency == 0 {
		return nil
	}

	ticker := time.NewTicker(l.cfg.Frequency)
	for {
		select {
		case <-ticker.C:
			if err := l.do(); err != nil {
				return err
			}
		case <-l.done:
			return nil
		}
	}
}

func (l *CMetalLoader) Stop() {
	close(l.done)
}

func (l *CMetalLoader) do() error {
	var wg sync.WaitGroup

	wg.Add(1)
	bands := make(chan interface{}, l.cfg.NumLoaders)
	runWorkers(&wg, nil, bands, 1, l.loadBands)

	wg.Add(1)
	bandEvents := make(chan interface{}, l.cfg.NumSavers)
	runWorkers(&wg, bands, bandEvents, l.cfg.NumLoaders, l.loadBandEvents)

	wg.Add(1)
	runWorkers(&wg, bandEvents, nil, l.cfg.NumSavers, l.saveBandEvents)

	wg.Wait()

	return nil
}

// loadBands loads bands without events and put them into outBands channel
// to load the events these bands.
func (l *CMetalLoader) loadBands(ignore <-chan interface{}, outBands chan<- interface{}) {
	/*
		r, err := os.Open("./en.concerts-metal.com_search.html")
		if err != nil {
			log.Fatal(err)
		}
		doc, err := goquery.NewDocumentFromReader(r)
	*/

	// Load the HTML document
	doc, err := goquery.NewDocument(l.cfg.BaseURL + "search.php")
	if err != nil {
		l.breaker.Process("HTTP", err)
		return
	}
	l.breaker.Process("HTTP", nil)

	doc.Find("#groupe").Each(func(i int, s *goquery.Selection) {
		// For each item found, get the band id and title
		if band := s.Find("option"); band != nil {
			band.Each(func(j int, ss *goquery.Selection) {
				id, _ := ss.Attr("value")
				name := ss.Text()
				if id != "" && name != "" {
					outBands <- cmetalBand{
						Id:   id,
						Name: name,
					}
				}
			})
		}
	})
}

// loadBandEvents loads events for band from inBands channel and
// put them into outEvents channel to save into DB.
func (l *CMetalLoader) loadBandEvents(inBands <-chan interface{}, outEvents chan<- interface{}) {
	for e := range inBands {
		band, ok := e.(cmetalBand)
		if !ok {
			l.breaker.Process("APP", fmt.Errorf("Illegal type of argument, expected dao.Band"))
			continue
		}
		doc, err := goquery.NewDocument(l.cfg.BaseURL + "search.php?g=" + band.Id)
		if err != nil {
			l.breaker.Process("HTTP", err)
			continue
		}
		l.breaker.Process("HTTP", nil)

		events := make([]dao.Event, 0)

		/* Next events */
		doc.Find("table tbody").Each(func(i int, s *goquery.Selection) {
			if td := s.Find("td"); td != nil {
				td.Each(func(j int, s1 *goquery.Selection) {
					if strings.HasPrefix(s1.Text(), "Next events (") {
						if nextEvents, err := l.getNextEvents(band, s1); err != nil {
							l.breaker.Process("PARSE", err)
						} else {
							events = append(events, nextEvents...)
						}
					}
				})
			}
		})

		/* Last events */
		doc.Find("table tbody").Each(func(i int, s *goquery.Selection) {
			if td := s.Find("td"); td != nil {
				td.Each(func(j int, s1 *goquery.Selection) {
					if strings.Contains(s1.Text(), "Last events (") {
						if lastEvents, err := l.getLastEvents(band, s1); err != nil {
							l.breaker.Process("PARSE", err)
						} else {
							events = append(events, lastEvents...)
						}
					}
				})
			}
		})
		outEvents <- events
	}
}

// saveBandEvents saves band's events from inEvents channel into DB.
func (l *CMetalLoader) saveBandEvents(inEvents <-chan interface{}, out chan<- interface{}) {
	for e := range inEvents {
		events, ok := e.([]dao.Event)
		if !ok {
			l.breaker.Process("APP", fmt.Errorf("Illegal type of argument, expected []dao.Event"))
			continue
		}
		if len(events) > 0 {
			log.Printf("saveBandEvents: %#v\n", events[0].Band)
		}
	}
}

// getNextEvents returns array of events which will be in the future from html nodes.
func (l *CMetalLoader) getNextEvents(band cmetalBand, s *goquery.Selection) ([]dao.Event, error) {
	clearDetail := func(s string) string {
		if idx := strings.Index(s, " <img"); idx != -1 {
			return s[:idx]
		}
		return s
	}
	if tdt := s.Find("table tbody td"); tdt != nil {
		events := make([]dao.Event, 0)
		tdt.Each(func(k int, s3 *goquery.Selection) {
			if tdHtml, err := s3.Html(); err == nil {
				eventDetail := strings.SplitN(tdHtml, "<br/>", 3)
				if len(eventDetail) > 2 {
					eventDate := eventDetail[1]
					eventLocation := clearDetail(eventDetail[2])
					if eventLink := s3.Find("a").Last(); eventLink != nil {
						eventTitle, _ := eventLink.Attr("title")
						eventHref, _ := eventLink.Attr("href")
						eventImg := ""
						if linkImg := eventLink.Find("img"); linkImg != nil {
							eventImg, _ = linkImg.Attr("src")
						}
						if from, to, err := parseDate(eventDate); err != nil {
							l.breaker.Process("PARSE", err)
						} else {
							events = append(events, dao.Event{
								Band:  band.Name,
								Title: eventTitle,
								From:  from,
								To:    to,
								City:  eventLocation,
								Link:  l.buildURL(eventHref),
								Img:   l.buildURL(eventImg),
							})
						}
					}
				}
			}
		})
		return events, nil
	}
	return nil, nil
}

// getLastEvents returns array of events whichi have been already from html nodes.
func (l *CMetalLoader) getLastEvents(band cmetalBand, s *goquery.Selection) ([]dao.Event, error) {
	if noTable := s.Not("table"); noTable != nil {
		children := noTable.Clone().Children().Remove().End()
		ret, err := children.Html()
		if err != nil {
			return nil, err
		}
		events, err := parseLastEvents(ret)
		if err != nil {
			return nil, err
		}

		k := len(events) - 1
		if k >= 0 {
			tmpEvents := make([]dao.Event, 0)
			noTable.Find("a").Each(func(n int, s_ *goquery.Selection) {
				eventTitle, _ := s_.Attr("title")
				eventHref, _ := s_.Attr("href")
				tmpEvents = append(tmpEvents, dao.Event{
					Title: strings.TrimSpace(eventTitle),
					Link:  l.buildURL(eventHref),
				})
			})

			for i, j := k, len(tmpEvents)-1; i >= 0 && j >= 0; i, j = i-1, j-1 {
				event := tmpEvents[j]
				events[i].Band = band.Name
				events[i].Title = event.Title
				events[i].Link = event.Link
			}
		}
		return events, nil
	}
	return nil, nil
}

// buildURL builds URL based on URL from config and href.
func (l *CMetalLoader) buildURL(href string) string {
	if href != "" {
		return l.cfg.BaseURL + href
	}
	return href
}

type Trigger struct {
	ErrorKind  string
	ErrorLimit int32
	Callback   func(kind string, err error)
}

type state struct {
	Trigger
	errors int32
}

type Breaker struct {
	states map[string]state // key is kind of error
}

func NewBreaker(t []Trigger) *Breaker {
	states := make(map[string]state)
	for _, a := range t {
		states[a.ErrorKind] = state{
			Trigger: a,
		}
	}
	return &Breaker{
		states: states,
	}
}

func (b *Breaker) Process(kind string, err error) {
	state, ok := b.states[kind]
	if !ok {
		log.Printf("Breaker warning: unknown kind of error (%s)\n", kind)
		return
	}
	if err == nil {
		atomic.StoreInt32(&state.errors, 0)
	} else {
		errors := atomic.AddInt32(&state.errors, 1)
		if errors >= state.ErrorLimit {
			state.Callback(kind, err)
		}
	}
}

func (b *Breaker) Fire(kind string, err error) {
	state, ok := b.states[kind]
	if !ok {
		log.Printf("Breaker warning: unknown kind of error (%s)\n", kind)
		return
	}
	state.Callback(kind, err)
}
