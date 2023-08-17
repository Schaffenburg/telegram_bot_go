// help me echs dee
package help

import (
	"github.com/Schaffenburg/telegram_bot_go/nyu"

	tele "gopkg.in/tucnak/telebot.v2"
	"log"
	"sort"
	"strings"
	"sync"
)

func init() {
	bot := nyu.Bot()

	AddCommand(tele.Command{
		Text:        "help",
		Description: "Zeigt eine list an befehlen an.",
	})

	bot.Handle("/help", handleHelp)
	bot.Handle("/hilfe", handleHelp)
	bot.Handle("/hilfe!", handleHelp)
	bot.Handle("/?", handleHelp)

	nyu.OnRun(func() {
		helpEntriesMu.Lock()
		defer helpEntriesMu.Unlock()

		err := bot.SetCommands(helpEntries)
		if err != nil {
			log.Printf("Error setting commands(%d): %s", len(helpEntries), err)
		}
	})
}

func handleHelp(m *tele.Message) {
	bot := nyu.Bot()

	bot.Send(m.Chat, "*Command List*:\n"+HelpText(), tele.ModeMarkdown)
}

func AddCommand(c tele.Command) {
	helpEntriesMu.Lock()
	defer helpEntriesMu.Unlock()

	helpEntries = append(helpEntries, c)
}

type Commands []tele.Command

func (c Commands) Len() int           { return len(c) }
func (c Commands) Less(i, j int) bool { return c[i].Text < c[j].Text }
func (c Commands) Swap(i, j int)      { c[i], c[j] = c[j], c[i] }

var helpEntriesMu sync.Mutex
var helpEntries Commands

var helpText string
var helpTextOnce sync.Once

func HelpText() string {
	helpTextOnce.Do(genHelpText)

	return helpText
}

func genHelpText() {
	helpEntriesMu.Lock()
	defer helpEntriesMu.Unlock()

	// sort help
	sort.Sort(helpEntries)

	b := &strings.Builder{}
	first := true
	for i := 0; i < len(helpEntries); i++ {
		if !first {
			b.WriteString("\n")
		}

		b.WriteString("/")
		b.WriteString(helpEntries[i].Text)
		b.WriteString(" - ")
		b.WriteString(helpEntries[i].Description)

		first = false
	}

	helpText = b.String()
}
