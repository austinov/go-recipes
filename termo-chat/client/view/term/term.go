package term

import (
	"errors"
	"fmt"
	"log"
	"sample/cui/client/handler"
	"sample/cui/client/view"
	"strings"

	"github.com/jroimartin/gocui"
)

type (
	Message struct {
		From string
		Msg  string
	}
)

var (
	done  = make(chan bool)
	msgs  = make(chan Message)
	peers = make(chan []string)
	//wg    sync.WaitGroup
)

type OnSendMessage func(message string) error

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

func (v *termView) ReceiveMessage(from, message string) {
	// TODO
	msgs <- Message{
		From: from,
		Msg:  message,
	}
}

func (v *termView) UpdatePeers(p []string) {
	// TODO
	peers <- p
}

func (v *termView) Show() <-chan struct{} {
	quit := make(chan struct{})
	go func() {
		v.gui = gocui.NewGui()
		if err := v.gui.Init(); err != nil {
			log.Panic(err)
		}
		defer v.gui.Close()

		v.gui.SetLayout(layout)

		if err := v.keybindings(); err != nil {
			log.Panic(err)
		}
		v.gui.SelBgColor = gocui.ColorGreen
		v.gui.SelFgColor = gocui.ColorBlack
		v.gui.Cursor = true

		//wg.Add(2)
		go receiveMsg(v.gui)
		go join(v.gui)

		if err := v.gui.MainLoop(); err != nil && err != gocui.ErrQuit {
			log.Panic(err)
		}
		quit <- struct{}{}
	}()
	return quit
}

func (v *termView) Quit() {
	// TODO
	v.gui.Close()
}

func (v *termView) sendMsg(g *gocui.Gui, vv *gocui.View) error {
	if v != nil {
		msg := strings.TrimRight(vv.Buffer(), "\n")
		if msg == "" {
			return nil
		}
		v.sender.SendMessage(msg)
		msgs <- Message{
			From: v.peerName,
			Msg:  msg,
		}
		return clearCmd(vv)
	}
	return nil
}

func receiveMsg(g *gocui.Gui) {
	//defer wg.Done()
	for {
		select {
		case <-done:
			return
		case m := <-msgs:
			g.Execute(func(g *gocui.Gui) error {
				v, err := g.View("main")
				if err != nil {
					return err
				}
				fmt.Fprintln(v, m.From, ":", m.Msg)
				return nil
			})
		}
	}
}

func join(g *gocui.Gui) {
	//defer wg.Done()
	for {
		select {
		case <-done:
			return
		case p := <-peers:
			g.Execute(func(g *gocui.Gui) error {
				v, err := g.View("side")
				if err != nil {
					return err
				}
				v.Clear()
				v.SetCursor(0, 0)
				for _, peerName := range p {
					fmt.Fprintln(v, peerName)
				}
				return nil
			})
		}
	}
}

func layout(g *gocui.Gui) error {
	maxX, maxY := g.Size()
	if v, err := g.SetView("side", 0, 0, int(0.2*float32(maxX)), maxY-5); err != nil {
		if err != gocui.ErrUnknownView {
			return err
		}
		v.Title = " Side "
		v.Frame = true
		if err := g.SetCurrentView("side"); err != nil {
			return err
		}
	}
	if v, err := g.SetView("main", int(0.2*float32(maxX)), 0, maxX-1, maxY-5); err != nil {
		if err != gocui.ErrUnknownView {
			return err
		}
		v.Title = " Main "
		v.Frame = true
		v.Wrap = true
		v.Autoscroll = true
		if err := g.SetCurrentView("main"); err != nil {
			return err
		}
	}
	if v, err := g.SetView("cmdline", 0, maxY-5, maxX-1, maxY-1); err != nil {
		if err != gocui.ErrUnknownView {
			return err
		}
		v.Title = " Type message bellow "
		v.Frame = true
		v.Wrap = true
		v.Editable = true
		v.Wrap = true
		v.Autoscroll = false
		if err := g.SetCurrentView("cmdline"); err != nil {
			return err
		}
	}

	return nil
}

func (v *termView) keybindings() error {
	if err := v.gui.SetKeybinding("", gocui.KeyCtrlC, gocui.ModNone, quit); err != nil {
		return err
	}
	if err := v.gui.SetKeybinding("side", gocui.KeyCtrlSpace, gocui.ModNone, nextView); err != nil {
		return err
	}
	if err := v.gui.SetKeybinding("main", gocui.KeyCtrlSpace, gocui.ModNone, nextView); err != nil {
		return err
	}
	if err := v.gui.SetKeybinding("cmdline", gocui.KeyCtrlSpace, gocui.ModNone, nextView); err != nil {
		return err
	}
	if err := v.gui.SetKeybinding("side", gocui.KeyArrowDown, gocui.ModNone, cursorDown); err != nil {
		return err
	}
	if err := v.gui.SetKeybinding("side", gocui.KeyArrowUp, gocui.ModNone, cursorUp); err != nil {
		return err
	}
	if err := v.gui.SetKeybinding("main", gocui.KeyArrowDown, gocui.ModNone, cursorDown); err != nil {
		return err
	}
	if err := v.gui.SetKeybinding("main", gocui.KeyArrowUp, gocui.ModNone, cursorUp); err != nil {
		return err
	}
	if err := v.gui.SetKeybinding("cmdline", gocui.KeyArrowDown, gocui.ModNone, cursorDown); err != nil {
		return err
	}
	if err := v.gui.SetKeybinding("cmdline", gocui.KeyArrowUp, gocui.ModNone, cursorUp); err != nil {
		return err
	}
	if err := v.gui.SetKeybinding("cmdline", gocui.KeyEnter, gocui.ModNone, v.sendMsg); err != nil {
		return err
	}
	return nil
}

func nextView(g *gocui.Gui, v *gocui.View) error {
	if v == nil || v.Name() == "main" {
		return g.SetCurrentView("cmdline")
	} else if v.Name() == "side" {
		return g.SetCurrentView("main")
	}
	return g.SetCurrentView("side")
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
	done <- true
	return gocui.ErrQuit
}
