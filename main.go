package main

import (
	"fmt"
	"log"
	"math/rand"
	"os"
	"strconv"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/mtslzr/pokeapi-go"
	"github.com/mtslzr/pokeapi-go/structs"
)

type errMsg error

type model struct {
	spinner  spinner.Model
	quitting bool
	err      error
	pokemon  structs.Pokemon
}

var quitKeys = key.NewBinding(
	key.WithKeys("q", "esc", "ctrl+c"),
	key.WithHelp("", "press q to quit"),
)

var fetchKeys = key.NewBinding(
	key.WithKeys("f", "p"),
	key.WithHelp("", "press 'f' or 'p' to fetch a pokemon"),
)

func initialModel() model {
	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("205"))
	return model{spinner: s}
}

func (m model) Init() tea.Cmd {
	return m.spinner.Tick
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {

	case tea.KeyMsg:
		if key.Matches(msg, quitKeys) {
			m.quitting = true
			return m, tea.Quit

		}
		if key.Matches(msg, fetchKeys) {
			i := rand.Intn(151)
			if i <= 0 {
				i = rand.Intn(600)
			}
			p, err := pokeapi.Pokemon(strconv.Itoa(i))
			if err != nil {
				log.Fatal("there was an error retrieving a pokemon", err)
			}

			m.pokemon = p
		}
		return m, nil
	case errMsg:
		m.err = msg
		return m, nil

	default:
		var cmd tea.Cmd

		m.spinner, cmd = m.spinner.Update(msg)
		return m, cmd
	}
}

func (m model) View() string {
	if m.err != nil {
		return m.err.Error()
	}
	pokeStyle := lipgloss.DefaultRenderer().NewStyle().
		Background(lipgloss.Color("69")).
		Foreground(lipgloss.Color("#000")).
		Padding(2, 4).
		Width(40)

	pokelabel := "Fetch a pokemon!"
	if m.pokemon.Name != "" {
		pokelabel = "Got a pokemon: "
	}

	str := fmt.Sprintf(
		"%s\n\n   %s Waiting command...\n%s \n %s\n\n",
		pokeStyle.Render(pokelabel, m.pokemon.Name),
		m.spinner.View(),
		fetchKeys.Help().Desc,
		quitKeys.Help().Desc,
	)

	if m.quitting {
		return str + "\n"
	}
	return str
}

func main() {
	p := tea.NewProgram(initialModel())
	if _, err := p.Run(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
