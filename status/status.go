package status

import (
	tele "gopkg.in/tucnak/telebot.v2"

	"github.com/Schaffenburg/telegram_bot_go/cron"
	db "github.com/Schaffenburg/telegram_bot_go/database"
	"github.com/Schaffenburg/telegram_bot_go/help"
	loc "github.com/Schaffenburg/telegram_bot_go/localize"
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

const TagWantBeThere = "want_be_here"

const (
	CallbackAmHere      = "status_am_here"
	CallbackDelay15Mins = "status_stb_here"
	CallbackWontCome    = "status_wont_come"

	CallbackDepart = "status_depart"
	CallbackBRB    = "status_be_right_back"
	CallbackReturn = "status_return"
)

var (
	PermsEV = &nyu.PermissionFailText{
		Perm: &perms.PermissionGroupTag{"perm_ev"},

		Text: loc.MustTrans("perms.FailGroupEV"), //,
	}
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

	// message sent automagically
	bot.HandleInlineCallback(CallbackAmHere, handleArrivalCallback)
	bot.HandleInlineCallback(CallbackDelay15Mins, handleMoveArrivalCallback)
	bot.HandleInlineCallback(CallbackWontCome, handleWontComeCallback)

	bot.HandleInlineCallback(CallbackDepart, handleDepartCallback)
	bot.HandleInlineCallback(CallbackBRB, handleBRBCallback)
	bot.HandleInlineCallback(CallbackReturn, handleReturnCallback)

	bot.Command("heikomaas", handleSetArrival, PermsEV)
	bot.Command("eta", handleSetArrival, PermsEV)
	bot.Command("ichkommeheute", handleSetArrival, PermsEV)
	help.AddCommand("ichkommeheute")
	bot.Command("ichkommdochnicht", handleReviseArrival, PermsEV)
	help.AddCommand("ichkommdochnicht")

	bot.Command("werkommtheute", handleListArrival, PermsEV)
	help.AddCommand("werkommtheute")

	bot.Command("weristda", handleWhoThere, PermsEV)
	help.AddCommand("weristda")

	bot.Command("ichbinda", handleArrival, PermsEV)
	help.AddCommand("ichbinda")

	bot.Command("ichwaeregernda", handleWantArrival)
	help.AddCommand("ichwaeregernda")

	bot.Command("ichwaeredochnichtgernda", handleDontWantArrival)
	help.AddCommand("ichwaeredochnichtgernda")

	bot.Command("ichbinweg", handleDepart, PermsEV)
	help.AddCommand("ichbinweg")
	bot.Command("ichgehjetzt", handleDepart)
	help.AddCommand("ichgehjetzt")

	bot.Command("afk", handleBeRightBack, PermsEV) // alias
	help.AddCommand("afk")

	bot.Command("brb", handleBeRightBack, PermsEV)
	help.AddCommand("brb")

	bot.Command("wiederda", handleReturn, PermsEV)
	help.AddCommand("wiederda")

	bot.Command("forceclean", handleClean,
		&perms.PermissionGroupTag{GroupTag: "perm_ev"},
	)
	help.AddCommand("forceclean")

	bot.Command("forceevict",
		func(m *tele.Message) {
			bot := nyu.GetBot()

			everyoneDepart()
			bot.Send(m.Chat, "alle weg.")
		},
		PermsEV)
	help.AddCommand("forceevict")

	// Clean space at 4:00 o Clock
	cron.Daily(func() {
		i, err := db.CleanLocation()
		if err != nil {
			log.Printf("Error cleaning location: %s", err)
		} else {
			log.Printf("Cleaned %d location entries", i)
		}
	}, time.Hour*4)

	cron.Every(updateArrivalTimers, time.Minute*5)
}

func handleListArrival(m *tele.Message) {
	bot := nyu.GetBot()
	l := loc.GetUserLanguage(m.Sender)
	//	if member, _ := stalk.IsTaggedGroupMember(m.Sender.ID, "perm_ev"); !member {
	//		bot.Send(m.Chat, "Sorry, nur e.V. gruppen mitgliederis ist es erlaubt die liste an ankuendigungen zu lesen.")
	//		return
	//	}

	a, err := db.GetArrivals()
	if err != nil {
		bot.Send(m.Chat, FailGeneric.Getf(l, err))
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

	uas, err := ListUsersWithTagArrivingToday(TagHasKey)
	if uas == nil || len(uas) < 1 {
		b.WriteString("\n\n**Noch hat sich niemand mit schluessel bereit erklaert zu kommen**")
	}

	bot.Send(m.Chat, b.String())
}

func handleSetArrival(m *tele.Message) {
	bot := nyu.GetBot()
	l := loc.GetUserLanguage(m.Sender)
	// is allowed to set own arrival
	//	if member, _ := stalk.IsTaggedGroupMember(m.Sender.ID, "perm_ev"); !member {
	//		bot.Send(m.Chat, "Sorry, nur e.V. gruppen mitgliederis ist es erlaubt den Space zu betreten.\nFrage doch einfach mal, ob dich jemand enlaed.")
	//		return
	//	}

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

	// check for someone with keys
	haskey, err := db.UserHasTag(m.Sender.ID, TagHasKey)
	if err != nil {
		bot.Send(m.Chat, FailGeneric.Getf(l, err))

		return
	}

	if !haskey {
		uas, err := ListUsersWithTagArrivingToday(TagHasKey)
		if err != nil {
			log.Printf("Failed getting users with tag arriving today: %s", err)
		} else {
			var s *UserArrival

			for _, ua := range uas {
				if s == nil || ua.Arrival.Before(s.Arrival) {
					s = &ua
				}
			}

			if s == nil {
				bot.Send(m.Chat, "Sieht so als als haettest du keinen schluessel und es wolle niemand mit einem heute da sein")
			} else {
				user, err := stalk.GetUserByID(s.User)
				if err != nil {
					bot.Sendf(m.Chat, "Die Person die kommen will und gleichzeitig einen schluessel hat ist mir unbekannt %d", s.User)
				} else {
					bot.Sendf(m.Chat, "Fruehste person mit Schluessel ist %s um %s", user.FirstName, s.Arrival.Format("15:03"))
				}
			}
		}
	}

	err = db.SetArrival(m.Sender.ID, t.Unix())
	if err != nil {
		bot.Send(m.Chat, FailGeneric.Getf(l, err))
	} else {
		if t.Equal(util.Today(0)) {
			bot.Send(m.Chat, "Ok, bis dann!")
		} else {
			bot.Send(m.Chat, "Ok, bis um "+t.Format("15:04")+"!")
		}
	}
}

func updateArrivalTimers() {
	as, err := db.GetArrivals()
	if err != nil {
		log.Printf("Error getting arrivals: %s", err)
	}

	for _, a := range as {
		if time.Now().Unix() >= a.Time { // if time is in past
			AskUserIfArrived(a.User)
		}
	}
}

func handleClean(m *tele.Message) {
	bot := nyu.GetBot()
	//	if member, _ := stalk.IsTaggedGroupMember(m.Sender.ID, "perm_ev"); !member {
	//		bot.Send(m.Chat, "Sorry, nur e.V. gruppen mitgliederis ist erlaubt aufzuraeumen.")
	//		return
	//	}

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
	l := loc.GetUserLanguage(m.Sender)
	//	if member, _ := stalk.IsTaggedGroupMember(m.Sender.ID, "perm_ev"); !member {
	//		bot.Send(m.Chat, "Sorry, nur e.V. gruppen mitgliederis ist es erlaubt menschen im space zu ueberwachen.")
	//		return
	//	}

	list, err := db.WhoThere()
	if err != nil {
		bot.Send(m.Chat, FailGeneric.Getf(l, err))
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
			bot.Send(m.Chat, FailGeneric.Getf(l, err))
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

	// in gedanken da sind
	users, err := db.GetUsersWithTag(TagWantBeThere)
	if err == nil && len(users) > 0 {
		b.WriteString("\n\nIn Gedanken sind ausserdem da:")

		for _, u := range users {
			u, err := stalk.GetUserByID(u)
			if err != nil {
				bot.Send(m.Chat, FailGeneric.Getf(l, err))
				return
			}

			b.WriteString("\n - ")
			b.WriteString(u.FirstName)
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

	_, err = db.RmAllUserTag(TagWantBeThere)
	if err != nil {
		log.Printf("Error making everyone depart: %s", err)
	}

	i, err := r.RowsAffected()
	return i, err
}

func handleDepartCallback(c *tele.Callback) {
	bot := nyu.GetBot()
	l := loc.GetUserLanguage(c.Sender)

	ok, err := db.SetLocationDepart(c.Sender.ID)
	if err != nil {
		bot.RespondText(c, FailGeneric.Getf(l, err))
	}

	if !ok {
		bot.RespondText(c, "Wusste gar nicht, dass du da wars O.o, naja trotzdem noch einen schoenen Tag!")
	} else {
		bot.RespondText(c, "Sad to see you go, have a nice day :(")
	}
}

func handleBRBCallback(c *tele.Callback) {
	bot := nyu.GetBot()
	l := loc.GetUserLanguage(c.Sender)

	err := db.SetLocation(c.Sender.ID, time.Now().Unix(), "brb")
	if err != nil {
		bot.RespondText(c, FailGeneric.Getf(l, err))
	} else {
		bot.RespondText(c, "Ok, bis gleich!\n\nWieder da? bitte mit /wiederda bestaetigen :)")

		SendReminderReturn(c.Sender.ID)
	}
}

func handleReturnCallback(c *tele.Callback) {
	bot := nyu.GetBot()
	l := loc.GetUserLanguage(c.Sender)

	err := db.SetLocation(c.Sender.ID, time.Now().Unix(), "")
	if err != nil {
		bot.RespondText(c, FailGeneric.Getf(l, err))
	} else {
		bot.RespondText(c, "Schoen dass du (wieder) da bist, "+c.Sender.FirstName+"!")
	}
}

func handleDepart(m *tele.Message) {
	bot := nyu.GetBot()
	l := loc.GetUserLanguage(m.Sender)

	ok, err := db.SetLocationDepart(m.Sender.ID)
	if err != nil {
		bot.Send(m.Chat, FailGeneric.Getf(l, err))
	}

	if !ok {
		bot.Send(m.Chat, "Wusste gar nicht, dass du da wars O.o, naja trotzdem noch einen schoenen Tag!")
	} else {
		bot.Send(m.Chat, "Sad to see you go, have a nice day :(")
	}
}

func Arrive(u int64, note string) error {
	SendArrivalMessage(u)

	return db.SetLocation(u, time.Now().Unix(), note)
}

func handleArrivalCallback(m *tele.Callback) {
	bot := nyu.GetBot()
	l := loc.GetUserLanguage(m.Sender)

	err := Arrive(m.Sender.ID, "")
	if err != nil {
		bot.RespondText(m, FailGeneric.Getf(l, err))
	} else {
		bot.RespondText(m, "Hi, schoen, dass du da bist, "+m.Sender.FirstName+"!")
	}
}

func handleMoveArrivalCallback(m *tele.Callback) {
	bot := nyu.GetBot()
	l := loc.GetUserLanguage(m.Sender)

	ok, err := db.MoveArrival(m.Sender.ID, 60*15) // 60s * 15min
	if err != nil {
		bot.RespondText(m, FailGeneric.Getf(l, err))
		return
	}

	if ok {
		bot.RespondText(m, "Ok, ich informiere die anderen")
	} else {
		bot.RespondText(m, "Wusse nicht das du kommen wolltest o.O")
	}
}

func handleWontComeCallback(m *tele.Callback) {
	bot := nyu.GetBot()
	l := loc.GetUserLanguage(m.Sender)

	ok, err := db.RmArrival(m.Sender.ID) // 60s * 15min
	if err != nil {
		bot.RespondText(m, FailGeneric.Getf(l, err))
		return
	}

	if ok {
		bot.RespondText(m, "Ok, ich informiere die anderen")
	} else {
		bot.RespondText(m, "Wusse nicht das du kommen wolltest o.O")
	}
}

func handleArrival(m *tele.Message) {
	bot := nyu.GetBot()
	l := loc.GetUserLanguage(m.Sender)

	args := strings.Split(m.Text, " ")
	var note string

	if len(args) >= 2 {
		note = strings.Join(args[1:], " ")
	}

	err := Arrive(m.Sender.ID, note)
	if err != nil {
		bot.Send(m.Chat, FailGeneric.Getf(l, err))
	} else {
		bot.Send(m.Chat, "Hi, schoen, dass du da bist, "+m.Sender.FirstName+"!")
	}
}

func handleWantArrival(m *tele.Message) {
	bot := nyu.GetBot()
	l := loc.GetUserLanguage(m.Sender)

	err := db.SetUserTag(m.Sender.ID, TagWantBeThere)
	if err != nil {
		bot.Send(m.Chat, FailGeneric.Getf(l, err))
		return
	}

	bot.Sendf(m.Chat, "Du bist jetzt in gedanken dabei, %s!", m.Sender.FirstName)
}

func handleDontWantArrival(m *tele.Message) {
	bot := nyu.GetBot()
	l := loc.GetUserLanguage(m.Sender)

	changed, err := db.RmUserTag(m.Sender.ID, TagWantBeThere)
	if err != nil {
		bot.Send(m.Chat, FailGeneric.Getf(l, err))
		return
	}

	if changed {
		bot.Sendf(m.Chat, "Du bist nun nicht mehr in Gedanken dabei.")
	} else {
		bot.Sendf(m.Chat, "Du warst gar nicht da o.O")

	}
}

func handleBeRightBack(m *tele.Message) {
	bot := nyu.GetBot()
	l := loc.GetUserLanguage(m.Sender)

	err := db.SetLocation(m.Sender.ID, time.Now().Unix(), "brb")
	if err != nil {
		bot.Send(m.Chat, FailGeneric.Getf(l, err))
	} else {
		bot.Send(m.Chat, "Ok, bis gleich!\n\nWieder da? bitte mit /wiederda bestaetigen :)")
	}
}

func handleReturn(m *tele.Message) {
	bot := nyu.GetBot()
	l := loc.GetUserLanguage(m.Sender)

	err := db.SetLocation(m.Sender.ID, time.Now().Unix(), "")
	if err != nil {
		bot.Send(m.Chat, FailGeneric.Getf(l, err))
	} else {
		bot.Send(m.Chat, "Schoen dass du (wieder) da bist, "+m.Sender.FirstName+"!")
	}
}

func handleReviseArrival(m *tele.Message) {
	bot := nyu.GetBot()
	l := loc.GetUserLanguage(m.Sender)

	ch, err := db.RmArrival(m.Sender.ID)
	if err != nil {
		bot.Send(m.Chat, FailGeneric.Getf(l, err))
		return
	}

	if ch {
		bot.Send(m.Chat, "Okay, schade :(, dann halt naechstes mal.")
	} else {
		bot.Send(m.Chat, "Es gibt nichts zu revidieren. Viel Spaß bei was auch immer du so treibst!")
	}
}
