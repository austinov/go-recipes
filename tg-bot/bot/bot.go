package bot

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"math/rand"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

const (
	apiURL          = "https://api.telegram.org/bot%s/%s"
	pollTimeout     = 30 // seconds
	pollDelay       = 1 * time.Second
	numPollers      = 2
	numSenders      = 3
	reverseCommand  = "/reverse"
	searchCommand   = "/search"
	roulleteCommand = "/roullete"
)

type Bot struct {
	token   string
	offset  uint64
	updates chan []Update
	replies chan Reply
	done    chan struct{}
}

func New(token string) *Bot {
	if token == "" {
		log.Fatal("Token is empty")
	}
	return &Bot{
		token: token,
	}
}

// Start runs bot to receive incoming updates using long polling,
// to handle messages and to send replies into chat.
// It blocks until the "Stop" method is called.
func (b *Bot) Start() {
	// reset offset to get all lost messages
	atomic.StoreUint64(&b.offset, 0)

	b.updates = make(chan []Update, numPollers)
	b.replies = make(chan Reply, numSenders)
	b.done = make(chan struct{})

	var wg sync.WaitGroup

	wg.Add(1)
	go b.pollUpdates(&wg)
	for i := 0; i < numPollers; i++ {
		wg.Add(1)
		go b.processUpdates(&wg)
	}
	for i := 0; i < numSenders; i++ {
		wg.Add(1)
		go b.processReplies(&wg)
	}
	wg.Wait()
}

// Stop initiates a stop of the bot.
func (b *Bot) Stop() {
	close(b.done)
}

func (b *Bot) pollUpdates(wg *sync.WaitGroup) {
	defer wg.Done()
	for {
		select {
		case _, ok := <-b.done:
			if !ok {
				return
			}
		default:
			b.poll()
		}
	}
}

func (b *Bot) processUpdates(wg *sync.WaitGroup) {
	defer wg.Done()
	for {
		select {
		case _, ok := <-b.done:
			if !ok {
				return
			}
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

func (b *Bot) processReplies(wg *sync.WaitGroup) {
	defer wg.Done()
	for {
		select {
		case _, ok := <-b.done:
			if !ok {
				return
			}
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

	ur := UpdateParams{
		Timeout: pollTimeout,
		Offset:  atomic.LoadUint64(&b.offset),
	}
	var updates Updates
	if err := b.callAPI("getUpdates", ur, &updates); err != nil {
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
	log.Printf("Process message: %#v\n", msg)

	trimCommand := func(command string) Message {
		msg.Text = strings.Trim(strings.TrimPrefix(msg.Text, command), " ")
		return msg
	}

	if strings.HasPrefix(msg.Text, reverseCommand) {
		b.replies <- b.reverseHandler(trimCommand(reverseCommand))
	} else if strings.HasPrefix(msg.Text, searchCommand) {
		b.replies <- b.searchHandler(trimCommand(searchCommand))
	} else if strings.HasPrefix(msg.Text, roulleteCommand) {
		b.replies <- b.roulleteHandler(trimCommand(roulleteCommand))
	} else {
		if i, err := strconv.Atoi(msg.Text); err == nil && i > 0 && i < 11 {
			b.replies <- b.spinRoullete(msg, i)
		} else {
			b.replies <- b.helpHandler(msg)
		}
	}
}

func (b *Bot) sendReply(reply Reply) {
	log.Printf("Send reply: %#v\n", reply)

	var result Result
	if err := b.callAPI("sendMessage", reply, &result); err != nil {
		log.Println(err)
	}
}

func (b *Bot) callAPI(method string, data interface{}, result interface{}) error {
	commandUrl := fmt.Sprintf(apiURL, b.token, method)

	jsonData, err := json.Marshal(data)
	if err != nil {
		return errors.New(fmt.Sprintf("Marshal data of %s error: %#v", method, err))
	}

	req, err := http.NewRequest("POST", commandUrl, bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return errors.New(fmt.Sprintf("Execute of %s error %#v", method, err))
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return errors.New(fmt.Sprintf("Request %s failed with code %d", method, resp.StatusCode))
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return errors.New(fmt.Sprintf("Read response of %s error: %#v", method, err))
	}

	if err := json.Unmarshal(body, result); err != nil {
		return errors.New(fmt.Sprintf("Unmarshal body of %s error: %#v", method, err))
	}
	if v, ok := result.(Responsable); ok {
		if r := v.GetResponse(); !r.Ok {
			return errors.New(fmt.Sprintf("%s error: error_code=%d, description=%s", method, r.ErrorCode, r.Description))
		}
	}
	return nil
}

// helpHandler returns a reply containing help text.
func (b *Bot) helpHandler(msg Message) Reply {
	cmd := "Please, use the follow commands:\n" +
		"/help - show this text\n" +
		"/reverse [text] - reverse any text\n" +
		"/search [text] - search any text in some engine\n" +
		"/roullete - play roullete"
	return Reply{
		ChatId: msg.Chat.Id,
		Text:   cmd,
	}
}

// reverseHandler returns reversed incoming text.
func (b *Bot) reverseHandler(msg Message) Reply {
	reverse := func(s string) string {
		r := []rune(s)
		for i, j := 0, len(r)-1; i < len(r)/2; i, j = i+1, j-1 {
			r[i], r[j] = r[j], r[i]
		}
		return string(r)
	}
	return Reply{
		ChatId: msg.Chat.Id,
		Text:   reverse(msg.Text),
		//ReplyToMessageId: msg.Id,
	}
}

// searchHandler demonstrates the use of inline keyboard buttons.
func (b *Bot) searchHandler(msg Message) Reply {
	kb := []InlineKeyboardButton{
		InlineKeyboardButton{
			Text: "Yandex",
			Url:  "https://yandex.ru/search/?text=" + msg.Text,
		},
		InlineKeyboardButton{
			Text: "Google",
			Url:  "https://www.google.com/search?q=" + msg.Text,
		},
	}
	var keyboard [][]InlineKeyboardButton
	keyboard = append(keyboard, kb)
	rm := InlineKeyboardMarkup{
		InlineKeyboard: keyboard,
	}
	return Reply{
		ChatId:      msg.Chat.Id,
		Text:        "Please select an engine",
		ReplyMarkup: rm,
	}

}

// roulleteHandler demonstrates the use of custom keyboard buttons.
func (b *Bot) roulleteHandler(msg Message) Reply {
	kb := make([]KeyboardButton, 10)
	for i := 1; i < 11; i++ {
		kb[i-1] = KeyboardButton{
			Text: fmt.Sprintf("%d", i),
		}
	}
	var keyboard [][]KeyboardButton
	keyboard = append(keyboard, kb)
	rm := ReplyKeyboardMarkup{
		Keyboard:       keyboard,
		ResizeKeyboard: true,
	}
	return Reply{
		ChatId:      msg.Chat.Id,
		Text:        "Please select a number",
		ReplyMarkup: rm,
	}
}

// spinRoullete returns reply for roullete game.
func (b *Bot) spinRoullete(msg Message, choice int) Reply {
	r := rand.New(rand.NewSource(time.Now().UnixNano())).Intn(9) + 1
	text := fmt.Sprintf("You missed the number - %d!", r)
	if choice == r {
		text = fmt.Sprintf("You guessed the number - %d!", choice)
	}
	return Reply{
		ChatId:           msg.Chat.Id,
		ReplyToMessageId: msg.Id,
		Text:             text,
	}
}
