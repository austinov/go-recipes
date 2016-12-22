package cmetal

import (
	"errors"
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/austinov/go-recipes/rocker-bot/store"
)

var re = regexp.MustCompile("\\d{1,2}/\\d{1,2}/\\d{4}")

// parseDate parses date string from concerts-metal.com.
// It returns begin and end Unix dates
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

// parseLastEvents parses html with events from concerts-metal.com.
// It returns array of events.
func parseLastEvents(text string) ([]store.Event, error) {
	idxs := re.FindAllStringIndex(text, -1)
	l := len(idxs)
	result := make([]store.Event, l)
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
		if strings.HasPrefix(city, "@ ") {
			city = strings.Trim(city, "@ ")
		}
		venue := ""
		if len(details) > 1 {
			venue = strings.TrimSpace(details[1])
		}
		if from, to, err := parseDate(date); err != nil {
			return nil, err
		} else {
			result[i] = store.Event{
				From:  from,
				To:    to,
				City:  toUtf8(city),
				Venue: toUtf8(venue),
			}
		}
	}
	return result, nil
}

func toUtf8(text string) string {
	bbuf := []byte(text)
	rbuf := make([]rune, len(bbuf))
	for i, b := range bbuf {
		rbuf[i] = rune(b)
	}
	return string(rbuf)
}
