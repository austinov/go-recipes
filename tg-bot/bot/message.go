package bot

// These are structures represent the response of the Telegram API
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

type Responsable interface {
	GetResponse() Response
}

// Update represents an incoming updates.
// It implements the Responsable interface for unified data unmarshalling.
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

// Result represents an result of reply message.
// It implements the Responsable interface for unified data unmarshalling.
type Result struct {
	Response
	Message Message `json:"result"`
}

func (u Result) GetResponse() Response {
	return u.Response
}

// UpdateParams represents request parameters for "getUpdates" method.
type UpdateParams struct {
	Timeout int    `json:"timeout"`
	Offset  uint64 `json:"offset"`
}

// These are structures represent the reply for user
type Reply struct {
	ChatId           uint64 `json:"chat_id"`
	Text             string `json:"text"`
	ReplyToMessageId uint64 `json:"reply_to_message_id"`
	// ReplyMarkup is InlineKeyboardMarkup or ReplyKeyboardMarkup type.
	ReplyMarkup interface{} `json:"reply_markup,omitempty"`
}

// InlineKeyboardMarkup represents an inline keyboard.
type InlineKeyboardMarkup struct {
	InlineKeyboard [][]InlineKeyboardButton `json:"inline_keyboard"`
}

type InlineKeyboardButton struct {
	Text                         string `json:"text"`
	Url                          string `json:"url"`
	CallbackData                 string `json:"callback_data"`
	SwitchInlineQuery            string `json:"switch_inline_query"`
	SwitchInlineQueryCurrentChat string `json:"switch_inline_query_current_chat"`
}

// ReplyKeyboardMarkup represents a custom keyboard.
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
