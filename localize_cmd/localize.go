package misc

import (
	tele "gopkg.in/tucnak/telebot.v2"

	"github.com/Schaffenburg/telegram_bot_go/help"
	"github.com/Schaffenburg/telegram_bot_go/localize"
	"github.com/Schaffenburg/telegram_bot_go/nyu"

	"strings"
)

const (
	CallbackSetLanguage = "set_language"
)

func init() {
	bot := nyu.GetBot()

	bot.Command("setlanguage", handleSetLang)
	help.AddCommand("setlanguage")

	bot.Command("getlanguage", handleGetLang)
	help.AddCommand("getlanguage")

	// setlanguage callback bot.HandleInlineCallback(CallbackDTMF0, callbackKeyboard("0"))

	bot.HandleInlineCallback(CallbackSetLanguage+"_de", func(c *tele.Callback) {})
}
func handleGetLang(m *tele.Message) {
	bot := nyu.GetBot()

	l, auto := loc.GetUserLanguageV(m.Sender)
	bot.Sendf(m.Chat, "Deine Sprache ist %s%s", l.Name(), t(auto, " (auto)", ""))
}

func handleSetLang(m *tele.Message) {
	bot := nyu.GetBot()

	args := strings.SplitN(m.Text, " ", 2)
	if len(args) != 2 {
		bot.Sendf(m.Chat, "Usage: /setlanguage <language>\nlanguage := %v or auto",
			loc.GetLanguagesS())

		return
	}

	if args[1] == "auto" {
		err := loc.SetUserLanguageAuto(m.Sender.ID)
		if err != nil {
			bot.Send(m.Chat, "Ohno, "+err.Error())
		} else {
			bot.Send(m.Chat, "Ok, me remember that")
		}

		return
	}

	lang := loc.GetLanguage(args[1])
	if lang == nil {
		bot.Sendf(m.Chat, "Sprache nicht gefunden\nlanguage := %v or auto",
			loc.GetLanguagesS())

		return
	}

	loc.SetUserLanguage(m.Sender.ID, *lang)

	bot.Send(m.Chat, "Ok, me remember that")
}

func t[K any](c bool, a, b K) K {
	if c {
		return a
	}

	return b
}
