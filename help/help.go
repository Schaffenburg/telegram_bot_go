// help me echs dee
package help

import (
	loc "github.com/Schaffenburg/telegram_bot_go/localize"
	"github.com/Schaffenburg/telegram_bot_go/nyu"

	tele "gopkg.in/tucnak/telebot.v2"
	"log"
	"sort"
	"strings"
	"sync"
)

type Command struct {
	Text        string // command w/o slash
	Description loc.Translation
}

func init() {
	bot := nyu.GetBot()

	AddCommand("help")

	bot.Command("help", handleHelp)
	bot.Command("hilfe", handleHelp)
	bot.Command("hilfe!", handleHelp)
	bot.Command("?", handleHelp)

	nyu.OnRun(func() {
		helpEntriesMu.Lock()
		defer helpEntriesMu.Unlock()

		err := bot.SetCommands(teleHelpEntries(loc.GetLanguage(loc.DefaultLanguage)))
		if err != nil {
			log.Printf("Error setting commands.len(%d): %s", len(helpEntries), err)
		}
	})
}

// if l is nil the thingy is returned
func teleHelpEntries(l *loc.Language) (s []tele.Command) {
	s = make([]tele.Command, len(helpEntries))
	if l == nil {
		for k, v := range helpEntries {
			s[k] = tele.Command{
				Text:        v.Text,
				Description: v.Description.Name(),
			}

		}
	} else {
		for k, v := range helpEntries {
			s[k] = tele.Command{
				Text:        v.Text,
				Description: v.Description.Get(l),
			}
		}
	}

	return
}

func handleHelp(m *tele.Message) {
	bot := nyu.GetBot()
	l := loc.GetUserLanguage(m.Sender)

	bot.Send(m.Chat, "*Command List*:\n"+HelpText(l), tele.ModeMarkdown)
}

func AddCommand(c string) {
	helpEntriesMu.Lock()
	defer helpEntriesMu.Unlock()

	helpEntries = append(helpEntries, &Command{
		Text:        c,
		Description: loc.MustTrans("help." + c),
	})
}

type Commands []*Command

func (c Commands) Len() int           { return len(c) }
func (c Commands) Less(i, j int) bool { return c[i].Text < c[j].Text }
func (c Commands) Swap(i, j int)      { c[i], c[j] = c[j], c[i] }

var helpEntriesMu sync.Mutex
var helpEntries Commands

var helpText = make(map[int]string) // int is language index
var helpTextOnce sync.Once

func HelpText(l *loc.Language) string {
	helpTextOnce.Do(genHelpText)

	if l == nil {
		return helpText[-1]
	}

	return helpText[l.ID()]
}

func genHelpText() {
	helpEntriesMu.Lock()
	defer helpEntriesMu.Unlock()

	// sort help
	sort.Sort(helpEntries)

	b := &strings.Builder{}
	first := true
	for _, l := range loc.GetLanguages() {
		first = true
		for i := 0; i < len(helpEntries); i++ {
			if !first {
				b.WriteString("\n")
			}

			b.WriteString("/")
			b.WriteString(helpEntries[i].Text)
			b.WriteString(" - ")
			b.WriteString(helpEntries[i].Description.Get(l))

			first = false
		}

		helpText[l.ID()] = b.String()
		b.Reset()
	}

	first = true
	for i := 0; i < len(helpEntries); i++ {
		if !first {
			b.WriteString("\n")
		}

		b.WriteString("/")
		b.WriteString(helpEntries[i].Text)
		b.WriteString(" - ")
		b.WriteString(helpEntries[i].Description.Name())

		first = false
	}

	helpText[-1] = b.String()
}
