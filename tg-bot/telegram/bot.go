package telegram

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"sync"
	"sync/atomic"
	"time"
)

const (
	apiURL      = "https://api.telegram.org/bot%s/%s"
	pollTimeout = "30" // in seconds
	pollDelay   = 1 * time.Second
	numPollers  = 2
	numSenders  = 3
)

type Bot struct {
	token   string
	offset  uint64
	updates chan []Update
	replies chan Message

	mu   sync.Mutex
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
	// reset offset to get all lost messages
	atomic.StoreUint64(&b.offset, 0)

	b.updates = make(chan []Update, numPollers)
	b.replies = make(chan Message, numSenders)
	b.mu.Lock()
	b.done = make(chan struct{})
	b.mu.Unlock()

	var wg sync.WaitGroup

	wg.Add(1)
	go b.startPoll(&wg)
	for i := 0; i < numPollers; i++ {
		wg.Add(1)
		go b.startProcess(&wg)
	}
	for i := 0; i < numSenders; i++ {
		wg.Add(1)
		go b.startReply(&wg)
	}
	wg.Wait()
}

func (b *Bot) Stop() {
	b.mu.Lock()
	close(b.done)
	b.mu.Unlock()
}

func (b *Bot) startPoll(wg *sync.WaitGroup) {
	defer wg.Done()
	for {
		select {
		case <-b.done:
			return
		default:
			b.pollUpdates()
		}
	}
}

func (b *Bot) startProcess(wg *sync.WaitGroup) {
	defer wg.Done()
	for {
		select {
		case <-b.done:
			return
		case updates, ok := <-b.updates:
			if !ok {
				return
			}
			for _, update := range updates {
				go b.processMessage(update.Message)
			}
		}
	}
}

func (b *Bot) startReply(wg *sync.WaitGroup) {
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

func (b *Bot) pollUpdates() {
	log.Printf("Try to get updates...\n")

	values := url.Values{}
	values.Add("timeout", pollTimeout)
	values.Add("offset", fmt.Sprintf("%d", atomic.LoadUint64(&b.offset)))

	var updates Updates
	if err := b.postRequest("getUpdates", values, &updates); err != nil {
		log.Println(err)
		return
	}
	l := len(updates.Result)
	if l > 0 {
		// get last update id, increment it and store value into offset
		atomic.StoreUint64(&b.offset, updates.Result[l-1].UpdateId+1)
		b.updates <- updates.Result
	}
}

func (b *Bot) processMessage(msg Message) {
	log.Printf("Process message: %#v\n\n", msg)

	b.replies <- Message{
		Id: msg.Id,
		Chat: Chat{
			Id: msg.Chat.Id,
		},
		// TODO handlers
		Text: msg.Text + "!!!",
	}
}

func (b *Bot) sendReply(msg Message) {
	log.Printf("Send reply: %#v\n\n", msg)

	values := url.Values{}
	values.Add("chat_id", fmt.Sprintf("%d", msg.Chat.Id))
	values.Add("text", msg.Text)
	values.Add("reply_to_message_id", fmt.Sprintf("%d", msg.Id))

	var result Result
	if err := b.postRequest("sendMessage", values, &result); err != nil {
		log.Println(err)
		return
	}
}

func (b *Bot) postRequest(command string, values url.Values, result interface{}) error {
	commandUrl := fmt.Sprintf(apiURL, b.token, command)
	resp, err := http.PostForm(commandUrl, values)
	if err != nil {
		return errors.New(fmt.Sprintf("Execute of %s error %#v", command, err))
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return errors.New(fmt.Sprintf("StatusCode of %s not OK [%d]\n", command, resp.StatusCode))
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return errors.New(fmt.Sprintf("Read response of %s error: %#v\n", command, err))
	}

	if err := json.Unmarshal(body, result); err != nil {
		return errors.New(fmt.Sprintf("Unmarshal body of %s error: %#v\n", command, err))
	}
	if v, ok := result.(Responsable); ok {
		if r := v.GetResponse(); !r.Ok {
			return errors.New(fmt.Sprintf("%s error: error_code=%d, description=%s\n", command, r.ErrorCode, r.Description))
		}
	}
	return nil
}
