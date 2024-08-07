package main

import (
	"fmt"
	"log"
	"math/rand"
	"os"
	"strconv"
	"strings"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/mtslzr/pokeapi-go"
	"github.com/mtslzr/pokeapi-go/structs"
)

type errMsg error

type Item struct {
	title, desc string
}

func (i Item) Title() string {
	return i.title
}

func (i Item) Description() string {
	return i.desc
}

func (i Item) FilterValue() string {
	return i.title
}

var quitKeys = key.NewBinding(
	key.WithKeys("q", "esc", "ctrl+c"),
	key.WithHelp("", "press q to quit"),
)

var fetchKeys = key.NewBinding(
	key.WithKeys("f", "p"),
	key.WithHelp("", "press 'f' or 'p' to fetch a pokemon"),
)

var (
	pokeCard = lipgloss.NewStyle().
			Border(lipgloss.NormalBorder()).
			Padding(2, 4).
			Width(40)

	appStyle           = lipgloss.NewStyle().Padding(1, 2)
	statusMessageStyle = lipgloss.NewStyle().
				Foreground(lipgloss.AdaptiveColor{Light: "#04B575", Dark: "#ff66ff"}).
				Render
	// TODOs:
	// [ ] pokemon with name, summary, stats, 4 random moves, number, typing
)

func initialModel() model {
	i := []list.Item{}
	delegateKeys := newDelegateKeyMap()
	delegate := newItemDelegate(delegateKeys)
	l := list.New(i, delegate, 0, 0)
	l.Title = "PokeDex"
	b := make(map[int]structs.Pokemon)
	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("205"))

	return model{spinner: s, bag: b, list: l, delegateKeys: delegateKeys}
}

type model struct {
	spinner      spinner.Model
	quitting     bool
	err          error
	pokemon      structs.Pokemon
	list         list.Model
	bag          map[int]structs.Pokemon
	delegateKeys *delegateKeyMap
}

func (m model) Init() tea.Cmd {
	return m.spinner.Tick
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		h, v := appStyle.GetFrameSize()
		m.list.SetSize(msg.Width-h, msg.Height-v-15)
	case tea.KeyMsg:
		if m.list.FilterState() == list.Filtering {
			break
		}

		switch {
		case key.Matches(msg, quitKeys):
			m.quitting = true
			return m, tea.Quit
		case key.Matches(msg, fetchKeys):
			i := rand.Intn(151) + 1
			for _, ok := m.bag[i]; ok; _, ok = m.bag[i] {
				i = rand.Intn(151) + 1
			}
			p, err := pokeapi.Pokemon(strconv.Itoa(i))
			if err != nil {
				log.Fatal("there was an error retrieving a pokemon", err)
			}
			item := Item{title: p.Name, desc: fmt.Sprint(p.Order)}
			insertCmd := m.list.InsertItem(0, item)
			m.bag[p.Order] = p
			pokeDexSize := len(m.bag)
			m.pokemon = p
			title := fmt.Sprintf("%s\t%d pokemon caught!", m.list.Title, pokeDexSize)
			statusCmd := m.list.NewStatusMessage(statusMessageStyle(title))
			cmds = append(cmds, insertCmd, statusCmd)
		}
		// case key.Matches(msg, listKeys):
		var listCmd tea.Cmd
		m.list, listCmd = m.list.Update(msg)
		cmds = append(cmds, listCmd)
		return m, tea.Batch(cmds...)

	case errMsg:
		m.err = msg
		return m, nil

	default:
		// we are waiting people
		newSpinner, cmd := m.spinner.Update(msg)
		m.spinner = newSpinner
		return m, cmd
	}

	// update the list ery time
	newListModel, cmd := m.list.Update(msg)
	m.list = newListModel

	return m, cmd
}

func (m model) View() string {
	if m.err != nil {
		return appStyle.Render(m.err.Error())
	}

	pokelabel := "Fetch a pokemon!"
	var sb strings.Builder

	fmt.Fprint(&sb, pokeCard.Render(pokelabel, m.pokemon.Name))
	if m.pokemon.Name == "" {
		fmt.Fprintf(&sb, "\n\n%s\tWaiting for a command...", m.spinner.View())
	} else {
		fmt.Fprintf(&sb, "\n%s", m.list.View())
	}
	fmt.Fprintf(&sb, "\n%s", fetchKeys.Help().Desc)
	fmt.Fprintf(&sb, "\n%s", quitKeys.Help().Desc)

	if m.quitting {
		fmt.Fprintf(&sb, "\n\n")
		fmt.Fprint(&sb, statusMessageStyle("Good Bye!"))
	}

	return appStyle.Render(sb.String())
}

func main() {
	p := tea.NewProgram(initialModel())

	if _, err := p.Run(); err != nil {
		fmt.Println(err)
		log.Fatal(err)
		os.Exit(1)
	}
}
