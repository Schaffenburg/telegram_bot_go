package interact

import (
	tele "gopkg.in/tucnak/telebot.v2"

	"github.com/Schaffenburg/telegram_bot_go/help"
	"github.com/Schaffenburg/telegram_bot_go/nyu"
	"github.com/Schaffenburg/telegram_bot_go/perms"
)

var (
	PermsInteract = &nyu.PermissionFailText{
		Perm: perms.MemberSpaceGroup,

		Text: "Um Befehle, die dinge im Space bewirken musst du Mitglied in der e.V. Gruppe sein",
	}
)

func init() {
	bot := nyu.GetBot()

	bot.Command("beep", handleRing, PermsInteract)
	bot.Command("gong", handleRing, PermsInteract)
	bot.Command("ring", handleRing, PermsInteract)
	help.AddCommand(tele.Command{
		Text:        "beep",
		Description: "Loese einen akkustischen Ton im Space aus.",
	})
	help.AddCommand(tele.Command{
		Text:        "gong",
		Description: "Manuell die Türklingel von Schaffenburg e.V. auslösen.",
	})
	help.AddCommand(tele.Command{
		Text:        "ring",
		Description: "Manuell die Türklingel von Schaffenburg e.V. auslösen.",
	})

	bot.Command("heitzungan", handleHeatingOn, PermsInteract)
	help.AddCommand(tele.Command{
		Text:        "heitzungan",
		Description: "Manuell die Heizung im Space anmachen.",
	})
	bot.Command("heizungaus", handleHeatingOff, PermsInteract)
	help.AddCommand(tele.Command{
		Text:        "heizungaus",
		Description: "Manuell die Heizung im Space ausmachen.",
	})

	bot.Command("wiewarmistes", handleGetTemperature, PermsInteract)
	help.AddCommand(tele.Command{
		Text:        "wiewarmistes",
		Description: "Zeigt die Temperatur im Hackspace Gebaeude.",
	})
}

func handleRing(m *tele.Message) {
	nyu.GetBot().Send(m.Chat, "TODO: actually do gong")
}

func handleHeatingOn(m *tele.Message) {
	nyu.GetBot().Send(m.Chat, "TODO: actually interact with heating")
}
func handleHeatingOff(m *tele.Message) {
	nyu.GetBot().Send(m.Chat, "TODO: actually interact with heating")
}

func handleGetTemperature(m *tele.Message) {
	nyu.GetBot().Send(m.Chat, "TODO: actually interact with sensor")

}
