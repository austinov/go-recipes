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

	"golang.org/x/net/websocket"
)

const (
	apiURL      = "https://api.slack.com/"
	startRtmURL = "https://slack.com/api/rtm.start?token=%s"
)

type Bot struct {
	token   string
	id      string
	ws      *websocket.Conn
	updates chan Message
	replies chan Message

	//mu   sync.Mutex
	done chan struct{}
}

func New(token string) *Bot {
	if token == "" {
		log.Fatal("Token is empty")
	}
	return &Bot{
		token: token,
	}
}

func (b *Bot) Start() {
	log.Println("Start bot.")
	// TODO
	if err := b.connect(); err != nil {
		log.Fatal(err)
	}

	b.updates = make(chan Message, 1)
	b.replies = make(chan Message, 1)
	//b.mu.Lock()
	b.done = make(chan struct{})
	//b.mu.Unlock()

	var wg sync.WaitGroup

	wg.Add(1)
	go b.pollMessages(&wg)
	for i := 0; i < 1; /*numPollers*/ i++ {
		wg.Add(1)
		go b.processMessages(&wg)
	}
	for i := 0; i < 1; /*numSenders*/ i++ {
		wg.Add(1)
		go b.processReplies(&wg)
	}
	wg.Wait()
}

func (b *Bot) Stop() {
	// TODO
	//b.ws.Close()
}

func (b *Bot) connect() error {
	resp, err := http.Get(fmt.Sprintf(startRtmURL, b.token))
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
		select {
		case <-b.done:
			return
		default:
			b.poll()
		}
	}
}

func (b *Bot) processReplies(wg *sync.WaitGroup) {
	defer wg.Done()
	for {
		select {
		case <-b.done:
			return
		case r, ok := <-b.replies:
			if !ok {
				return
			}
			b.sendReply(r)
		}
	}
}

func (b *Bot) poll() {
	log.Printf("Try to get updates...\n")
	var m Message
	if err := websocket.JSON.Receive(b.ws, &m); err != nil {
		log.Fatal(err) // TODO
	}
	b.updates <- m
}

/*
func (b *Bot) getMessage() (Message, error) {
	var m Message
	err := websocket.JSON.Receive(b.ws, &m)
	return m, err
}
*/

func (b *Bot) processMessages(wg *sync.WaitGroup) {
	defer wg.Done()
	for {
		select {
		case <-b.done:
			return
		case message, ok := <-b.updates:
			if !ok {
				return
			}
			go b.processMessage(message)
		}
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
		"calendar <band> [next|last] - show calendar for the band (next or last events, all events when no mode)\n"
}

// calendarHandler returns calendar for the band.
func (b *Bot) calendarHandler(band, mode string) string {
	// TODO
	return "TODO " + mode + " events for " + band + "\n"
}

func (b *Bot) sendReply(m Message) {
	m.Id = atomic.AddUint64(&counter, 1)
	if err := websocket.JSON.Send(b.ws, m); err != nil {
		log.Println("send reply failed with", err)
	}
}

var counter uint64
