package misc

import (
	owm "github.com/briandowns/openweathermap"
	gs "github.com/rocketlaunchr/google-search"
	tele "gopkg.in/tucnak/telebot.v2"

	"git.schaffenburg.org/nyu/schaffenbot/config"
	db "git.schaffenburg.org/nyu/schaffenbot/database"
	"git.schaffenburg.org/nyu/schaffenbot/help"
	"git.schaffenburg.org/nyu/schaffenbot/nyu"
	"git.schaffenburg.org/nyu/schaffenbot/util"

	"context"
	"fmt"
	"log"
	"math/rand"
	"net/url"
	"strconv"
	"strings"
	"time"
)

func init() {
	bot := nyu.Bot()

	bot.Handle("/datum", util.AnswerHandler(bot, "Es scheint so als gäbe es in Aschaffenburg kein Konzept für Zeitrechnung."))
	help.AddCommand(tele.Command{
		Text:        "datum",
		Description: "zeigt Uhrzeit/Datum abhängig vom Standort.",
	})
	bot.Handle("/hallo", util.AnswerHandler(bot, "Hallo @%h!"))
	help.AddCommand(tele.Command{
		Text:        "hallo",
		Description: "sagt dem Bot hallo.",
	})
	bot.Handle("/lol", util.AnswerHandler(bot, "rofl @%h hat lol gesagt!"))
	help.AddCommand(tele.Command{
		Text:        "lol",
		Description: "Einfach nur lol",
	})
	bot.Handle("/rofl", util.AnswerHandler(bot, "lol @%h hat rofl gesagt!"))
	help.AddCommand(tele.Command{
		Text:        "rofl",
		Description: "Einfach rofl",
	})

	bot.Handle("/schleudern", handleSlap("@sender schleudert @argument ein bisschen herum mit einer großen Forelle"))
	help.AddCommand(tele.Command{
		Text:        "schleudern",
		Description: "schleudert jemanden herum über seinen usernamen",
	})
	bot.Handle("/slap", handleSlap("@sender slaps @argument around a bit with a large trout"))
	help.AddCommand(tele.Command{
		Text:        "slap",
		Description: "slaps someone",
	})
	bot.Handle("/forelliere", handleSlap("@sender schlägt @argument eine große Forelle um die Ohren"))
	help.AddCommand(tele.Command{
		Text:        "forelliere",
		Description: "forelliert jemanden",
	})
	bot.Handle("/batsche", handleSlap("@sender batscht @argument mithilfe eines Barsches"))
	help.AddCommand(tele.Command{
		Text:        "batsche",
		Description: "batsche jemanden ein Barsch ins Gesicht",
	})

	bot.Handle("/mussichhaben", util.AnswerHandler(bot, "*nicht!"))
	help.AddCommand(tele.Command{
		Text:        "mussichhaben",
		Description: "muss ich haben",
	})
	bot.Handle("/ping", util.ReplyHandler(bot, "pong"))
	help.AddCommand(tele.Command{
		Text:        "ping",
		Description: "Einfach ping",
	})
	bot.Handle("/wielautetdieantwort", func(m *tele.Message) {
		bot.Send(m.Chat, "Die Antwort auf die endgültige Frage nach dem Leben, dem Universum und dem ganzen Rest lautet..")
		time.Sleep(time.Second * 3)

		bot.Send(m.Chat, "*42*!", tele.ModeMarkdown)
	})
	help.AddCommand(tele.Command{
		Text:        "wielautetdieantwort",
		Description: "Stellt DIE eine entgültige Frage, nach dem Leben, dem Universum und dem ganzen Rest.",
	})

	bot.Handle("/echo", func(m *tele.Message) {
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

	bot.Handle("/gidf", handleGoogle)
	help.AddCommand(tele.Command{
		Text:        "gidf",
		Description: "[Google ist dein Freund](https://gidf.at)",
	})

	bot.Handle("/wecker", handleTimer)
	help.AddCommand(tele.Command{
		Text:        "wecker",
		Description: "sagt nyu, dass du in X Zeit erinnert werden willst.",
	})
	bot.Handle("/werbinich", handleWhoAmI)
	help.AddCommand(tele.Command{
		Text:        "werbinich",
		Description: "zeigt deine ID, Name, Username und Profilbild an.",
	})
	bot.Handle("/wuerfeln", handleDiceRoll)
	help.AddCommand(tele.Command{
		Text:        "wuerfeln",
		Description: "würfelt.",
	})
	bot.Handle("/featurerequest", handleAddFeatureRequest)
	help.AddCommand(tele.Command{
		Text:        "featurerequest",
		Description: "ein neues Feature fuer den Bot anfragen.",
	})
	bot.Handle("/wetter", handleWeather)
	help.AddCommand(tele.Command{
		Text:        "wetter",
		Description: "zeigt das wetter im space.",
	})
}

func handleGoogle(m *tele.Message) {
	bot := nyu.Bot()

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

		Limit: 5,
	})
	if err != nil {
		log.Printf("Error searching google for '%s': %s", query, err)
		bot.Send(m.Chat, "Ohno, "+err.Error())

		// doesnt return as the link and "Google ist dein Freund!" text is still better that just errors
		// return
	}

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

func handleTimer(m *tele.Message) {
	bot := nyu.Bot()

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
	bot := nyu.Bot()

	photo, err := util.GetCurrentPFP(bot, m.Sender)
	if err != nil {
		log.Printf("Faield to get current PFP of %s: %s", m.Sender.ID, err)
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
	bot := nyu.Bot()

	var Sides = 6
	// alt num of sides:
	args := strings.SplitN(m.Text, " ", 2)
	if len(args) == 2 {
		s, err := strconv.ParseInt(args[1], 10, 32)
		if err == nil {
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
	bot := nyu.Bot()

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
		bot := nyu.Bot()

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

	w, err := owm.NewCurrent("c", "de", conf.WeatherToken)
	if err != nil {
		log.Printf("Error creating current owm: %s", err)
		return
	}

	w.CurrentByName(conf.WeatherLocation)

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

	nyu.Bot().Send(m.Chat, s.String())
}
