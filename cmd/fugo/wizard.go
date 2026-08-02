package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

// wizard prompts the user on stdin for values a caller didn't already pin
// down via flags — used by 'fugo init' to ask for the project name,
// organization, template, and theme when run interactively. It reuses a
// single buffered reader across prompts so line input isn't lost between
// calls.
type wizard struct {
	r *bufio.Reader
}

func newWizard() *wizard {
	return &wizard{r: bufio.NewReader(os.Stdin)}
}

// ask prompts label, showing def in brackets if non-empty, and returns the
// trimmed line the user typed, or def if they just pressed enter.
func (w *wizard) ask(label, def string) string {
	if def != "" {
		out.printf("%s %s: ", label, out.paint(cDim, "["+def+"]"))
	} else {
		out.printf("%s: ", label)
	}

	line, _ := w.r.ReadString('\n')
	line = strings.TrimSpace(line)
	if line == "" {
		return def
	}

	return line
}

// askRequired re-prompts until the user types something.
func (w *wizard) askRequired(label string) string {
	for {
		if v := w.ask(label, ""); v != "" {
			return v
		}

		out.warnf("required")
	}
}

// askChoice prompts among options (shown joined with "/"), re-prompting on
// an answer that isn't one of them. def (if non-empty) is used on a bare
// enter and must be one of options.
func (w *wizard) askChoice(label string, options []string, def string) string {
	prompt := fmt.Sprintf("%s (%s)", label, strings.Join(options, "/"))

	for {
		v := strings.ToLower(w.ask(prompt, def))
		for _, o := range options {
			if v == o {
				return o
			}
		}

		out.warnf("choose one of: %s", strings.Join(options, ", "))
	}
}

// askYesNo prompts a yes/no question, defaulting to def on a bare enter.
func (w *wizard) askYesNo(label string, def bool) bool {
	d := "Y/n"
	if !def {
		d = "y/N"
	}

	for {
		v := strings.ToLower(w.ask(fmt.Sprintf("%s (%s)", label, d), ""))
		switch v {
		case "":
			return def
		case "y", "yes":
			return true
		case "n", "no":
			return false
		}

		out.warnf("please answer y or n")
	}
}
