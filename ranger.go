package main

import (
	"github.com/gdamore/tcell"
	"github.com/gdamore/tcell/encoding"
	"github.com/mattn/go-runewidth"
)

type unit struct{}
type Ranger struct {
	screen    tcell.Screen
	textStyle tcell.Style
	quit      chan unit
}

type TextBox struct {
	StartX int
	row    int
}

func puts(s tcell.Screen, style tcell.Style, x, y int, str string) {
	i := 0
	var deferred []rune
	dwidth := 0
	for _, r := range str {
		switch runewidth.RuneWidth(r) {
		case 0:
			if len(deferred) == 0 {
				deferred = append(deferred, ' ')
				dwidth = 1
			}
		case 1:
			if len(deferred) != 0 {
				s.SetContent(x+i, y, deferred[0], deferred[1:], style)
				i += dwidth
			}
			deferred = nil
			dwidth = 1
		case 2:
			if len(deferred) != 0 {
				s.SetContent(x+i, y, deferred[0], deferred[1:], style)
				i += dwidth
			}
			deferred = nil
			dwidth = 2
		}
		deferred = append(deferred, r)
	}
	if len(deferred) != 0 {
		s.SetContent(x+i, y, deferred[0], deferred[1:], style)
		i += dwidth
	}
}

func (txt *TextBox) PutLn(rng *Ranger, str string) {
	puts(rng.screen, rng.textStyle, txt.StartX, txt.row, str)
	txt.row++
}

func (rng *Ranger) ioLoop() {
	for {
		ev := rng.screen.PollEvent()
		switch ev := ev.(type) {
		case *tcell.EventKey:
			switch ev.Key() {
			case tcell.KeyEscape, tcell.KeyEnter:
				close(rng.quit)
				return
			case tcell.KeyCtrlL:
				rng.screen.Sync()
			}
		case *tcell.EventResize:
			rng.screen.Sync()
		}
	}
}

func (rng *Ranger) Run(ctx RunnerContext) (err error) {
	rng.screen, err = tcell.NewScreen()
	if err != nil {
		return
	}

	encoding.Register()

	if err = rng.screen.Init(); err != nil {
		return
	}

	rng.screen.Clear()
	txt := TextBox{StartX: 1}
	txt.PutLn(rng, "English:   October")
	txt.PutLn(rng, "Icelandic: október")
	txt.PutLn(rng, "Arabic:    أكتوبر")
	txt.PutLn(rng, "Russian:   октября")
	txt.PutLn(rng, "Greek:     Οκτωβρίου")
	txt.PutLn(rng, "Chinese:   十月 (note, two double wide characters)")
	txt.PutLn(rng, "Combining: A\u030a (should look like Angstrom)")
	txt.PutLn(rng, "Emoticon:  \U0001f618 (blowing a kiss)")
	txt.PutLn(rng, "Airplane:  \u2708 (fly away)")
	txt.PutLn(rng, "Command:   \u2318 (mac clover key)")
	txt.PutLn(rng, "Enclose:   !\u20e3 (should be enclosed exclamation)")
	rng.screen.Show()
	go rng.ioLoop()

	<-rng.quit
	rng.screen.Fini()
	return nil
}

func (ctx *Context) NewRanger() (rng *Ranger, err error) {
	rng = &Ranger{
		quit:      make(chan unit),
		textStyle: tcell.StyleDefault,
	}
	return
}
