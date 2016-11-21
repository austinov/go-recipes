package telegram

type Responsable interface {
	GetResponse() Response
}

type Response struct {
	Ok          bool   `json:"ok"`
	ErrorCode   int    `json:"error_code"`
	Description string `json:"description"`
}

type Message struct {
	Id   uint64 `json:"message_id"`
	From User   `json:"from"`
	Date uint64 `json:"date"`
	Chat Chat   `json:"chat"`
	Text string `json:"text"`
}

type User struct {
	Id        uint64 `json:"id"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	UserName  string `json:"username"`
}

type Chat struct {
	Id        uint64 `json:"id"`
	Type      string `json:"type"`
	Title     string `json:"title"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	UserName  string `json:"username"`
}

type Update struct {
	UpdateId uint64  `json:"update_id"`
	Message  Message `json:"message"`
}

type Updates struct {
	Response
	Result []Update `json:"result"`
}

func (u Updates) GetResponse() Response {
	return u.Response
}

type Result struct {
	Response
	Message Message `json:"result"`
}

func (u Result) GetResponse() Response {
	return u.Response
}

type UpdateParams struct {
	Timeout int    `json:"timeout"`
	Offset  uint64 `json:"offset"`
}

type Reply struct {
	ChatId           uint64 `json:"chat_id"`
	Text             string `json:"text"`
	ReplyToMessageId uint64 `json:"reply_to_message_id"`
	// ReplyMarkup are InlineKeyboardMarkup or ReplyKeyboardMarkup
	ReplyMarkup interface{} `json:"reply_markup,omitempty"`
}

// reply markups
type InlineKeyboardMarkup struct {
	InlineKeyboard [][]InlineKeyboardButton `json:"inline_keyboard"`
}

type ReplyKeyboardMarkup struct {
	Keyboard        [][]KeyboardButton `json:"keyboard"`
	ResizeKeyboard  bool               `json:"resize_keyboard"`
	OneTimeKeyboard bool               `json:"one_time_keyboard"`
	Selective       bool               `json:"selective"`
}

type KeyboardButton struct {
	Text            string `json:"text"`
	RequestContact  bool   `json:"request_contact"`
	RequestLocation bool   `json:"request_location"`
}

type InlineKeyboardButton struct {
	Text                         string `json:"text"`
	Url                          string `json:"url"`
	CallbackData                 string `json:"callback_data"`
	SwitchInlineQuery            string `json:"switch_inline_query"`
	SwitchInlineQueryCurrentChat string `json:"switch_inline_query_current_chat"`
}
