package bot

import (
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
	"github.com/austinov/go-recipes/slack-bot/common"
	"github.com/austinov/go-recipes/slack-bot/config"
	"github.com/austinov/go-recipes/slack-bot/store"

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
		// calendar <band> [next|last]
		fields := strings.Fields(msg.Text)
		l := len(fields)
		if l > 2 && fields[1] == "calendar" {
			// TODO process band/city with several words
			band := fields[2]
			mode := "all"
			if l > 3 {
				mode = fields[3]
			}
			msg.Text = b.calendarHandler(band, mode)
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
	return "Please, use the follow commands:\n" +
		"calendar <band> [next|past] - show calendar for the band (next or past events, all events when no mode)\n"
}

// calendarHandler returns calendar for the band.
func (b *Bot) calendarHandler(band, mode string) string {
	// TODO count, offset
	now := time.Now()
	var from, to int64
	switch mode {
	case "all":
		// all events
		to = now.AddDate(10, 0, 0).Unix()
	case "past":
		// only past events to the present day
		to = common.EndOfDate(now).Unix()
	case "next":
		// only future events
		from = common.BeginOfDate(now).Unix()
		to = now.AddDate(10, 0, 0).Unix()
	default:
		return b.helpHandler()
	}
	events, err := b.dao.GetBandEvents(band, from, to, 0, 10)
	//events, err := b.dao.GetCityEvents(city, from, to, 0, 10)
	//events, err := b.dao.GetBandInCityEvents(band, city, from, to, 0, 10)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return "Sorry, we have some troubles"
	} else {
		// TODO
		if len(events) == 0 {
			return fmt.Sprintf("Sorry, we have not info about *%s*'s events.", band)
		} else {
			out := fmt.Sprintf("We known about the following events of *%s*:\n", band)
			for _, event := range events {
				out = out + formatEvent(event)
			}
			return out
		}
	}
}

func formatEvent(e store.Event) string {
	fd := func(sec int64) string {
		return time.Unix(sec, 0).Format("02 Jar 2006")
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
	return fmt.Sprintf("> %s, *%s* %s %s\n", dates, e.Title, location, link)
}
