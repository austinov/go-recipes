package telegram

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
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
	done    chan struct{}
}

func New(token string) *Bot {
	if token == "" {
		log.Fatal("Token is empty")
	}
	return &Bot{
		token: token,
		done:  make(chan struct{}),
	}
}

func (b *Bot) Start() {
	b.updates = make(chan []Update, numPollers)
	b.replies = make(chan Message, numSenders)
	defer close(b.updates)
	defer close(b.replies)

	go b.startPoll()
	// TODO multiple processes
	go b.startProcess()
	// TODO multiple senders
	go b.startReply()

	<-b.done
}

func (b *Bot) Stop() {
	close(b.done)
}

func (b *Bot) startPoll() {
	for {
		select {
		case <-b.done:
			return
		case <-time.After(pollDelay):
			b.poll()
		}
	}
}

func (b *Bot) startProcess() {
	for {
		select {
		case <-b.done:
			return
		case updates := <-b.updates:
			for _, update := range updates {
				go b.process(update)
			}
		}
	}
}

func (b *Bot) startReply() {
	for {
		select {
		case <-b.done:
			return
		case reply := <-b.replies:
			b.send(reply)
		}
	}
}

func (b *Bot) poll() {
	log.Printf("Try to get updates...\n")
	updateUrl := fmt.Sprintf(apiURL, b.token, "getUpdates")
	values := url.Values{}
	values.Add("timeout", pollTimeout)
	values.Add("offset", fmt.Sprintf("%d", b.offset))

	resp, err := http.PostForm(updateUrl, values)
	if err != nil {
		log.Printf("Post on getUpdates error: %#v\n", err)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		log.Printf("StatusCode of getUpdates not OK [%d]\n", resp.StatusCode)
		return
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Printf("Read response of getUpdates error: %#v\n", err)
		return
	}

	var updates Updates
	if err := json.Unmarshal(body, &updates); err != nil {
		log.Printf("Unmarshal body of getUpdates error: %#v\n", err)
		return
	}
	if !updates.Ok {
		log.Printf("getUpdates error: error_code=%d, description=%s\n", updates.ErrorCode, updates.Description)
		return
	}
	log.Printf("updates: %#v\n", updates)
	if len(updates.Result) > 0 {
		b.updates <- updates.Result
	}
}

func (b *Bot) process(update Update) {
	log.Printf("Update: %#v\n\n", update)
	// TODO thread safe
	b.offset = update.UpdateId + 1

	b.replies <- Message{
		Id: update.Message.Id,
		Chat: Chat{
			Id: update.Message.Chat.Id,
		},
		Text: update.Message.Text + "!!!",
	}
}

func (b *Bot) send(msg Message) {
	sendUrl := fmt.Sprintf(apiURL, b.token, "sendMessage")
	values := url.Values{}
	values.Add("chat_id", fmt.Sprintf("%d", msg.Chat.Id))
	values.Add("text", msg.Text)
	values.Add("reply_to_message_id", fmt.Sprintf("%d", msg.Id))

	resp, err := http.PostForm(sendUrl, values)
	if err != nil {
		log.Printf("Post on sendMessage error: %#v", err)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		log.Printf("StatusCode of sendMessage not OK [%d]\n", resp.StatusCode)
		return
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Printf("Read response of sendMessag error: %#v\n", err)
		return
	}

	var result SendResult
	if err := json.Unmarshal(body, &result); err != nil {
		log.Printf("Unmarshal body of sendMessage error: %#v\n", err)
		return
	}
	if !result.Ok {
		log.Printf("getUpdates error: error_code=%d, description=%s\n", result.ErrorCode, result.Description)
		return
	}
	log.Printf("Result: %#v\n\n", result)
}
