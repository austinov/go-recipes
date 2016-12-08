package bot

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/austinov/go-recipes/slack-bot/common"
	"github.com/austinov/go-recipes/slack-bot/config"
	"github.com/austinov/go-recipes/slack-bot/store"

	"golang.org/x/net/websocket"
)

const (
	apiURL      = "https://api.slack.com/"
	startRtmURL = "https://slack.com/api/rtm.start?token=%s"
)

type Bot struct {
	cfg      config.BotConfig
	dao      store.Dao
	ws       *websocket.Conn
	id       string
	messages chan Message
	replies  chan Message
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

	b.messages = make(chan Message, 1)
	b.replies = make(chan Message, 1)

	var wg sync.WaitGroup

	wg.Add(1)
	go b.pollMessages(&wg)
	for i := 0; i < b.cfg.NumHandlers; i++ {
		wg.Add(1)
		go b.processMessages(&wg)
	}
	for i := 0; i < b.cfg.NumSenders; i++ {
		wg.Add(1)
		go b.processReplies(&wg)
	}
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

func (b *Bot) pollMessages(wg *sync.WaitGroup) {
	defer wg.Done()
	for {
		b.poll()
	}
}

func (b *Bot) processReplies(wg *sync.WaitGroup) {
	defer wg.Done()
	for reply := range b.replies {
		b.sendReply(reply)
	}
}

func (b *Bot) poll() {
	var m Message
	if err := websocket.JSON.Receive(b.ws, &m); err != nil {
		log.Fatal(err) // TODO backoff
	}
	b.messages <- m
}

func (b *Bot) processMessages(wg *sync.WaitGroup) {
	defer wg.Done()
	for message := range b.messages {
		go b.processMessage(message)
	}
}

func (b *Bot) processMessage(msg Message) {
	log.Printf("Process message: %#v\n", msg)

	if msg.Type == "message" && strings.HasPrefix(msg.Text, b.id) {
		// calendar <band> [next|last]
		fields := strings.Fields(msg.Text)
		l := len(fields)
		if l > 2 && fields[1] == "calendar" {
			band := fields[2]
			mode := "all"
			if l > 3 {
				mode = fields[3]
			}
			msg.Text = b.calendarHandler(band, mode)
		} else {
			msg.Text = b.helpHandler()
		}
		b.replies <- msg
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
	events, err := b.dao.GetCalendar(band, from, to)
	if err != nil {
		log.Println(err)
		return "Sorry, we have some troubles"
	} else {
		return fmt.Sprintf("TODO %s (%v - %v) events for %s: %v\n", mode, time.Unix(from, 0), time.Unix(to, 0), band, events)
	}
}

// sequentially increased message counter
var messageId uint64

func (b *Bot) sendReply(m Message) {
	m.Id = atomic.AddUint64(&messageId, 1)
	if err := websocket.JSON.Send(b.ws, m); err != nil {
		log.Println("send reply failed with", err)
	}
}
