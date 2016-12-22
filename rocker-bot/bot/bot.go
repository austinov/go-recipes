package bot

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/austinov/go-recipes/backoff"
	"github.com/austinov/go-recipes/rocker-bot/common"
	"github.com/austinov/go-recipes/rocker-bot/config"
	"github.com/austinov/go-recipes/rocker-bot/store"

	"golang.org/x/net/websocket"
)

const (
	apiURL               = "https://api.slack.com/"
	startRtmURL          = "https://slack.com/api/rtm.start?token=%s"
	attemptsToReceiveMsg = 3
)

type Bot struct {
	cfg config.BotConfig
	dao store.Dao
	ws  *websocket.Conn
	id  string
}

func New(cfg config.BotConfig, dao store.Dao) *Bot {
	if cfg.NumHandlers <= 0 {
		cfg.NumHandlers = 1
	}
	if cfg.NumSenders <= 0 {
		cfg.NumSenders = 1
	}
	return &Bot{
		cfg: cfg,
		dao: dao,
	}
}

func (b *Bot) Start() {
	log.Println("Start bot.")
	if err := b.connect(); err != nil {
		log.Fatal(err)
	}

	var wg sync.WaitGroup

	wg.Add(1)
	messages := make(chan interface{}, b.cfg.NumHandlers)
	common.RunWorkers(&wg, nil, messages, 1, b.pollMessages)
	wg.Add(1)
	replies := make(chan interface{}, b.cfg.NumSenders)
	common.RunWorkers(&wg, messages, replies, b.cfg.NumHandlers, b.processMessages)
	wg.Add(1)
	common.RunWorkers(&wg, replies, nil, b.cfg.NumSenders, b.processReplies)

	wg.Wait()
}

func (b *Bot) connect() error {
	resp, err := http.Get(fmt.Sprintf(startRtmURL, b.cfg.Token))
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return fmt.Errorf("Start RTM request failed with code %d", resp.StatusCode)
	}
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	var respRtm ResponseRtmStart
	if err = json.Unmarshal(body, &respRtm); err != nil {
		return err
	}

	if !respRtm.Ok {
		return fmt.Errorf("Slack error: %s", respRtm.Error)
	}

	ws, err := websocket.Dial(respRtm.Url, "", apiURL)
	if err != nil {
		return err
	}

	b.ws = ws
	b.id = "<@" + respRtm.Self.Id + ">"

	return nil
}

func (b *Bot) pollMessages(ignore <-chan interface{}, outMessages chan<- interface{}) {
	eb := backoff.NewExpBackoff()
	for {
		var m Message
		if err := websocket.JSON.Receive(b.ws, &m); err != nil {
			fmt.Fprintf(os.Stderr, "receive message from socket failed with %#v\n", err)
			if eb.Attempts() >= uint64(attemptsToReceiveMsg) {
				os.Exit(1)
			}
			<-eb.Delay()
		} else {
			eb.Reset()
			outMessages <- m
		}
	}
}

func (b *Bot) processMessages(inMessages <-chan interface{}, outReplies chan<- interface{}) {
	for e := range inMessages {
		message, ok := e.(Message)
		if !ok {
			log.Fatalln("Illegal type of argument, expected Message")
		}
		go b.processMessage(message, outReplies)
	}
}

func (b *Bot) processMessage(msg Message, outReplies chan<- interface{}) {
	if msg.Type == "message" && strings.HasPrefix(msg.Text, b.id) {
		query := Parse(msg.Text)
		if query.IsValid() && query.Command == "events" {
			msg.Text = b.calendarHandler(query)
		} else {
			msg.Text = b.helpHandler()
		}
		outReplies <- msg
	}
}

// sequentially increased message counter
var messageId uint64

func (b *Bot) processReplies(inReplies <-chan interface{}, ignore chan<- interface{}) {
	for e := range inReplies {
		reply, ok := e.(Message)
		if !ok {
			log.Fatalln("Illegal type of argument, expected Message")
		}
		reply.Id = atomic.AddUint64(&messageId, 1)
		if err := websocket.JSON.Send(b.ws, reply); err != nil {
			fmt.Fprintf(os.Stderr, "send reply failed with %#v\n", err)
		}
	}
}

// helpHandler returns a reply containing help text.
func (b *Bot) helpHandler() string {
	buffer := bytes.NewBufferString("Please, use commands like the follow:\n")
	buffer.WriteString(fmt.Sprintf(">%s events of Metallica - list events of Metallica\n", b.id))
	buffer.WriteString(fmt.Sprintf(">%s events in Paris     - list events in Paris\n", b.id))
	buffer.WriteString(fmt.Sprintf(">%s events in London at 27 May 2017 - list events in city at the date (date format may be also dd.MM.yyyy or dd/MM/yyyy)\n", b.id))
	buffer.WriteString(fmt.Sprintf(">%s events of System of a Down in Dresden since 1 Jan 2017 - list events of band in city since the date\n", b.id))
	buffer.WriteString(fmt.Sprintf(">%s events in Helsinki till 1 Jan 2017 - list events in city till the date\n", b.id))
	buffer.WriteString(fmt.Sprintf(">%s events in St Petersburg since 15 Dec 2016 till 1 Jan 2017 - list events in city since/till dates\n", b.id))
	buffer.WriteString(fmt.Sprintf(">%s events of Aerosmith for 15 Dec 2016 and 13 Jan 2017 - list events of band for period\n", b.id))
	return buffer.String()
}

// calendarHandler returns calendar for the band.
func (b *Bot) calendarHandler(query Query) string {
	if query.To == 0 {
		query.To = time.Now().AddDate(10, 0, 0).Unix()
	}
	offset, limit := 0, 42
	events, err := b.dao.GetEvents(query.Band, query.City, query.From, query.To, offset, limit)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return "Sorry, we have some troubles"
	} else {
		l := len(events)
		if l == 0 {
			return formatHeader(query, true)
		} else {
			out := formatHeader(query, false)
			for _, event := range events {
				out += formatEvent(event)
			}
			if l >= limit {
				out += formatFooter(b.id, query, events[l-1])
			}
			return out
		}
	}
}

func formatHeader(q Query, empty bool) string {
	var band, city string
	if q.Band != "" {
		band = fmt.Sprintf(" of *%s*", q.Band)
	}
	if q.City != "" {
		city = fmt.Sprintf(" in _%s_", q.City)
	}
	if empty {
		return fmt.Sprintf("We have no more info about events%s%s.", band, city)
	} else {
		return fmt.Sprintf("We known about the following events%s%s:\n", band, city)
	}
}

func formatEvent(e store.Event) string {
	fd := func(sec int64) string {
		return time.Unix(sec, 0).Format("2 Jan 2006")
	}
	var dates, location, link string
	if e.From != e.To {
		dates = fmt.Sprintf("%s - %s", fd(e.From), fd(e.To))
	} else {
		dates = fmt.Sprintf("%s", fd(e.From))
	}
	if e.City != "" && e.Venue != "" {
		location = fmt.Sprintf("(%s - _%s_)", e.City, e.Venue)
	} else if e.City != "" {
		location = fmt.Sprintf("(%s)", e.City)
	} else if e.Venue != "" {
		location = fmt.Sprintf("(_%s_)", e.Venue)
	}
	if e.Link != "" {
		link = fmt.Sprintf("- %s", e.Link)
	}
	return fmt.Sprintf(">%s, *%s* %s %s\n", dates, e.Title, location, link)
}

func formatFooter(id string, q Query, e store.Event) string {
	var band, city string
	if q.Band != "" {
		band = fmt.Sprintf(" of %s", q.Band)
	}
	if q.City != "" {
		city = fmt.Sprintf(" in %s", q.City)
	}
	since := time.Unix(e.From, 0).Format("02 Jan 2006")
	return fmt.Sprintf("To load next portion of events you may use:\n>%s events%s%s since %s", id, band, city, since)
}
