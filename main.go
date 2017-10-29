package main

import (
	"fmt"
	"log"
	"os"

	"github.com/chzyer/readline"
	"github.com/gdamore/tcell"
	"github.com/gdamore/tcell/encoding"
)

type State struct {
}

type RunnerContext interface {
	State() *State
	NextRunner(Runner) error
	NewShell() (*Shell, error)
	NewRanger() (*Ranger, error)
}

type Runner interface {
	Run(RunnerContext) error
}

type Context struct {
	CurrentState State

	runner     Runner
	nextRunner *Runner
}

func (ctx *Context) State() *State {
	return &ctx.CurrentState
}

func (ctx *Context) Run() error {
	if ctx.nextRunner == nil {
		shell, err := ctx.NewShell()
		if err != nil {
			return err
		}
		if err := ctx.NextRunner(shell); err != nil {
			return err
		}
	}
	next := *ctx.nextRunner
	ctx.nextRunner = nil
	return next.Run(ctx)
}

func (ctx *Context) NextRunner(runner Runner) error {
	ctx.nextRunner = &runner
	return nil
}

type Ranger struct {
}

func (rng *Ranger) Run(ctx RunnerContext) error {
	s, e := tcell.NewScreen()
	if e != nil {
		fmt.Fprintf(os.Stderr, "%v\n", e)
		os.Exit(1)
	}

	encoding.Register()

	if e = s.Init(); e != nil {
		fmt.Fprintf(os.Stderr, "%v\n", e)
		os.Exit(1)
	}

	s.SetStyle(tcell.StyleDefault.
		Foreground(tcell.ColorBlack).
		Background(tcell.ColorWhite))
	s.Clear()

	quit := make(chan struct{})
	s.Show()
	go func() {
		for {
			ev := s.PollEvent()
			switch ev := ev.(type) {
			case *tcell.EventKey:
				switch ev.Key() {
				case tcell.KeyEscape, tcell.KeyEnter:
					close(quit)
					return
				case tcell.KeyCtrlL:
					s.Sync()
				}
			case *tcell.EventResize:
				s.Sync()
			}
		}
	}()
	<-quit

	s.Fini()
	return nil
}

func (ctx *Context) NewRanger() (*Ranger, error) {
	return &Ranger{}, nil
}

type Shell struct {
}

func (ctx *Context) NewShell() (*Shell, error) {
	return &Shell{}, nil
}

func (sh *Shell) Run(ctx RunnerContext) error {
	rl, err := readline.NewEx(&readline.Config{
		UniqueEditLine: true,
	})
	if err != nil {
		panic(err)
	}
	defer rl.Close()

	rl.SetPrompt("username: ")
	username, err := rl.Readline()
	if err != nil {
		return err
	}
	rl.ResetHistory()
	log.SetOutput(rl.Stderr())

	rl.SetPrompt(username + "> ")

	for {
		ln := rl.Line()
		if ln.CanContinue() {
			continue
		} else if ln.CanBreak() {
			break
		}
		log.Println(username+":", ln.Line)
		if ln.Line == "rng" {
			rng, err := ctx.NewRanger()
			if err != nil {
				return err
			}
			if err := ctx.NextRunner(rng); err != nil {
				return err
			}
		}
	}
	rl.Clean()
	return nil
}

func main() {
	ctx := Context{}
	for {
		if err := ctx.Run(); err != nil {
			panic(err)
		}
	}
}
