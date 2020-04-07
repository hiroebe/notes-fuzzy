package main

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/jessevdk/go-flags"
	"github.com/ktr0731/go-fuzzyfinder"
	notesCli "github.com/rhysd/notes-cli"
)

var opts struct {
	Edit bool `short:"e" long:"edit" description:"Edit selected note"`
}

func main() {
	flags.NewParser(&opts, flags.IgnoreUnknown).Parse()

	config, err := notesCli.NewConfig()
	if err != nil {
		panic(err)
	}
	var buf bytes.Buffer
	cmd := &notesCli.ListCmd{
		Config: config,
		Out:    &buf,
	}
	if err := cmd.Do(); err != nil {
		panic(err)
	}
	paths := strings.Split(strings.Trim(buf.String(), "\n"), "\n")
	notes := make([]*notesCli.Note, 0, len(paths))
	for _, path := range paths {
		note, err := notesCli.LoadNote(path, config)
		if err != nil {
			continue
		}
		notes = append(notes, note)
	}

	itemFunc := func(i int) string {
		n := notes[i]
		return fmt.Sprintf("[%s] %s {%s}", n.Category, n.Title, strings.Join(n.Tags, ", "))
	}
	previewFunc := func(i, width, height int) string {
		if i == -1 {
			return "no result"
		}
		body, _, err := notes[i].ReadBodyLines(height)
		if err != nil {
			return err.Error()
		}
		return body
	}
	idx, err := fuzzyfinder.Find(notes, itemFunc, fuzzyfinder.WithPreviewWindow(previewFunc))
	if err == fuzzyfinder.ErrAbort {
		return
	} else if err != nil {
		fmt.Println(err)
	}

	path := notes[idx].FilePath()
	if opts.Edit {
		if config.EditorCmd == "" {
			fmt.Println("Editor is not set.")
			return
		}
		cmd := exec.Command(config.EditorCmd, path)
		cmd.Stdin = os.Stdin
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		cmd.Dir = config.HomePath
		if err := cmd.Run(); err != nil {
			panic(err)
		}
	} else {
		fmt.Println(path)
	}
}
