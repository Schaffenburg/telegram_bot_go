package misc

import (
	owm "github.com/briandowns/openweathermap"
	tad "github.com/derzombiiie/timeanddate"
	gs "github.com/rocketlaunchr/google-search"
	tele "gopkg.in/tucnak/telebot.v2"

	"github.com/Schaffenburg/telegram_bot_go/config"
	db "github.com/Schaffenburg/telegram_bot_go/database"
	"github.com/Schaffenburg/telegram_bot_go/help"
	"github.com/Schaffenburg/telegram_bot_go/localize"
	"github.com/Schaffenburg/telegram_bot_go/nyu"
	"github.com/Schaffenburg/telegram_bot_go/perms"
	"github.com/Schaffenburg/telegram_bot_go/util"

	"context"
	"fmt"
	"log"
	"math/rand"
	"net/url"
	"strconv"
	"strings"
	"time"
)

var (
	PermsGroupEV = &nyu.PermissionFailText{
		Perm: perms.GroupEV,

		Text: loc.MustTrans("perms.FailGroupEV"), //"Um Befehle, die den Bot nachtichten schreiben lassen musst du Mitglied der e.V. Gruppe sein",
	}

	PermsGroupVorstand = &nyu.PermissionFailText{
		Perm: perms.GroupVorstand,

		Text: loc.MustTrans("perms.FailGroupVorstand"),
	}

	PermsDB = &nyu.PermissionFailText{
		Perm: perms.MemberSpaceGroup,

		Text: loc.MustTrans("perms.FailDB"),
	}

	PermsInternet = &nyu.PermissionFailText{
		Perm: perms.MemberSpaceGroup,

		Text: loc.MustTrans("perms.FailInternet"),
	}

	FailGeneric = loc.MustTrans("fail.generic")

	// good enough for a dice & ringedingens
	RSource = rand.New(rand.NewSource(time.Now().UnixNano() + 42))
)

func init() {

	bot := nyu.GetBot()

	//func (b *Bot) AnswerCommand(command, text string, perms ...Permission) {
	//bot.AnswerCommand("datum", "Es scheint so als gäbe es in Aschaffenburg kein Konzept für Zeitrechnung.")
	bot.Command("datum", handleGetTime)
	help.AddCommand("datum")

	bot.AnswerCommand("hallo", "Hallo @%h!")
	help.AddCommand("hallo")
	bot.AnswerCommand("hello", "Hello @%h!")
	help.AddCommand("hello")

	bot.AnswerCommand("lol", "rofl @%h hat lol gesagt!")
	help.AddCommand("lol")
	bot.AnswerCommand("rofl", "lol @%h hat rofl gesagt!")
	help.AddCommand("rofl")

	bot.AnswerCommand("mussichhaben", "*nicht!")
	help.AddCommand("mussichhaben")

	bot.Command("schleudern", handleSlap("@sender schleudert @argument ein bisschen herum mit einer großen Forelle"))
	help.AddCommand("schleudern")
	bot.Command("slap", handleSlap("@sender slaps @argument around a bit with a large trout"))
	help.AddCommand("slap")
	bot.Command("forelliere", handleSlap("@sender schlägt @argument eine große Forelle um die Ohren"))
	help.AddCommand("forelliere")
	bot.Command("batsche", handleSlap("@sender batscht @argument mithilfe eines Barsches"))
	help.AddCommand("batsche")

	bot.ReplyCommand("ping", "pong")
	help.AddCommand("ping")

	var (
		TheAnswerIs   = loc.MustTrans("misc.TheAnswerIs")
		TheAnswerIs42 = loc.MustTrans("misc.42")
	)

	bot.Command("wielautetdieantwort", func(m *tele.Message) {
		bot := nyu.GetBot()
		l := loc.GetUserLanguage(m.Sender)

		bot.Send(m.Chat, TheAnswerIs.Get(l))
		time.Sleep(time.Second * 3)

		bot.Send(m.Chat, TheAnswerIs42.Get(l), tele.ModeMarkdown)
	})
	help.AddCommand("wielautetdieantwort")

	bot.Command("echo", func(m *tele.Message) {
		if len(m.Text) > 5 {
			bot.Send(m.Chat, m.Text[5:])
		} else {
			bot.Send(m.Chat, "Usage: /echo <text>")
		}
	})
	help.AddCommand("echo")

	bot.Command("gidf", handleGoogle, PermsInternet)
	help.AddCommand("gidf")

	bot.Command("lmgt", handleLetMeXThat("lmgt",
		"Let Me Google That", "https://letmegooglethat.com/?q=")) // no need internet so no privs
	help.AddCommand("lmgt")

	bot.Command("lmgpt", handleLetMeXThat("lmgpt",
		"Let Me ChatGPT That", "https://letmegpt.com/?q=")) // no need internet so no privs
	help.AddCommand("lmgpt")

	bot.Command("wecker", handleTimer)
	help.AddCommand("wecker")
	bot.Command("werbinich", handleWhoAmI)
	help.AddCommand("werbinich")
	bot.Command("wuerfeln", handleDiceRoll)
	help.AddCommand("wuerfeln")
	bot.Command("featurerequest", handleAddFeatureRequest, PermsDB)
	help.AddCommand("featurerequest")
	bot.Command("wetter", handleWeather)
	help.AddCommand("wetter")

	bot.Command("laden", handleLoading)
	help.AddCommand("laden")

	bot.Command("cix", handleBroadcastCIX, PermsGroupVorstand)
	bot.Command("nyusletter", handleBroadcastCIX, PermsGroupVorstand)
}

func handleLoading(m *tele.Message) {
	bot := nyu.GetBot()

	args := strings.SplitN(m.Text, " ", 2)
	if len(args) != 2 {
		bot.Send(m.Chat, "Usage: /laden <umdrehungen>")

		return
	}

	uu, err := strconv.ParseUint(args[1], 10, 64)
	if err != nil {
		bot.Send(m.Chat, "Usage: /laden <umdrehungen>")

		return
	}

	const alphabet = "-\\|/-|/"

	go func() {
		bot := nyu.GetBot()

		msg, err := bot.Send(m.Chat, string(alphabet[len(alphabet)-1]))
		if err != nil {
			log.Printf("err :%s", err)

			return
		}

		t := time.NewTicker(time.Second)

		for i := uint64(0); i < uu; i++ {
			<-t.C
			msg, err = bot.Edit(msg, string(alphabet[i%uint64(len(alphabet))]))
			if err != nil {
				log.Printf("error editing: %s", err)

				return
			}
		}

		t.Stop()
	}()
}

var (
	NoConceptTime = loc.MustTrans("misc.noconcepttime")

	TimeAnswer = loc.MustTrans("misc.timeanswer")
)

func handleGetTime(m *tele.Message) {
	bot := nyu.GetBot()
	lang := loc.GetUserLanguage(m.Sender)

	args := strings.SplitN(m.Text, " ", 2)
	var query string

	if len(args) != 2 {
		query = config.Get().DefaultTimeLocation
	} else {
		query = args[1]

	}

	res, err := tad.Search(query)
	if err != nil {
		log.Printf("TAD: Failed to search '%s': %s", query, err)

		bot.Send(m.Chat, NoConceptTime.Getf(lang, query))
		return
	}

	if len(res) == 0 {
		bot.Send(m.Chat, NoConceptTime.Getf(lang, query))
		return
	}

	data, err := tad.Get(res[0].Path)
	if err != nil {
		log.Printf("TAD: Failed to get: %s", err)

		bot.Send(m.Chat, NoConceptTime.Getf(lang, query))
		return
	}

	bot.Sendf(m.Chat, TimeAnswer.Get(lang),
		res[0].City, data.Country,
		data.Time,
		data.Date.Day, data.Date.Month, data.Date.Year,
	)
}

var (
	BroadcastCIX = loc.MustTrans("misc.broadcastcix")
)

func handleBroadcastCIX(m *tele.Message) {
	bot := nyu.GetBot()
	lang := loc.GetUserLanguage(m.Sender)

	args := strings.SplitN(m.Text, " ", 2)
	if len(args) != 2 {
		bot.Send(m.Chat, "Usage: /nyusletter <text>")

		return
	}

	text := args[1]

	g, err := db.GetTaggedGroups("perm_cix") // get cix group
	if err != nil {
		log.Printf("Failed to determin cix group(s): %s", err)
		bot.Send(m.Chat, FailGeneric.Get(lang))

		return
	}

	if len(g) == 0 {
		log.Printf("Failed to determin cix group(s): no gropus")
		bot.Send(m.Chat, FailGeneric.Get(lang))

		return
	}

	bot.Sendf(m.Chat, BroadcastCIX.Get(lang), g)

	for i := 0; i < len(g); i++ {
		bot.Send(nyu.Recipient(g[i]), text)
	}
}

func handleGoogle(m *tele.Message) {
	bot := nyu.GetBot()
	l := loc.GetUserLanguage(m.Sender)

	var query string

	args := strings.SplitN(m.Text, " ", 2)
	if len(args) != 2 {
		// if this is a reply use text of message replied to
		if m.ReplyTo == nil || len(m.ReplyTo.Text) <= 0 {
			bot.Send(m.Chat, "Usage: /gidf <text>")
			return
		}

		query = m.ReplyTo.Text
	} else {
		query = args[1]
	}

	r, err := gs.Search(context.TODO(), query, gs.SearchOptions{
		CountryCode:  "de",
		LanguageCode: "de",

		Limit: 10,
	})
	if err != nil {
		log.Printf("Error searching google for '%s': %s", query, err)
		bot.Send(m.Chat, FailGeneric.Get(l)+err.Error())

		// doesnt return as the link and "Google ist dein Freund!" text is still better that just errors
		// return
	}

	log.Printf("Google results for '%s': %v", query, r)

	// no need to localize as just links german version anyways
	b := &strings.Builder{}
	b.WriteString("[Google](")
	b.WriteString("https://www.google.com/search?q=")
	b.WriteString(url.QueryEscape(query))
	b.WriteString(") ist dein [Freund](https://gidf.at)!")
	for _, l := range r {
		b.WriteString("\n - [")
		b.WriteString(l.Title)
		b.WriteString("](")
		b.WriteString(l.URL)
		b.WriteString(")\n    ")
		b.WriteString(l.Description)
	}

	if m.ReplyTo != nil {
		bot.Reply(m.ReplyTo, b.String(), &tele.SendOptions{
			ParseMode: tele.ModeMarkdown,
		})
	} else {
		bot.Send(m.Chat, b.String(), &tele.SendOptions{
			ParseMode: tele.ModeMarkdown,
		})
	}
}

func handleLetMeXThat(cmd, thing, qurl string) func(m *tele.Message) {
	return func(m *tele.Message) {
		bot := nyu.GetBot()

		var query string

		args := strings.SplitN(m.Text, " ", 2)
		if len(args) != 2 {
			// if this is a reply use text of message replied to
			if m.ReplyTo == nil || len(m.ReplyTo.Text) <= 0 {
				bot.Send(m.Chat, "Usage: /"+cmd+" <text>")
				return
			}

			query = m.ReplyTo.Text
		} else {
			query = args[1]
		}

		// no need to localise, as command is allready
		b := &strings.Builder{}
		b.WriteString("[")
		b.WriteString(thing)
		b.WriteString("](")
		b.WriteString(qurl)
		b.WriteString(url.QueryEscape(query))
		b.WriteString(") for you")

		if m.ReplyTo != nil {
			bot.Reply(m.ReplyTo, b.String(), &tele.SendOptions{
				ParseMode: tele.ModeMarkdown,
			})
		} else {
			bot.Send(m.Chat, b.String(), &tele.SendOptions{
				ParseMode: tele.ModeMarkdown,
			})
		}
	}
}

var (
	rings = []loc.Translation{
		loc.MustTrans("misc.ring1"),
		loc.MustTrans("misc.ring2"),
		loc.MustTrans("misc.ring3"),
		loc.MustTrans("misc.ring4"),
		loc.MustTrans("misc.ring5"),
	}

	Remind     = loc.MustTrans("misc.remind")
	WillRemind = loc.MustTrans("misc.WillRemind")
)

func handleTimer(m *tele.Message) {
	bot := nyu.GetBot()
	lang := loc.GetUserLanguage(m.Sender)

	args := strings.SplitN(m.Text, " ", 2)
	if len(args) != 2 {
		bot.Send(m.Chat, "Usage: /wecker <dauer>")

		return
	}

	d, err := time.ParseDuration(strings.ReplaceAll(args[1], " ", ""))
	if err != nil {
		bot.Send(m.Chat, FailGeneric.Get(lang)+err.Error())

		return
	}

	//_, offset := time.Now().Zone()
	msgtime := m.Time() //.Add(time.Second * time.Duration(offset))
	bot.Sendf(m.Chat, WillRemind.Get(lang), msgtime.Add(d).Format("15:04:05"))

	go func(dur time.Duration, chat *tele.Chat) {
		time.Sleep(msgtime.Add(d).Sub(time.Now())) // msg send + duration - now

		const Rings = 5

		for i := 0; i < Rings; i++ {
			bot.Send(chat, util.OneOf(RSource, rings).Get(lang))
			time.Sleep(time.Second / 2 / Rings) // send all in 1/2 second
		}

		bot.Send(chat, Remind.Get(lang))
	}(d, m.Chat)
}

var utctime, _ = time.LoadLocation("UTC")

var LWhoAmI = loc.MustTrans("misc.WhoAmI")

func handleWhoAmI(m *tele.Message) {
	bot := nyu.GetBot()
	lang := loc.GetUserLanguage(m.Sender)

	photo, err := bot.GetCurrentPFP(m.Sender)
	if err != nil {
		log.Printf("Failed to get current PFP of %d: %s", m.Sender.ID, err)
		bot.Sendf(m.Chat, FailGeneric.Get(lang), err)
		return
	}

	photo.Caption = fmt.Sprintf(LWhoAmI.Get(lang),
		m.Sender.ID, m.Sender.FirstName, m.Sender.LastName, m.Sender.Username,
	)

	bot.Send(m.Chat, photo)
}

var (
	LRollingDice   = loc.MustTrans("misc.rolling")
	LRollingResult = loc.MustTrans("misc.RollingResult")
)

func handleDiceRoll(m *tele.Message) {
	bot := nyu.GetBot()
	lang := loc.GetUserLanguage(m.Sender)

	rolling := LRollingDice.Get(lang)

	var Sides = 6
	// alt num of sides:
	args := strings.SplitN(m.Text, " ", 2)
	if len(args) == 2 {
		s, err := strconv.ParseInt(args[1], 10, 32)
		if err == nil && s >= 2 { // >= 2? -> otherwise \infty loop when generating intn in number() as != lastone
			Sides = int(s)
		}
	}

	lastnum := 0
	number := func() string {
		n := RSource.Intn(Sides) + 1
		for n == lastnum {
			n = RSource.Intn(Sides) + 1
		}

		lastnum = n

		return strconv.FormatInt(int64(lastnum), 10)
	}

	m, err := bot.Send(m.Chat, rolling+number())
	if err != nil {
		log.Printf("Failed to initialize Dice: %s", err)
		return
	}

	const RollTime = time.Second * 2
	const RollTimes = 10
	for i := 0; i < RollTimes; i++ {
		time.Sleep(RollTime / RollTimes)

		m, err = bot.Edit(m, rolling+number())
		if err != nil {
			log.Printf("Failed to roll dice: %s", err)
			return
		}
	}

	time.Sleep(RollTime / RollTimes)

	m, err = bot.Edit(m, fmt.Sprintf(LRollingResult.Get(lang), lastnum))
	if err != nil {
		log.Printf("failed to set final number of diceroll: %s", err)
	}
}

var (
	Ltemperature          = loc.MustTrans("misc.temperature")
	LFeatureRequestAccept = loc.MustTrans("misc.featurerequest.accept")
)

func handleAddFeatureRequest(m *tele.Message) {
	bot := nyu.GetBot()
	lang := loc.GetUserLanguage(m.Sender)

	args := strings.SplitN(m.Text, " ", 2)
	if len(args) != 2 {
		bot.Send(m.Chat, "Usage: /featurerequest <request>")
		return
	}

	err := db.AddLog(m.Sender.ID, time.Now(), "featurerequest", args[1])
	if err != nil {
		log.Printf("Failed to add featurerequest to log: %s", err)
		bot.Sendf(m.Chat, FailGeneric.Get(lang), err.Error())
		return
	}

	bot.Send(m.Chat, LFeatureRequestAccept.Get(lang))
}

var Lsling = loc.MustTrans("misc.sling")

func handleSlap(text string) func(m *tele.Message) {
	return func(m *tele.Message) {
		bot := nyu.GetBot()
		lang := loc.GetUserLanguage(m.Sender)

		args := strings.SplitN(m.Text, " ", 2)
		if len(args) != 2 {
			bot.Send(m.Chat, Lsling.Get(lang))

			return
		}

		// no real need to localise, as is only/mostly puns
		bot.Send(m.Chat, util.ReplaceMulti(map[string]string{
			"@sender":   "@" + m.Sender.Username,
			"@argument": "@" + strings.ReplaceAll(args[1], "@", ""),
		}, text))
	}
}

func handleWeather(m *tele.Message) {
	conf := config.Get()
	l := loc.GetUserLanguage(m.Sender)

	args := strings.Split(m.Text, " ")

	w, err := owm.NewCurrent("c", l.ISO(), conf.WeatherToken)
	if err != nil {
		log.Printf("Error creating current owm: %s", err)
		return
	}

	var location = conf.WeatherLocation

	if len(args) > 1 {
		location = args[1]
	}

	w.CurrentByName(location)

	s := &strings.Builder{}

	//TODO localize
	s.WriteString(Ltemperature.Getf(l, w.Name, w.Sys.Country, w.Main.Temp))

	doComma := false
	for i := 0; i < len(w.Weather); i++ {
		if doComma {
			s.WriteString(", ")
		}
		s.WriteString(w.Weather[i].Description)

		doComma = true
	}

	nyu.GetBot().Send(m.Chat, s.String())
}
