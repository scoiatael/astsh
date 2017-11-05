package main

import (
	"log"

	"github.com/chzyer/readline"
)

type State struct {
	ShouldBreak bool
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

func (ctx *Context) Loop() error {
	for !ctx.CurrentState.ShouldBreak {
		if err := ctx.Run(); err != nil {
			return err
		}
	}
	return nil
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

	rl.ResetHistory()
	log.SetOutput(rl.Stderr())

	rl.SetPrompt("> ")

	for {
		ln := rl.Line()
		if ln.CanContinue() {
			continue
		}
		if ln.CanBreak() {
			ctx.State().ShouldBreak = true
			break
		}
		// TODO: Tokenize, parse, execute
		if ln.Line == "" {
			continue
		}
		if ln.Line == "rng" {
			rng, err := ctx.NewRanger()
			if err != nil {
				return err
			}
			if err := ctx.NextRunner(rng); err != nil {
				return err
			}
			break
		}
		log.Println("!", ln.Line)
	}
	rl.Clean()
	return nil
}

func main() {
	ctx := Context{}
	if err := ctx.Loop(); err != nil {
		panic(err)
	}
}
