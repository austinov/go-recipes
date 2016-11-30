package cmetal

import (
	"errors"
	"fmt"
	"log"
	_ "os"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/austinov/go-recipes/slack-bot/config"
	"github.com/austinov/go-recipes/slack-bot/dao"
	"github.com/austinov/go-recipes/slack-bot/loader"
)

type CMetalLoader struct {
	cfg    config.CMetalConfig
	bands  chan dao.Band
	events chan dao.Band
	done   chan struct{}
}

func New(cfg config.CMetalConfig) loader.Loader {
	return &CMetalLoader{
		cfg:  cfg,
		done: make(chan struct{}),
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
	bands := l.startLoadBands(&wg)

	wg.Add(1)
	bandEvents := l.startLoadBandEvents(&wg, bands)

	wg.Add(1)
	l.startSaveBandEvents(&wg, bandEvents)

	wg.Wait()

	return nil
}

// loadBands loads bands without events and put them into channel
// to load the events these bands.
func (l *CMetalLoader) startLoadBands(wg *sync.WaitGroup) <-chan dao.Band {
	bands := make(chan dao.Band, l.cfg.NumLoaders)
	go func() {
		defer wg.Done()
		defer close(bands)
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
			log.Fatal(err)
		}

		doc.Find("#groupe").Each(func(i int, s *goquery.Selection) {
			// For each item found, get the band id and title
			if band := s.Find("option"); band != nil {
				band.Each(func(j int, ss *goquery.Selection) {
					id, _ := ss.Attr("value")
					name := ss.Text()
					if id != "" && name != "" {
						bands <- dao.Band{
							Id:   id,
							Name: name,
						}
					}
				})
			}
		})
	}()
	return bands
}

func (l *CMetalLoader) startLoadBandEvents(wg *sync.WaitGroup, bands <-chan dao.Band) <-chan dao.Band {
	bandEvents := make(chan dao.Band, l.cfg.NumSavers)

	go func() {
		defer wg.Done()
		defer close(bandEvents)

		var wg_ sync.WaitGroup
		for i := 0; i < l.cfg.NumLoaders; i++ {
			wg_.Add(1)
			go func() {
				defer wg_.Done()
				l.loadBandEvents(bands, bandEvents)
			}()
		}
		wg_.Wait()
	}()

	return bandEvents
}

func (l *CMetalLoader) loadBandEvents(bands <-chan dao.Band, bandEvents chan<- dao.Band) {
	for band := range bands {
		doc, err := goquery.NewDocument(l.cfg.BaseURL + "search.php?g=" + band.Id)
		if err != nil {
			log.Fatal(err)
		}

		events := make([]dao.Event, 0)

		/* Next events */
		doc.Find("table tbody").Each(func(i int, s *goquery.Selection) {
			if td := s.Find("td"); td != nil {
				td.Each(func(j int, s1 *goquery.Selection) {
					if strings.HasPrefix(s1.Text(), "Next events (") {
						if nextEvents, err := l.getNextEvents(s1); err != nil {
							log.Printf("parse next events failed with %#v\n", err)
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
						if lastEvents, err := l.getLastEvents(s1); err != nil {
							log.Printf("parse last events failed with %#v\n", err)
						} else {
							events = append(events, lastEvents...)
						}
					}
				})
			}
		})
		band.Events = events
		bandEvents <- band
	}
}

func (l *CMetalLoader) startSaveBandEvents(wg *sync.WaitGroup, bandEvents <-chan dao.Band) {
	go func() {
		defer wg.Done()

		var wg_ sync.WaitGroup
		for i := 0; i < l.cfg.NumSavers; i++ {
			wg_.Add(1)
			go func() {
				defer wg_.Done()
				for band := range bandEvents {
					log.Printf("saveBandEvents: %#v\n", band.Name)
				}
			}()
		}
		wg_.Wait()
	}()
}

func (l *CMetalLoader) getNextEvents(s *goquery.Selection) ([]dao.Event, error) {
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
							fmt.Printf("%#v\n", err)
						} else {
							events = append(events, dao.Event{
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

func (l *CMetalLoader) getLastEvents(s *goquery.Selection) ([]dao.Event, error) {
	if noTable := s.Not("table"); noTable != nil {
		children := noTable.Clone().Children().Remove().End()
		ret, err := children.Html()
		if err != nil {
			return nil, err
		}
		events := parseLastEvents(ret)

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
				events[i].Title = event.Title
				events[i].Link = event.Link
			}
		}
		return events, nil
	}
	return nil, nil
}

func (l *CMetalLoader) buildURL(href string) string {
	if href != "" {
		return l.cfg.BaseURL + href
	}
	return href
}

// TODO test
func parseDate(date string) (int64, int64, error) {
	//t1 := "30/05/2012"
	if t, err := time.Parse("02/01/2006", date); err == nil {
		from := t.Unix()
		return from, from, nil
	}

	//t2 := "Tuesday 29 November 2016"
	if t, err := time.Parse("Monday 2 January 2006", date); err == nil {
		from := t.Unix()
		return from, from, nil
	}

	//t3 := "From 31 March to 11 April 2017"
	prefix := "From "
	if strings.HasPrefix(date, prefix) {
		parts := strings.Split(date, " to ")
		var to time.Time
		if len(parts) > 1 {
			var err error
			to, err = time.Parse("2 January 2006", strings.TrimSpace(parts[1]))
			if err != nil {
				return 0, 0, err
			}
		}
		partFrom := strings.TrimSpace(parts[0][len(prefix):])
		from, err := time.Parse("2 January 2006", partFrom)
		if err != nil {
			partFrom += fmt.Sprintf(" %d", to.Year())
			from, err = time.Parse("2 January 2006", partFrom)
		}
		return from.Unix(), to.Unix(), nil
	}
	return 0, 0, errors.New("cannot parse [" + date + "]")
}

// TODO test
func parseLastEvents(text string) []dao.Event {
	re := regexp.MustCompile("\\d{1,2}/\\d{1,2}/\\d{4}")
	idxs := re.FindAllStringIndex(text, -1)
	l := len(idxs)
	result := make([]dao.Event, l)
	for i := 0; i < l; i++ {
		idx := idxs[i]
		date := strings.TrimSpace(text[idx[0]:idx[1]])
		tail := ""
		if i < l-1 {
			tail = text[idx[1]:idxs[i+1][0]]
		} else {
			tail = text[idx[1]:]
		}
		details := strings.Split(tail, ",")
		city := strings.TrimSpace(details[0])
		venue := ""
		if len(details) > 1 {
			venue = strings.TrimSpace(details[1])
		}
		if from, to, err := parseDate(date); err != nil {
			fmt.Printf("%#v\n", err)
		} else {
			result[i] = dao.Event{
				From:  from,
				To:    to,
				City:  city,
				Venue: venue,
			}
		}
	}
	return result
}
