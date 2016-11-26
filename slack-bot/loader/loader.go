package loader

import (
	"errors"
	"fmt"
	"log"
	_ "os"
	"regexp"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/austinov/go-recipes/slack-bot/model"
)

const (
	baseURL = "http://en.concerts-metal.com/"
)

var (
	searchURL = baseURL + "search.php"
)

func Start() {
	/*r, err := os.Open("./en.concerts-metal.com_search.html")
	if err != nil {
		log.Fatal(err)
	}*/
	// Load the HTML document
	doc, err := goquery.NewDocument(searchURL)
	//doc, err := goquery.NewDocumentFromReader(r)
	if err != nil {
		log.Fatal(err)
	}

	/* Bands */
	bands := make([]model.Band, 0)
	doc.Find("#groupe").Each(func(i int, s *goquery.Selection) {
		// For each item found, get the band id and title
		if band := s.Find("option"); band != nil {
			band.Each(func(j int, ss *goquery.Selection) {
				id, _ := ss.Attr("value")
				name := ss.Text()
				if id != "" && name != "" {
					bands = append(bands, model.Band{
						Id:   id,
						Name: name,
					})
				}
			})
		}
	})
	fmt.Printf("%#v\n", bands)
	/**/

	return

	/* Next events */
	doc.Find("table tbody").Each(func(i int, s1 *goquery.Selection) {
		if td := s1.Find("td"); td != nil {
			td.Each(func(j int, s2 *goquery.Selection) {
				if strings.HasPrefix(s2.Text(), "Next events (") {
					nextEvents, err := getNextEvents(s2)
					fmt.Printf("%#v\nerror: %#v\n", nextEvents, err)
				}
			})
		}
	})

	fmt.Println()

	/* Last events */
	doc.Find("table tbody").Each(func(i int, s1 *goquery.Selection) {
		if td := s1.Find("td"); td != nil {
			td.Each(func(j int, s2 *goquery.Selection) {
				if strings.Contains(s2.Text(), "Last events (") {
					lastEvents, err := getLastEvents(s2)
					fmt.Printf("%#v\nerror: %#v\n", lastEvents, err)
				}
			})
		}
	})
}

func getNextEvents(s *goquery.Selection) ([]model.Event, error) {
	clearDetail := func(s string) string {
		if idx := strings.Index(s, " <img"); idx != -1 {
			return s[:idx]
		}
		return s
	}
	if tdt := s.Find("table tbody td"); tdt != nil {
		events := make([]model.Event, 0)
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
							events = append(events, model.Event{
								Title: eventTitle,
								From:  from,
								To:    to,
								City:  eventLocation,
								Link:  buildURL(eventHref),
								Img:   buildURL(eventImg),
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

func getLastEvents(s *goquery.Selection) ([]model.Event, error) {
	if noTable := s.Not("table"); noTable != nil {
		children := noTable.Clone().Children().Remove().End()
		ret, err := children.Html()
		if err != nil {
			return nil, err
		}
		events := parseLastEvents(ret)

		k := len(events) - 1
		if k >= 0 {
			tmpEvents := make([]model.Event, 0)
			noTable.Find("a").Each(func(n int, s_ *goquery.Selection) {
				eventTitle, _ := s_.Attr("title")
				eventHref, _ := s_.Attr("href")
				tmpEvents = append(tmpEvents, model.Event{
					Title: strings.TrimSpace(eventTitle),
					Link:  buildURL(eventHref),
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

func buildURL(href string) string {
	if href != "" {
		return baseURL + href
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
func parseLastEvents(text string) []model.Event {
	re := regexp.MustCompile("\\d{1,2}/\\d{1,2}/\\d{4}")
	idxs := re.FindAllStringIndex(text, -1)
	l := len(idxs)
	result := make([]model.Event, l)
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
			result[i] = model.Event{
				From:  from,
				To:    to,
				City:  city,
				Venue: venue,
			}
		}
	}
	return result
}
