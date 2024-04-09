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

var (
	FailGeneric = loc.MustTrans("fail.generic")
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

var (
	LYourLanguageIs     = loc.MustTrans("localize.yourlanguageis")
	LYourLanguageIsAuto = loc.MustTrans("localize.yourlanguageis.auto")
)

func handleGetLang(m *tele.Message) {
	bot := nyu.GetBot()

	l, auto := loc.GetUserLanguageV(m.Sender)
	if auto {
		bot.Sendf(m.Chat, LYourLanguageIsAuto.Getf(l, l.Name()))
	} else {
		bot.Sendf(m.Chat, LYourLanguageIs.Getf(l, l.Name()))

	}
}

var (
	LSetLangageConfirm = loc.MustTrans("localize.setlanguage.confirm")
)

func handleSetLang(m *tele.Message) {
	bot := nyu.GetBot()
	l := loc.GetUserLanguage(m.Sender)

	args := strings.SplitN(m.Text, " ", 2)
	if len(args) != 2 {
		bot.Sendf(m.Chat, "Usage: /setlanguage <language>\nlanguage := %v or auto",
			loc.GetLanguagesS())

		return
	}

	if args[1] == "auto" {
		err := loc.SetUserLanguageAuto(m.Sender.ID)
		if err != nil {
			bot.Send(m.Chat, FailGeneric.Getf(l, err))
		} else {
			l = loc.GetUserLanguage(m.Sender)

			bot.Send(m.Chat, LSetLangageConfirm.Get(l))
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
	l = lang

	bot.Send(m.Chat, LSetLangageConfirm.Get(l))
}

func t[K any](c bool, a, b K) K {
	if c {
		return a
	}

	return b
}
