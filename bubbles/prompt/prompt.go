package prompt

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/erikgeiser/promptkit/textinput"

	"github.com/khulnasoft-labs/bubble"
)

type Prompt struct {
	complete bool
	babble.PromptWriter
	tea.Model
	value func() (string, error)
	*textinput.TextInput
}

func New(prompter babble.PromptWriter) *Prompt {
	// candidates: ‣⧗⧖⌛💬ⓘ■⬛⬢◼⧓►❖
	spec := textinput.New(" ❖ " + prompter.PromptMessage())
	spec.Hidden = prompter.IsSensitive()
	spec.InputWidth = 12
	spec.HideMask = '●' // candidates: ●•✦*⬤⁕
	spec.Template = `
	{{- Bold .Prompt }} {{ .Input -}}
	{{- if .ValidationError }} {{ Foreground "1" (Bold "✘") }}
	{{- else }} {{ Foreground "2" (Bold "✔") }}
	{{- end -}}
    {{- if .ValidationError }} {{ Italic (Foreground "240" (ErrorStr (.ValidationError))) }}
    {{- end -}}
	`
	spec.Validate = func(s string) error {
		if len(strings.TrimSpace(s)) == 0 {
			return fmt.Errorf("value required")
		}

		return prompter.Validate(s)
	}
	spec.ExtendedTemplateFuncs = map[string]any{
		"ErrorStr": func(err error) string {
			return err.Error()
		},
	}
	specModel := textinput.NewModel(spec)
	teaModel := &Prompt{
		PromptWriter: prompter,
		Model:        specModel,
		value:        specModel.Value,
		TextInput:    spec,
	}
	return teaModel
}

func (m *Prompt) View() string {
	return strings.TrimRight(m.Model.View(), "\n")
}

func (m *Prompt) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if m.complete {
		return m, nil
	}

	switch msg := msg.(type) { //nolint:gocritic
	case tea.KeyMsg:
		if msg.String() == "enter" {
			v, err := m.value()
			if err != nil {
				// log.Errorf("unable to get prompt value: %+v", err)
				return m, nil
			}
			if err := m.PromptWriter.Respond(v); err != nil {
				// log.Errorf("unable to set prompt: %+v", err)
				return m, nil
			}

			m.Template = textinput.DefaultResultTemplate // don't show any validation once something has been entered
			m.Model.Update(msg)                          // update the state but ignore any future messages
			m.complete = true                            // don't respond to any other update events

			return m, nil
		}
	}

	_, cmd := m.Model.Update(msg)
	return m, cmd
}

func (m *Prompt) RunPrompt() (string, error) {
	value, err := m.TextInput.RunPrompt()
	if err == nil {
		err = m.PromptWriter.Respond(value)
	}
	return value, err
}
