package command

import (
	"fmt"
	"slices"

	"github.com/kyomel/audiofile/internal/interfaces"
)

type Parser struct{ commands []interfaces.Command }

func NewParser(commands []interfaces.Command) *Parser { return &Parser{commands: commands} }

func (p *Parser) Parse(args []string) error {
	if len(args) == 0 {
		help()
		return nil
	}
	name, rest := args[0], args[1:]

	idx := slices.IndexFunc(p.commands, func(c interfaces.Command) bool {
		return c.Name() == name
	})
	if idx < 0 {
		return fmt.Errorf("unknown subcommand: %s", name)
	}
	if err := p.commands[idx].ParseFlags(rest); err != nil {
		return err
	}
	return p.commands[idx].Run()
}

func help() {
	fmt.Println("usage: audiofile-cli <command> [flags]")
	fmt.Println()
	fmt.Println("commands:")
	fmt.Println("  upload -filename <file>   upload an audio file")
	fmt.Println("  get -id <id>              fetch metadata for one audio file")
	fmt.Println("  list                      list all audio files")
}
