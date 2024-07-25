package main

import (
	"fmt"
	"log"
	"math/rand"
	"os"
	"strconv"
	"strings"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/mtslzr/pokeapi-go"
	"github.com/mtslzr/pokeapi-go/structs"
)

type errMsg error
type listItem struct {
	name   string
	number int
}

type model struct {
	spinner  spinner.Model
	quitting bool
	err      error
	pokemon  structs.Pokemon
	bag      map[int]string
}

var quitKeys = key.NewBinding(
	key.WithKeys("q", "esc", "ctrl+c"),
	key.WithHelp("", "press q to quit"),
)

var fetchKeys = key.NewBinding(
	key.WithKeys("f", "p"),
	key.WithHelp("", "press 'f' or 'p' to fetch a pokemon"),
)

var listKeys = key.NewBinding(
	key.WithKeys("l"),
)

var (
	pokeCard = lipgloss.NewStyle().
		Border(lipgloss.NormalBorder()).
		Padding(2, 4).
		Width(40)
)

func initialModel() model {
	m := map[int]string{}
	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("205"))
	return model{spinner: s, bag: m}
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
			i := rand.Intn(151) + 1
			for m.bag[i] != "" {
				i = rand.Intn(151) + 1
			}
			p, err := pokeapi.Pokemon(strconv.Itoa(i))
			if err != nil {
				log.Fatal("there was an error retrieving a pokemon", err)
			}

			m.bag[p.Order] = p.Name
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

	pokelabel := "Fetch a pokemon!"
	var sb strings.Builder

	fmt.Fprint(&sb, pokeCard.Render(pokelabel, m.pokemon.Name))
	if m.pokemon.Name == "" {
		fmt.Fprintf(&sb, "\n\n%s\tWaiting for a command...", m.spinner.View())
	}
	if len(m.bag) > 0 {
		fmt.Fprintf(&sb, "\n%v", m.bag)
	}
	fmt.Fprintf(&sb, "\n%s", fetchKeys.Help().Desc)
	fmt.Fprintf(&sb, "\n%s", quitKeys.Help().Desc)

	// TODO: render sprite as an image
	if m.quitting {
		fmt.Fprint(&sb, "\nGood Bye")
	}
	return sb.String()
}

func main() {
	p := tea.NewProgram(initialModel())
	if _, err := p.Run(); err != nil {
		fmt.Println(err)
		log.Fatal(err)
		os.Exit(1)
	}
}
