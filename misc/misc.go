package misc

import (
	owm "github.com/briandowns/openweathermap"
	tad "github.com/derzombiiie/timeanddate"
	gs "github.com/rocketlaunchr/google-search"
	tele "gopkg.in/tucnak/telebot.v2"

	"github.com/Schaffenburg/telegram_bot_go/config"
	db "github.com/Schaffenburg/telegram_bot_go/database"
	"github.com/Schaffenburg/telegram_bot_go/help"
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

		Text: "Um Befehle, die den Bot nachtichten schreiben lassen musst du Mitglied der e.V. Gruppe sein",
	}

	PermsDB = &nyu.PermissionFailText{
		Perm: perms.MemberSpaceGroup,

		Text: "Um misc Befehle, die dinge in einer Datenbank speichern musst du Mitglied einer Spaceeigenen Gruppe sein",
	}

	PermsInternet = &nyu.PermissionFailText{
		Perm: perms.MemberSpaceGroup,

		Text: "Um misc Befehle, die mit dem Internet interaggieren, musst du mitglied in der e.V. oder CIX gruppe sein",
	}
)

func init() {

	bot := nyu.GetBot()

	//func (b *Bot) AnswerCommand(command, text string, perms ...Permission) {
	//bot.AnswerCommand("datum", "Es scheint so als gäbe es in Aschaffenburg kein Konzept für Zeitrechnung.")
	bot.Command("datum", handleGetTime)
	help.AddCommand(tele.Command{
		Text:        "datum",
		Description: "zeigt Uhrzeit/Datum abhängig vom Standort.",
	})
	bot.AnswerCommand("hallo", "Hallo @%h!")
	help.AddCommand(tele.Command{
		Text:        "hallo",
		Description: "sagt dem Bot hallo.",
	})
	bot.AnswerCommand("lol", "rofl @%h hat lol gesagt!")
	help.AddCommand(tele.Command{
		Text:        "lol",
		Description: "Einfach nur lol",
	})
	bot.AnswerCommand("rofl", "lol @%h hat rofl gesagt!")
	help.AddCommand(tele.Command{
		Text:        "rofl",
		Description: "Einfach rofl",
	})

	bot.AnswerCommand("hallo", "Hallo @%h!")
	help.AddCommand(tele.Command{
		Text:        "hallo",
		Description: "Sag dem Bot hallo",
	})
	bot.AnswerCommand("hello", "Hello @%h!")
	help.AddCommand(tele.Command{
		Text:        "hello",
		Description: "Sag dem Bot hallo",
	})

	bot.Command("schleudern", handleSlap("@sender schleudert @argument ein bisschen herum mit einer großen Forelle"))
	help.AddCommand(tele.Command{
		Text:        "schleudern",
		Description: "schleudert jemanden herum über seinen usernamen",
	})
	bot.Command("slap", handleSlap("@sender slaps @argument around a bit with a large trout"))
	help.AddCommand(tele.Command{
		Text:        "slap",
		Description: "slaps someone",
	})
	bot.Command("forelliere", handleSlap("@sender schlägt @argument eine große Forelle um die Ohren"))
	help.AddCommand(tele.Command{
		Text:        "forelliere",
		Description: "forelliert jemanden",
	})
	bot.Command("batsche", handleSlap(
		fmt.Sprintf("@sender batscht @argument mithilfe eines Barsches (staerke: %d/100)", config.Get().Batschigkeit)),
	)
	help.AddCommand(tele.Command{
		Text:        "batsche",
		Description: "batsche jemanden ein Barsch ins Gesicht",
	})

	bot.AnswerCommand("mussichhaben", "*nicht!")
	help.AddCommand(tele.Command{
		Text:        "mussichhaben",
		Description: "muss ich haben",
	})
	bot.ReplyCommand("ping", "pong")
	help.AddCommand(tele.Command{
		Text:        "ping",
		Description: "Einfach ping",
	})
	bot.Command("wielautetdieantwort", func(m *tele.Message) {
		bot.Send(m.Chat, "Die Antwort auf die endgültige Frage nach dem Leben, dem Universum und dem ganzen Rest lautet..")
		time.Sleep(time.Second * 3)

		bot.Send(m.Chat, "*42*!", tele.ModeMarkdown)
	})
	help.AddCommand(tele.Command{
		Text:        "wielautetdieantwort",
		Description: "Stellt DIE eine entgültige Frage, nach dem Leben, dem Universum und dem ganzen Rest.",
	})

	bot.Command("echo", func(m *tele.Message) {
		if len(m.Text) > 5 {
			bot.Send(m.Chat, m.Text[5:])
		} else {
			bot.Send(m.Chat, "Usage: /echo <text>")
		}
	})
	help.AddCommand(tele.Command{
		Text:        "echo",
		Description: "zeigt text an",
	})

	bot.Command("gidf", handleGoogle, PermsInternet)
	help.AddCommand(tele.Command{
		Text:        "gidf",
		Description: "[Google ist dein Freund](https://gidf.at)",
	})

	bot.Command("lmgt", handleLetMeXThat("lmgt",
		"Let Me Google That", "https://letmegooglethat.com/?q=")) // no need internet so no privs
	help.AddCommand(tele.Command{
		Text:        "lmgt",
		Description: "[Let Me Google That](https://letmegooglethat.com/) For you",
	})

	bot.Command("lmgpt", handleLetMeXThat("lmgpt",
		"Let Me ChatGPT That", "https://letmegpt.com/?q=")) // no need internet so no privs
	help.AddCommand(tele.Command{
		Text:        "lmgpt",
		Description: "[Let Me ChatGPT That](https://letmegpt.com/) For you",
	})

	bot.Command("wecker", handleTimer)
	help.AddCommand(tele.Command{
		Text:        "wecker",
		Description: "sagt nyu, dass du in X Zeit erinnert werden willst.",
	})
	bot.Command("werbinich", handleWhoAmI)
	help.AddCommand(tele.Command{
		Text:        "werbinich",
		Description: "zeigt deine ID, Name, Username und Profilbild an.",
	})
	bot.Command("wuerfeln", handleDiceRoll)
	help.AddCommand(tele.Command{
		Text:        "wuerfeln",
		Description: "würfelt.",
	})
	bot.Command("featurerequest", handleAddFeatureRequest, PermsDB)
	help.AddCommand(tele.Command{
		Text:        "featurerequest",
		Description: "ein neues Feature fuer den Bot anfragen.",
	})
	bot.Command("wetter", handleWeather)
	help.AddCommand(tele.Command{
		Text:        "wetter",
		Description: "zeigt das wetter im space.",
	})

	bot.Command("laden", handleLoading)
	help.AddCommand(tele.Command{
		Text:        "laden",
		Description: "laedt fuer x sekunden.",
	})

	bot.Command("cix", handleBroadcastCIX, PermsGroupEV)
	bot.Command("nyusletter", handleBroadcastCIX, PermsGroupEV)
	help.AddCommand(tele.Command{
		Text:        "nyusletter",
		Description: "Broadcast <text> to e.V. gruppe",
	})
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

func handleGetTime(m *tele.Message) {
	bot := nyu.GetBot()

	args := strings.SplitN(m.Text, " ", 2)
	var query string

	if len(args) != 2 {
		query = config.Get().DefaultTimeLocation
	} else {
		query = args[1]

	}

	var ErrText = "Es scheint so als gäbe es in " + query + " kein Konzept für Zeitrechnung."

	res, err := tad.Search(query)
	if err != nil {
		log.Printf("TAD: Failed to search '%s': %s", query, err)

		bot.Send(m.Chat, ErrText)
		return
	}

	if len(res) == 0 {
		bot.Send(m.Chat, ErrText)
		return
	}

	data, err := tad.Get(res[0].Path)
	if err != nil {
		log.Printf("TAD: Failed to get: %s", err)

		bot.Send(m.Chat, ErrText)
		return
	}

	bot.Sendf(m.Chat, "Die Zeit in %s (%s) betaegt grade %s\nEs ist der %d. %s %d",
		res[0].City, data.Country,
		data.Time,
		data.Date.Day, data.Date.Month, data.Date.Year,
	)
}

func handleBroadcastCIX(m *tele.Message) {
	bot := nyu.GetBot()

	args := strings.SplitN(m.Text, " ", 2)
	if len(args) != 2 {
		bot.Send(m.Chat, "Usage: /nyusletter <text>")

		return
	}

	text := args[1]

	g, err := db.GetTaggedGroups("perm_cix") // get ev group
	if err != nil {
		log.Printf("Failed to determin ev group: %s", err)
		bot.Send(m.Chat, "Ohno, es gab einen Fehler")

		return
	}

	if len(g) == 0 {
		log.Printf("Failed to determin ev group: no gropus")
		bot.Send(m.Chat, "Ohno, es gab einen Fehler")

		return
	}

	bot.Sendf(m.Chat, "Ok, broadcased to group(s): %v", g)

	for i := 0; i < len(g); i++ {
		bot.Send(nyu.Recipient(g[i]), text)
	}
}

func handleGoogle(m *tele.Message) {
	bot := nyu.GetBot()

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
		bot.Send(m.Chat, "Ohno, "+err.Error())

		// doesnt return as the link and "Google ist dein Freund!" text is still better that just errors
		// return
	}

	log.Printf("Google results for '%s': %v", query, r)

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

func handleTimer(m *tele.Message) {
	bot := nyu.GetBot()

	args := strings.SplitN(m.Text, " ", 2)
	if len(args) != 2 {
		bot.Send(m.Chat, "Usage: /wecker <dauer>")

		return
	}

	d, err := time.ParseDuration(strings.ReplaceAll(args[1], " ", ""))
	if err != nil {
		bot.Send(m.Chat, "Ohno, "+err.Error())

		return
	}

	//_, offset := time.Now().Zone()
	msgtime := m.Time() //.Add(time.Second * time.Duration(offset))
	bot.Send(m.Chat, "Ok, bis um "+msgtime.Add(d).Format("15:04:05")+"!")

	go func(dur time.Duration, chat *tele.Chat) {
		time.Sleep(msgtime.Add(d).Sub(time.Now())) // msg send + duration - now

		var Messages = []string{"Klingeling...", "Ringdingding...", "Ringkelingdeling...", "BEEP BEEP...", "Ringeding..."}
		const Rings = 5

		for i := 0; i < Rings; i++ {
			bot.Send(chat, util.OneOf(RSource, Messages))
			time.Sleep(time.Second / 2 / Rings) // send all in 1/2 second
		}

		bot.Send(chat, "Ich sollte dich an etwas erinnern")
	}(d, m.Chat)
}

var utctime, _ = time.LoadLocation("UTC")

func handleWhoAmI(m *tele.Message) {
	bot := nyu.GetBot()

	photo, err := bot.GetCurrentPFP(m.Sender)
	if err != nil {
		log.Printf("Faield to get current PFP of %d: %s", m.Sender.ID, err)
		bot.Send(m.Chat, "Ohno, "+err.Error())
		return
	}

	photo.Caption = fmt.Sprintf("Deine ID: %d\nName: %s %s\nUsername: %s",
		m.Sender.ID, m.Sender.FirstName, m.Sender.LastName, m.Sender.Username,
	)

	bot.Send(m.Chat, photo)
}

// good enough for a dice & ringedingens
var RSource = rand.New(rand.NewSource(time.Now().UnixNano() + 42))

func handleDiceRoll(m *tele.Message) {
	const Text = "Wuerfeln... "
	bot := nyu.GetBot()

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

	m, err := bot.Send(m.Chat, Text+number())
	if err != nil {
		log.Printf("Failed to initialize Dice: %s", err)
		return
	}

	const RollTime = time.Second * 2
	const RollTimes = 10
	for i := 0; i < RollTimes; i++ {
		time.Sleep(RollTime / RollTimes)

		m, err = bot.Edit(m, Text+number())
		if err != nil {
			log.Printf("Failed to roll dice: %s", err)
			return
		}
	}

	time.Sleep(RollTime / RollTimes)

	m, err = bot.Edit(m, "Habe eine "+strconv.FormatInt(int64(lastnum), 10)+" gewuerfelt!")
	if err != nil {
		log.Printf("failed to set final number of diceroll: %s", err)
	}
}

func handleAddFeatureRequest(m *tele.Message) {
	bot := nyu.GetBot()

	args := strings.SplitN(m.Text, " ", 2)
	if len(args) != 2 {
		bot.Send(m.Chat, "Usage: /featurerequest <request>")
		return
	}

	err := db.AddLog(m.Sender.ID, time.Now(), "featurerequest", args[1])
	if err != nil {
		log.Printf("Failed to add featurerequest to log: %s", err)
		bot.Send(m.Chat, "Ohno, "+err.Error())
		return
	}

	bot.Send(m.Chat, "Ok, schreib ich mir auf ^^")
}

func handleSlap(text string) func(m *tele.Message) {
	return func(m *tele.Message) {
		bot := nyu.GetBot()

		args := strings.SplitN(m.Text, " ", 2)
		if len(args) != 2 {
			bot.Send(m.Chat, "@"+m.Sender.Username+" tut mir Leid. Es gibt niemanden, den man rumschleudern könnte...")
			return
		}

		bot.Send(m.Chat, util.ReplaceMulti(map[string]string{
			"@sender":   "@" + m.Sender.Username,
			"@argument": "@" + strings.ReplaceAll(args[1], "@", ""),
		}, text))
	}
}

func handleWeather(m *tele.Message) {
	conf := config.Get()

	args := strings.Split(m.Text, " ")

	w, err := owm.NewCurrent("c", "de", conf.WeatherToken)
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

	s.WriteString("Die Temperatur in ")
	s.WriteString(w.Name)
	s.WriteString(" (")
	s.WriteString(w.Sys.Country)
	s.WriteString(") betraegt aktuell ")
	s.WriteString(strconv.FormatFloat(w.Main.Temp, 'g', 4, 64))
	s.WriteString("°C\nDie aktuellen Wetterbedingungen sind: ")

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
