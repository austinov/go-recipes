package term

import (
	"errors"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/austinov/go-recipes/termo-chat/client/handler"
	"github.com/austinov/go-recipes/termo-chat/client/view"

	"github.com/jroimartin/gocui"
)

type (
	message struct {
		kind byte
		from string
		msg  string
	}
)

var (
	done  = make(chan struct{})
	msgs  = make(chan message)
	room  = make(chan string)
	peers = make(chan []string)
)

type termView struct {
	gui      *gocui.Gui
	peerName string
	sender   handler.Sender
}

func New(peerName string, sender handler.Sender) view.View {
	if sender == nil {
		log.Panic(errors.New("Sender must assigned"))
	}
	return &termView{
		peerName: peerName,
		sender:   sender,
	}
}

func (v *termView) ViewRoom(id string) {
	room <- id
}

func (v *termView) ViewMessage(kind byte, from, msg string) {
	msgs <- message{
		kind: kind,
		from: from,
		msg:  msg,
	}
}

func (v *termView) UpdatePeers(p []string) {
	peers <- p
}

func (v *termView) Show() <-chan struct{} {
	quit := make(chan struct{})
	go func() {
		v.gui = gocui.NewGui()
		if err := v.gui.Init(); err != nil {
			log.Panic(err)
		}
		defer func() {
			v.gui.Close()
			quit <- struct{}{}
		}()

		v.gui.SetLayout(layout)

		if err := v.keybindings(); err != nil {
			log.Panic(err)
		}
		v.gui.SelBgColor = gocui.ColorGreen
		v.gui.SelFgColor = gocui.ColorBlack
		v.gui.Cursor = true

		go receiveMsg(v.gui)
		go join(v.gui)

		if err := v.gui.MainLoop(); err != nil && err != gocui.ErrQuit {
			log.Panic(err)
		}
	}()
	return quit
}

func (v *termView) Quit() {
	v.gui.Close()
}

func (v *termView) sendMsg(g *gocui.Gui, vv *gocui.View) error {
	if v != nil {
		msg := strings.TrimRight(vv.Buffer(), "\n")
		if msg == "" {
			return nil
		}
		if err := v.sender.SendMessage(msg); err != nil {
			v.ViewMessage(view.ErrorMessage, "", err.Error())
		} else {
			v.ViewMessage(view.ChatMessage, v.peerName, msg)
		}
		return clearCmd(vv)
	}
	return nil
}

func receiveMsg(g *gocui.Gui) {
	currDate := func() string {
		return time.Now().Format("2 Jan 15:04")
	}
	for {
		select {
		case _, ok := <-done:
			if !ok {
				return
			}
		case m := <-msgs:
			g.Execute(func(g *gocui.Gui) error {
				v, err := g.View("chat")
				if err != nil {
					return err
				}
				if m.kind == view.TailMessage {
					fmt.Fprintf(v, "%s", m.msg)
				} else if m.kind != view.ChatMessage {
					fmt.Fprintf(v, "\n%v %s: %s", currDate(), view.MessageKinds[m.kind], m.msg)
				} else {
					fmt.Fprintf(v, "\n%v %s: %s", currDate(), m.from, m.msg)
				}
				return nil
			})
		}
	}
}

func join(g *gocui.Gui) {
	for {
		select {
		case _, ok := <-done:
			if !ok {
				return
			}
		case p := <-peers:
			g.Execute(func(g *gocui.Gui) error {
				v, err := g.View("peers")
				if err != nil {
					return err
				}
				v.Clear()
				v.SetCursor(0, 0)
				fmt.Fprintf(v, "\n\n")
				for _, peerName := range p {
					fmt.Fprintln(v, peerName)
				}
				return nil
			})
		case id := <-room:
			g.Execute(func(g *gocui.Gui) error {
				v, err := g.View("room")
				if err != nil {
					return err
				}
				v.Clear()
				v.SetCursor(0, 0)
				fmt.Fprintf(v, "\n\n%s", id)
				return nil
			})
		}
	}
}

func layout(g *gocui.Gui) error {
	maxX, maxY := g.Size()
	if v, err := g.SetView("room", 0, 0, int(0.2*float32(maxX)), int(0.1*float32(maxY))); err != nil {
		if err != gocui.ErrUnknownView {
			return err
		}
		v.Title = " Room "
		v.Frame = true
		if err := g.SetCurrentView("room"); err != nil {
			return err
		}
	}
	if v, err := g.SetView("peers", 0, int(0.1*float32(maxY)), int(0.2*float32(maxX)), int(0.8*float32(maxY))); err != nil {
		if err != gocui.ErrUnknownView {
			return err
		}
		v.Title = " Peers "
		v.Frame = true
		if err := g.SetCurrentView("peers"); err != nil {
			return err
		}
	}
	if v, err := g.SetView("hints", 0, int(0.8*float32(maxY)), int(0.2*float32(maxX)), int(0.9*float32(maxY))); err != nil {
		if err != gocui.ErrUnknownView {
			return err
		}
		v.Title = " Hints "
		v.Frame = true
		if err := g.SetCurrentView("hints"); err != nil {
			return err
		}
		fmt.Fprintf(v, "\n\n")
		fmt.Fprintf(v, "Enter        Send message\n")
		fmt.Fprintf(v, "Ctrl+Space   Navigate between spaces\n")
		fmt.Fprintf(v, "Ctrl+C       Close chat")
	}
	if v, err := g.SetView("chat", int(0.2*float32(maxX)), 0, maxX-1, int(0.9*float32(maxY))); err != nil {
		if err != gocui.ErrUnknownView {
			return err
		}
		v.Title = " Chat "
		v.Frame = true
		v.Autoscroll = true
		if err := g.SetCurrentView("chat"); err != nil {
			return err
		}
	}
	if v, err := g.SetView("cmd", 0, int(0.9*float32(maxY)), maxX-1, maxY-1); err != nil {
		if err != gocui.ErrUnknownView {
			return err
		}
		v.Title = " Type your message bellow "
		v.Frame = true
		v.Autoscroll = false
		v.Editable = true
		if err := g.SetCurrentView("cmd"); err != nil {
			return err
		}
	}
	return nil
}

func (v *termView) keybindings() error {
	if err := v.gui.SetKeybinding("", gocui.KeyCtrlC, gocui.ModNone, quit); err != nil {
		return err
	}
	// set general keybindings
	for _, name := range []string{"room", "peers", "hints", "chat", "cmd"} {
		if err := v.gui.SetKeybinding(name, gocui.KeyCtrlSpace, gocui.ModNone, nextView); err != nil {
			return err
		}
		if err := v.gui.SetKeybinding(name, gocui.KeyArrowDown, gocui.ModNone, cursorDown); err != nil {
			return err
		}
		if err := v.gui.SetKeybinding(name, gocui.KeyArrowUp, gocui.ModNone, cursorUp); err != nil {
			return err
		}
	}
	// set keybinding for input text and send on Enter
	if err := v.gui.SetKeybinding("cmd", gocui.KeyEnter, gocui.ModNone, v.sendMsg); err != nil {
		return err
	}
	return nil
}

func nextView(g *gocui.Gui, v *gocui.View) error {
	if v == nil || v.Name() == "chat" {
		return g.SetCurrentView("cmd")
	} else if v.Name() == "peers" {
		return g.SetCurrentView("chat")
	}
	return g.SetCurrentView("peers")
}

func cursorDown(g *gocui.Gui, v *gocui.View) error {
	if v != nil {
		cx, cy := v.Cursor()
		if err := v.SetCursor(cx, cy+1); err != nil {
			ox, oy := v.Origin()
			if err := v.SetOrigin(ox, oy+1); err != nil {
				return err
			}
		}
	}
	return nil
}

func cursorUp(g *gocui.Gui, v *gocui.View) error {
	if v != nil {
		ox, oy := v.Origin()
		cx, cy := v.Cursor()
		if err := v.SetCursor(cx, cy-1); err != nil && oy > 0 {
			if err := v.SetOrigin(ox, oy-1); err != nil {
				return err
			}
		}
	}
	return nil
}

func clearCmd(v *gocui.View) error {
	v.Clear()
	return v.SetCursor(0, 0)
}

func quit(g *gocui.Gui, v *gocui.View) error {
	close(done)
	return gocui.ErrQuit
}
