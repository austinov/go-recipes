package bot

type Message struct {
	Id          uint64       `json:"id"`
	Type        string       `json:"type"`
	Channel     string       `json:"channel"`
	Text        string       `json:"text"`
	Attachments []Attachment `json:"attachments"`
}

type ResponseRtmStart struct {
	Ok    bool         `json:"ok"`
	Error string       `json:"error"`
	Url   string       `json:"url"`
	Self  ResponseSelf `json:"self"`
}

type ResponseSelf struct {
	Id string `json:"id"`
}

type Attachment struct {
	Text       string   `json:"text"`
	Fallback   string   `json:"fallback"`
	CallbackId string   `json:"callback_id"`
	Color      string   `json:"color"`
	Type       string   `json:"attachment_type"`
	Actions    []Action `json:"actions"`
}

type Action struct {
	Name  string  `json:"name"`
	Text  string  `json:"text"`
	Type  string  `json:"type"`
	Value string  `json:"value"`
	Style string  `json:"style"`
	Cfm   Confirm `json:"confirm"`
}

type Confirm struct {
	Title   string `json:"title"`
	Text    string `json:"text"`
	Ok      string `json:"ok_text"`
	Dismiss string `json:"dismiss_text"`
}
