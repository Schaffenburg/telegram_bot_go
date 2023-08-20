package status

import (
	tele "gopkg.in/tucnak/telebot.v2"

	"github.com/Schaffenburg/telegram_bot_go/cron"
	db "github.com/Schaffenburg/telegram_bot_go/database"
	"github.com/Schaffenburg/telegram_bot_go/help"
	"github.com/Schaffenburg/telegram_bot_go/nyu"
	"github.com/Schaffenburg/telegram_bot_go/perms"
	"github.com/Schaffenburg/telegram_bot_go/stalk"
	"github.com/Schaffenburg/telegram_bot_go/util"

	"fmt"
	"log"
	"strconv"
	"strings"
	"time"
)

func init() {
	err := db.StartDB()
	if err != nil {
		log.Fatalf("Failed to open DB: %s", err)
	}

	database := db.DB()

	_, err = database.Exec("CREATE TABLE IF NOT EXISTS `arrivalTimes` ( `user` BIGINT PRIMARY KEY, `time` BIGINT );")
	if err != nil {
		log.Printf("error creating table arrivalTimes: %s", err)
		return
	}

	_, err = database.Exec("CREATE TABLE IF NOT EXISTS `location` ( `user` BIGINT PRIMARY KEY, `since` BIGINT, `note` TEXT );")
	if err != nil {
		log.Printf("error creating table location: %s", err)
		return
	}

	bot := nyu.GetBot()

	bot.Command("ichkommeheute", handleSetArrival)
	help.AddCommand(tele.Command{
		Text:        "ichkommeheute",
		Description: "kündigt an, dass du heute im Space sein wirst.",
	})
	bot.Command("ichkommdochnicht", handleReviseArrival)
	help.AddCommand(tele.Command{
		Text:        "ichkommdochnicht",
		Description: "Revidiere deine Ankündung zu kommen.",
	})

	bot.Command("werkommtheute", handleListArrival)
	help.AddCommand(tele.Command{
		Text:        "werkommtheute",
		Description: "listet auf, wer heute im space sein will.",
	})

	bot.Command("weristda", handleWhoThere)
	help.AddCommand(tele.Command{
		Text:        "weristda",
		Description: "listet auf, wer grade im space ist.",
	})

	bot.Command("ichbinda", handleArrival)
	help.AddCommand(tele.Command{
		Text:        "ichbinda",
		Description: "bestaetigt, dass du im space bist.",
	})

	bot.Command("ichbinweg", handleDepart)
	help.AddCommand(tele.Command{
		Text:        "ichbinweg",
		Description: "bestaetigt, dass du den space verlassen hast.",
	})
	bot.Command("ichgehjetzt", handleDepart)
	help.AddCommand(tele.Command{
		Text:        "ichgehjetzt",
		Description: "bestaetigt, dass du den space verlassen hast.",
	})

	bot.Command("brb", handleBeRightBack)
	help.AddCommand(tele.Command{
		Text:        "brb",
		Description: "sag bescheid, dass du kurz weg bist.",
	})
	bot.Command("wiederda", handleAmRightBack)
	help.AddCommand(tele.Command{
		Text:        "wiederda",
		Description: "sag bescheid, dass wieder da bist.",
	})

	bot.Command("forceclean", handleClean,
		&perms.PermissionGroupTag{GroupTag: "perm_ev"},
	)
	help.AddCommand(tele.Command{
		Text:        "forceclean",
		Description: "raeumt die datenbank auf.",
	})

	bot.Command("forceevict",
		func(m *tele.Message) {
			bot := nyu.GetBot()
			if member, _ := stalk.IsTaggedGroupMember(m.Sender.ID, "perm_ev"); !member {
				bot.Send(m.Chat, "Sorry, nur e.V. gruppen mitgliederis ist erlaubt leute rauszuschmeissen.")
				return

			}

			everyoneDepart()
		},
		&perms.PermissionGroupTag{GroupTag: "perm_ev"},
	)
	help.AddCommand(tele.Command{
		Text:        "forceevict",
		Description: "schmeisst alle aus dem space.",
	})

	// Clean space at 4:00 o Clock
	cron.Daily(func() {
		i, err := db.CleanLocation()
		if err != nil {
			log.Printf("Error cleaning location: %s", err)
		} else {
			log.Printf("Cleaned %d location entries", i)
		}
	}, time.Hour*4)
}

func handleListArrival(m *tele.Message) {
	bot := nyu.GetBot()
	if member, _ := stalk.IsTaggedGroupMember(m.Sender.ID, "perm_ev"); !member {
		bot.Send(m.Chat, "Sorry, nur e.V. gruppen mitgliederis ist es erlaubt die liste an ankuendigungen zu lesen.")
		return
	}

	a, err := db.GetArrivals()
	if err != nil {
		bot.Send(m.Chat, "Ohno, ging nicht "+err.Error())
		return
	}

	if len(a) == 0 {
		bot.Send(m.Chat, "Heute will (noch) niemand da sein :(,\nJetzt kannst du noch als erstes da sein :D")
		return
	}

	b := &strings.Builder{}
	b.WriteString("Heute wollen noch kommen:")

	for i := 0; i < len(a); i++ {
		u, err := stalk.GetUserByID(a[i].User)
		if err != nil {
			panic(err)
		}

		if time.Unix(a[i].Time, 0).Equal(util.Today(0)) {
			fmt.Fprintf(b, "\n - %s irgendwann", u.FirstName)
			continue
		}

		fmt.Fprintf(b, "\n - %s um %s", u.FirstName, time.Unix(a[i].Time, 0).Format("15:04"))
	}

	bot.Send(m.Chat, b.String())
}

func handleSetArrival(m *tele.Message) {
	bot := nyu.GetBot()
	// is allowed to set own arrival
	if member, _ := stalk.IsTaggedGroupMember(m.Sender.ID, "perm_ev"); !member {
		bot.Send(m.Chat, "Sorry, nur e.V. gruppen mitgliederis ist es erlaubt den Space zu betreten.\nFrage doch einfach mal, ob dich jemand enlaed.")
		return
	}

	args := strings.Split(m.Text, " ")
	var t time.Time
	var err error

	if len(args) < 2 {
		t = util.Today(0)
	} else {
		t, err = util.ParseTime(args[1])
		if err != nil {
			bot.Send(m.Chat, "please speak more clearlier!\nValid formats: "+
				strings.Join(util.TimeFormats(), ", "))
			return
		}
	}

	err = db.SetArrival(m.Sender.ID, t.Unix())
	if err != nil {
		bot.Send(m.Chat, "Ohno, ging nicht "+err.Error())
	} else {
		if t.Equal(util.Today(0)) {
			bot.Send(m.Chat, "Ok, bis dann!")
		} else {
			bot.Send(m.Chat, "Ok, bis um "+t.Format("15:04")+"!")
		}
	}
}

func handleClean(m *tele.Message) {
	bot := nyu.GetBot()
	if member, _ := stalk.IsTaggedGroupMember(m.Sender.ID, "perm_ev"); !member {
		bot.Send(m.Chat, "Sorry, nur e.V. gruppen mitgliederis ist erlaubt aufzuraeumen.")
		return
	}

	streamer, err := bot.NewEditStreamer(m.Chat, "Cleaning...")
	if err != nil {
		log.Printf("handleClean: failed to initialize EditSteamer: %s", err)
		return
	}

	streamer.Append("\n arrivals...")

	rows, err := db.CleanArrivals()
	if err != nil {
		streamer.Append(" err: " + err.Error())
	} else {
		streamer.Append(" ok: " + strconv.FormatInt(rows, 10) + " rows")
	}

	streamer.Append("\n location...")

	rows, err = db.CleanLocation()
	if err != nil {
		streamer.Append("err: " + err.Error())
	} else {
		streamer.Append("ok: " + strconv.FormatInt(rows, 10) + " rows")
	}
}

func handleWhoThere(m *tele.Message) {
	bot := nyu.GetBot()
	if member, _ := stalk.IsTaggedGroupMember(m.Sender.ID, "perm_ev"); !member {
		bot.Send(m.Chat, "Sorry, nur e.V. gruppen mitgliederis ist es erlaubt menschen im space zu ueberwachen.")
		return
	}

	list, err := db.WhoThere()
	if err != nil {
		bot.Send(m.Chat, "Ohno, "+err.Error())
	}

	if len(list) < 1 {
		bot.Send(m.Chat, "Es ist grade niemand da :(")
		return
	}

	var b = &strings.Builder{}
	b.WriteString("Es sind grade da (")
	b.WriteString(strconv.FormatInt(int64(len(list)), 10))
	b.WriteString("): ")

	for _, s := range list {
		u, err := stalk.GetUserByID(s.ID)
		if err != nil {
			bot.Send(m.Chat, "Ohno, "+err.Error())
			return
		}

		b.WriteString("\n - ")
		b.WriteString(u.FirstName)

		if s.Note != "" {
			b.WriteString(" (")
			b.WriteString(s.Note)
			b.WriteString(")")
		}
	}

	bot.Send(m.Chat, b.String())
}

func everyoneDepart() (int64, error) {
	log.Printf("everyone departs now!")

	r, err := db.StmtExec("DELETE FROM location")
	if err != nil {
		log.Printf("Error making everyone depart: %s", err)
	}

	i, err := r.RowsAffected()
	return i, err
}

func handleDepart(m *tele.Message) {
	bot := nyu.GetBot()

	ok, err := db.SetLocationDepart(m.Sender.ID)
	if err != nil {
		bot.Send(m.Chat, "Ohno, ging nicht "+err.Error())
	}

	if !ok {
		bot.Send(m.Chat, "Wusste gar nicht, dass du da wars O.o, naja trotzdem noch einen schoenen Tag!")
	} else {
		bot.Send(m.Chat, "Sad to see you go, have a nice day :(")
	}
}

func handleArrival(m *tele.Message) {
	bot := nyu.GetBot()

	args := strings.Split(m.Text, " ")
	var note string

	if len(args) >= 2 {
		note = strings.Join(args[1:], " ")
	}

	err := db.SetLocation(m.Sender.ID, time.Now().Unix(), note)
	if err != nil {
		bot.Send(m.Chat, "Ohno, ging nicht "+err.Error())
	} else {
		bot.Send(m.Chat, "Hi, schoen, dass du da bist, "+m.Sender.FirstName+"!")
	}
}

func handleBeRightBack(m *tele.Message) {
	bot := nyu.GetBot()

	err := db.SetLocation(m.Sender.ID, time.Now().Unix(), "brb")
	if err != nil {
		bot.Send(m.Chat, "Ohno, ging nicht "+err.Error())
	} else {
		bot.Send(m.Chat, "Ok, bis gleich!\n\nWieder da? bitte mit /wiederda bestaetigen :)")
	}
}

func handleAmRightBack(m *tele.Message) {
	bot := nyu.GetBot()

	err := db.SetLocation(m.Sender.ID, time.Now().Unix(), "")
	if err != nil {
		bot.Send(m.Chat, "Ohno, ging nicht "+err.Error())
	} else {
		bot.Send(m.Chat, "Schoen dass du (wieder) da bist, "+m.Sender.FirstName+"!")
	}
}

func handleReviseArrival(m *tele.Message) {
	bot := nyu.GetBot()

	ch, err := db.RmArrival(m.Sender.ID)
	if err != nil {
		bot.Send(m.Chat, "Ohno, "+err.Error())
		return
	}

	if ch {
		bot.Send(m.Chat, "Okay, schade :(, dann halt naechstes mal.")
	} else {
		bot.Send(m.Chat, "Es gibt nichts zu revidieren. Viel Spaß bei was auch immer du so treibst!")
	}
}
