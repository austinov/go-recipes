package dao

type Band struct {
	Id     string
	Name   string
	Events []Event
}

type Event struct {
	Title string
	From  int64
	To    int64
	City  string
	Venue string
	Link  string
	Img   string
}
