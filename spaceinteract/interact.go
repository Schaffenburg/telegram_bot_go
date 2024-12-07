package interact

import (
	tele "gopkg.in/tucnak/telebot.v2"

	"github.com/Schaffenburg/telegram_bot_go/config"
	"github.com/Schaffenburg/telegram_bot_go/help"
	"github.com/Schaffenburg/telegram_bot_go/localize"
	"github.com/Schaffenburg/telegram_bot_go/nyu"
	"github.com/Schaffenburg/telegram_bot_go/perms"
	"github.com/Schaffenburg/telegram_bot_go/util"

	"encoding/json"
	"log"
	"math/rand"
	"net/http"
	"os"
	"time"
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

	bot.Command("heizungan", handleHeatingOn, PermsInteract)
	help.AddCommand("heizungan")
	bot.Command("heizungaus", handleHeatingOff, PermsInteract)
	help.AddCommand("heizungaus")

	bot.Command("wiewarmistes", handleGetTemperature, PermsInteract)
	help.AddCommand("wiewarmistes")
}

type heizungtype struct {
	Heizung    string
	HeizungDef struct {
		Heizung []string
	} `json:"_"`
}

func WriteHeizung(on bool) {
	log.Printf("writeHeizung(%v)", on)

	p := config.Get().HeizungJsonPath
	f, err := os.OpenFile(p, os.O_CREATE|os.O_RDWR|os.O_TRUNC, 0755)
	if err != nil {
		log.Printf("Error writing to '%s': %s", p, err)
		return
	}

	defer f.Close()

	onoff := map[bool][]string{
		true:  []string{"true", "on", "heizend"},
		false: []string{"false", "off", "unheizend"},
	}

	RSource := rand.New(rand.NewSource(time.Now().UnixNano() + 42))

	enc := json.NewEncoder(f)
	err = enc.Encode(&heizungtype{
		Heizung: util.OneOf(RSource, onoff[on]),
		HeizungDef: struct{ Heizung []string }{
			Heizung: append(onoff[true], onoff[false]...),
		},
	})

	if err != nil {
		log.Printf("Failed to encode heizung status: %s", err)
		return
	}
}

func handleRing(m *tele.Message) {
	bot := nyu.GetBot()
	l := loc.GetUserLanguage(m.Sender)

	resp, err := http.Get(config.Get().SpaceStatusGong)
	if err != nil {
		bot.Send(m.Chat, FailGeneric.Get(l)+": "+err.Error())
		log.Printf("Failed to gong im space: %s", err)
		return
	}

	log.Printf("Gong space got: %s", resp.Status)
	bot.Send(m.Chat, resp.Status)
}

func handleHeatingOn(m *tele.Message) {
	WriteHeizung(true)
	nyu.GetBot().Send(m.Chat, "Oki ^^")
}
func handleHeatingOff(m *tele.Message) {
	WriteHeizung(false)

	nyu.GetBot().Send(m.Chat, "Oki ^^")
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
