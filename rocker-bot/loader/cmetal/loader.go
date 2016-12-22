package cmetal

import (
	"fmt"
	"log"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/austinov/go-recipes/rocker-bot/common"
	"github.com/austinov/go-recipes/rocker-bot/config"
	"github.com/austinov/go-recipes/rocker-bot/loader"
	"github.com/austinov/go-recipes/rocker-bot/store"
)

type cmetalBand struct {
	Id   string
	Name string
}

type CMetalLoader struct {
	cfg        config.CMetalConfig
	dao        store.Dao
	httpclient *common.HTTPClient
	fuse       *common.Fuse
	bands      chan cmetalBand
	events     chan store.Event
	done       chan struct{}
}

func New(cfg config.CMetalConfig, dao store.Dao) loader.Loader {
	loader := &CMetalLoader{
		cfg:        cfg,
		dao:        dao,
		done:       make(chan struct{}),
		httpclient: common.NewHTTPClient(30 * time.Second),
	}
	fuseTriggers := make([]common.FuseTrigger, 0)
	fuseTriggers = append(fuseTriggers,
		common.NewFuseTrigger("APP", 1, func(kind string, err error) {
			loader.fuseHandler(kind, err)
		}))
	fuseTriggers = append(fuseTriggers,
		common.NewFuseTrigger("HTTP", 10, func(kind string, err error) {
			loader.fuseHandler(kind, err)
		}))
	fuseTriggers = append(fuseTriggers,
		common.NewFuseTrigger("PARSE", 1, func(kind string, err error) {
			loader.fuseHandler(kind, err)
		}))
	loader.fuse = common.NewFuse(fuseTriggers)
	return loader
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
	log.Println("Loader stopping")
	close(l.done)
}

func (l *CMetalLoader) fuseHandler(kind string, err error) {
	fmt.Fprintf(os.Stderr, "loader failed due %s error: %#v\n", kind, err)
	l.Stop()
}

func (l *CMetalLoader) do() error {
	var wg sync.WaitGroup

	wg.Add(1)
	bands := make(chan interface{}, l.cfg.NumLoaders)
	common.RunWorkers(&wg, nil, bands, 1, l.loadBands)

	wg.Add(1)
	bandEvents := make(chan interface{}, l.cfg.NumSavers)
	common.RunWorkers(&wg, bands, bandEvents, l.cfg.NumLoaders, l.loadBandEvents)

	wg.Add(1)
	common.RunWorkers(&wg, bandEvents, nil, l.cfg.NumSavers, l.saveBandEvents)

	wg.Wait()

	return nil
}

// loadBands loads bands without events and put them into outBands channel
// to load the events these bands.
func (l *CMetalLoader) loadBands(ignore <-chan interface{}, outBands chan<- interface{}) {
	doc := l.loadHTMLDocument(l.cfg.BaseURL + "search.php")
	if doc == nil {
		return
	}
	doc.Find("#groupe").Each(func(i int, s *goquery.Selection) {
		// For each item found, get the band id and title
		if band := s.Find("option"); band != nil {
			band.Each(func(j int, ss *goquery.Selection) {
				id, _ := ss.Attr("value")
				name := ss.Text()
				if id != "" && name != "" {
					select {
					case <-l.done:
						return
					default:
						if name != "band" { // reserved word
							outBands <- cmetalBand{
								Id:   id,
								Name: name,
							}
						}
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
			l.fuse.Process("APP", fmt.Errorf("Illegal type of argument, expected dao.Band"))
			continue
		}
		doc := l.loadHTMLDocument(l.cfg.BaseURL + "search.php?g=" + band.Id)
		if doc == nil {
			continue
		}

		events := make([]store.Event, 0)

		/* Next events */
		if table := doc.Find("table tbody"); table != nil {
			table.Each(func(i int, s *goquery.Selection) {
				if td := s.Find("td"); td != nil {
					td.Each(func(j int, s1 *goquery.Selection) {
						if strings.HasPrefix(s1.Text(), "Next events (") {
							if nextEvents, err := l.getNextEvents(band, s1); err != nil {
								l.fuse.Process("PARSE", fmt.Errorf("parse next event for %#v failed with %#v", band, err))
							} else {
								events = append(events, nextEvents...)
								l.fuse.Process("PARSE", nil)
							}
						}
					})
				}
			})
		}

		/* Last events */
		doc.Find("table tbody").Each(func(i int, s *goquery.Selection) {
			if td := s.Find("td"); td != nil {
				td.Each(func(j int, s1 *goquery.Selection) {
					if strings.Contains(s1.Text(), "Last events (") {
						if lastEvents, err := l.getLastEvents(band, s1); err != nil {
							l.fuse.Process("PARSE", fmt.Errorf("parse last event for %#v failed with %#v", band, err))
						} else {
							events = append(events, lastEvents...)
							l.fuse.Process("PARSE", nil)
						}
					}
				})
			}
		})
		outEvents <- events
	}
}

// saveBandEvents saves band's events from inEvents channel into DB.
func (l *CMetalLoader) saveBandEvents(inEvents <-chan interface{}, ignore chan<- interface{}) {
	for e := range inEvents {
		events, ok := e.([]store.Event)
		if !ok {
			l.fuse.Process("APP", fmt.Errorf("Illegal type of argument, expected []dao.Event"))
			continue
		}
		if len(events) > 0 {
			if err := l.dao.AddBandEvents(events); err != nil {
				fmt.Fprintf(os.Stderr, "save band's (%s) events failed with %#v\n", events[0].Band, err)
			}
			log.Printf("saveBandEvents: %#v\n", events[0].Band)
		}
	}
}

// getNextEvents returns array of events which will be in the future from html nodes.
func (l *CMetalLoader) getNextEvents(band cmetalBand, s *goquery.Selection) ([]store.Event, error) {
	clearCity := func(s string) string {
		if idx := strings.Index(s, " <img"); idx != -1 {
			s = s[:idx]
		}
		return strings.Split(s, " - ")[0]
	}
	if tdt := s.Find("table tbody td"); tdt != nil {
		events := make([]store.Event, 0)
		tdt.Each(func(k int, s3 *goquery.Selection) {
			if tdHtml, err := s3.Html(); err == nil {
				eventDetail := strings.SplitN(tdHtml, "<br/>", 3)
				if len(eventDetail) > 2 {
					if eventLink := s3.Find("a").Last(); eventLink != nil {
						eventTitle, _ := eventLink.Attr("title")
						eventHref, _ := eventLink.Attr("href")
						eventHref = l.buildURL(eventHref)
						eventImg := ""
						if linkImg := eventLink.Find("img"); linkImg != nil {
							eventImg, _ = linkImg.Attr("src")
						}
						eventDate := eventDetail[1]
						eventCity := clearCity(eventDetail[2])
						from, to, err := parseDate(eventDate)
						if err != nil {
							// when event has no image city locates in date place in html
							// but date locates right after <h5>
							if dates := strings.Split(eventDetail[0], "</a></h5>"); len(dates) == 2 {
								eventCity = clearCity(eventDate)
								from, to, err = parseDate(dates[1])
							}
						}
						if err != nil {
							l.fuse.Process("PARSE", fmt.Errorf("parse date next event for %#v failed with %#v", band, err))
						}
						events = append(events, store.Event{
							Band:  toUtf8(band.Name),
							Title: toUtf8(eventTitle),
							From:  from,
							To:    to,
							City:  toUtf8(eventCity),
							Link:  eventHref,
							Img:   l.buildURL(eventImg),
							Venue: l.getNextEventVenue(eventHref),
						})
						l.fuse.Process("PARSE", nil)
					}
				}
			}
		})
		return events, nil
	}
	return nil, nil
}

func (l *CMetalLoader) getNextEventVenue(url string) string {
	doc := l.loadHTMLDocument(url)
	if doc == nil {
		return ""
	}
	if div := doc.Find("div[itemprop='address']").First(); div != nil {
		if td := div.Find("td"); td != nil {
			if ftd := td.First(); ftd != nil && len(ftd.Nodes) > 0 {
				if venueNode := ftd.Nodes[0].FirstChild; venueNode != nil {
					return venueNode.Data
				}
			}
		}
	}
	return ""
}

// getLastEvents returns array of events whichi have been already from html nodes.
func (l *CMetalLoader) getLastEvents(band cmetalBand, s *goquery.Selection) ([]store.Event, error) {
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
			tmpEvents := make([]store.Event, 0)
			noTable.Find("a").Each(func(n int, s_ *goquery.Selection) {
				eventTitle, _ := s_.Attr("title")
				eventHref, _ := s_.Attr("href")
				tmpEvents = append(tmpEvents, store.Event{
					Title: strings.TrimSpace(eventTitle),
					Link:  l.buildURL(eventHref),
				})
			})

			for i, j := k, len(tmpEvents)-1; i >= 0 && j >= 0; i, j = i-1, j-1 {
				event := tmpEvents[j]
				events[i].Band = toUtf8(band.Name)
				events[i].Title = toUtf8(event.Title)
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
	return ""
}

func (l *CMetalLoader) loadHTMLDocument(url string) *goquery.Document {
	resp, err := l.httpclient.Get(url)
	if err != nil {
		l.fuse.Process("HTTP", err)
		return nil
	}
	doc, err := goquery.NewDocumentFromResponse(resp)
	if err != nil {
		l.fuse.Process("HTTP", err)
		return nil

	}
	l.fuse.Process("HTTP", nil)
	return doc
}
