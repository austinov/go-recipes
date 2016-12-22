package bot

import (
	"container/list"
	"strings"
	"time"

	"github.com/austinov/go-recipes/rocker-bot/common"
)

type Query struct {
	Command string
	Band    string
	City    string
	From    int64
	To      int64
}

func (q Query) IsValid() bool {
	return q.Command != "" && (q.Band != "" || q.City != "")
}

const (
	band byte = iota
	city
	at
	since
	till
	between
)

var paramKinds = map[byte]string{
	band:    " of ",
	city:    " in ",
	at:      " at ",
	since:   " since ",
	till:    " till ",
	between: " for ",
}

type param struct {
	kind  byte
	pos   int
	value string
}

func Parse(text string) Query {
	m := strings.ToLower(text)

	l := list.New()
	addToList(l, band, strings.Index(m, paramKinds[band]))
	addToList(l, city, strings.Index(m, paramKinds[city]))
	addToList(l, at, strings.Index(m, paramKinds[at]))
	addToList(l, since, strings.Index(m, paramKinds[since]))
	addToList(l, till, strings.Index(m, paramKinds[till]))
	addToList(l, between, strings.Index(m, paramKinds[between]))

	query := Query{}
	fields := strings.Fields(text)
	if len(fields) > 1 {
		query.Command = fields[1]
	}
	for e := l.Front(); e != nil; e = e.Next() {
		p := e.Value.(param)
		begin, end := p.pos, len(m)
		if e.Next() != nil {
			end = e.Next().Value.(param).pos
		}
		p.value = text[begin+len(paramKinds[p.kind]) : end]
		fillQuery(&query, p)
	}
	return query
}

func fillQuery(q *Query, p param) {
	switch p.kind {
	case band:
		q.Band = p.value
	case city:
		q.City = p.value
	case at:
		q.From, q.To = parseAtDate(p.value)
	case since:
		q.From = parseSinceDate(p.value)
	case till:
		q.To = parseTillDate(p.value)
	case between:
		q.From, q.To = parseBetweenDates(p.value)
	}
}

func parseDate(d string) *time.Time {
	layouts := []string{
		"2.01.2006",
		"02.01.2006",
		"2/01/2006",
		"02/01/2006",
		"2 Jan 2006",
		"02 Jan 2006",
		"2 January 2006",
		"02 January 2006",
	}
	for _, l := range layouts {
		if t, err := time.Parse(l, d); err == nil {
			return &t
		}
	}
	return nil
}

func parseAtDate(d string) (int64, int64) {
	if t := parseDate(d); t != nil {
		return common.BeginOfDate(*t).Unix(), common.EndOfDate(*t).Unix()
	}
	return 0, 0
}

func parseSinceDate(d string) int64 {
	if t := parseDate(d); t != nil {
		return common.BeginOfDate(*t).Unix()
	}
	return 0
}

func parseTillDate(d string) int64 {
	if t := parseDate(d); t != nil {
		return common.EndOfDate(*t).Unix()
	}
	return 0
}

func parseBetweenDates(d string) (int64, int64) {
	from, to := "", ""
	if parts := strings.Split(d, "-"); len(parts) == 2 {
		from, to = strings.TrimSpace(parts[0]), strings.TrimSpace(parts[1])
	} else if parts := strings.Split(d, "and"); len(parts) == 2 {
		from, to = strings.TrimSpace(parts[0]), strings.TrimSpace(parts[1])
	}
	if from != "" && to != "" {
		if tf, tt := parseDate(from), parseDate(to); tf != nil && tt != nil {
			return common.BeginOfDate(*tf).Unix(), common.EndOfDate(*tt).Unix()
		}
	}
	return 0, 0
}

func addToList(l *list.List, kind byte, pos int) {
	if pos == -1 {
		return
	}
	inserted := false
	for e := l.Front(); e != nil; e = e.Next() {
		if p := e.Value.(param); p.pos > pos {
			l.InsertBefore(param{kind, pos, ""}, e)
			inserted = true
			break
		}
	}
	if !inserted {
		l.PushBack(param{kind, pos, ""})
	}
}
