package interact

import (
	tele "gopkg.in/tucnak/telebot.v2"

	"github.com/Schaffenburg/telegram_bot_go/help"
	"github.com/Schaffenburg/telegram_bot_go/localize"
	"github.com/Schaffenburg/telegram_bot_go/nyu"
	"github.com/Schaffenburg/telegram_bot_go/perms"

	"log"
)

var (
	PermsInteract = &nyu.PermissionFailText{
		Perm: perms.MemberSpaceGroup,

		Text: loc.MustTrans("perms.FailDoSpaceEV"),
	}

	LTemperatureIs = loc.MustTrans("spaceinteract.temperatureis")
	FailGeneric    = loc.MustTrans("fail.generic")
)

func init() {
	bot := nyu.GetBot()

	bot.Command("beep", handleRing, PermsInteract)
	bot.Command("gong", handleRing, PermsInteract)
	bot.Command("ring", handleRing, PermsInteract)
	help.AddCommand("beep")
	help.AddCommand("gong")
	help.AddCommand("ring")

	bot.Command("heitzungan", handleHeatingOn, PermsInteract)
	help.AddCommand("heitzungan")
	bot.Command("heizungaus", handleHeatingOff, PermsInteract)
	help.AddCommand("heizungaus")

	bot.Command("wiewarmistes", handleGetTemperature, PermsInteract)
	help.AddCommand("wiewarmistes")
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
	l := loc.GetUserLanguage(m.Sender)

	temp, time, err := GetTemp()
	if err != nil {
		log.Printf("Failed to get temperature: %s", err)

		nyu.GetBot().Send(m.Chat, FailGeneric.Get(l))
	} else {
		timestr := time.Format("15:04 01.02.")

		nyu.GetBot().Send(m.Chat, LTemperatureIs.Getf(l, temp, timestr))
	}
}
