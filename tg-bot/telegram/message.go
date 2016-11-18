package telegram

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

type SendResult struct {
	Response
	Message Message `json:"result"`
}
