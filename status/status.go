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

		Text: loc.MustTrans("perms.FailGroupEV"),
	}
)

var (
	LQueryNoone                   = loc.MustTrans("status.query.noone")
	LQueryList                    = loc.MustTrans("status.query.list")
	LSometime                     = loc.MustTrans("status.query.sometime")
	LQueryNoKey                   = loc.MustTrans("status.query.nokey")
	ParsetimeInvalid              = loc.MustTrans("status.parsetime.invalid")
	LKeyNoOne                     = loc.MustTrans("status.key.noone")
	LKeySomeoneID                 = loc.MustTrans("status.key.someoneid")
	LKeySomeone                   = loc.MustTrans("status.key.someone")
	LSetArrivalConfirm            = loc.MustTrans("status.setarrival.confirm")
	LSetArrivalConfirmTime        = loc.MustTrans("status.setarrival.confirm.time")
	LVisitingNoone                = loc.MustTrans("status.visiting.noone")
	LVisitingList                 = loc.MustTrans("status.visiting.list")
	LForceevictConfirm            = loc.MustTrans("status.forceevict.confirm")
	LDepartNochange               = loc.MustTrans("status.depart.nochange")
	LDepart                       = loc.MustTrans("status.depart")
	LBRBConfirm                   = loc.MustTrans("status.brb.confirm")
	LReturnConfirm                = loc.MustTrans("status.return.confirm")
	LDepartConfirmNochange        = loc.MustTrans("status.depart.confirm.nochange")
	LDepartConfirm                = loc.MustTrans("status.depart.confirm")
	LArriveConfirm                = loc.MustTrans("status.arrive.confirm")
	LMoveArrivalConfirm           = loc.MustTrans("status.movearrival.confirm")
	LMoveArrivalSchedule          = loc.MustTrans("status.movearrival.schedule")
	LRmArrivalConfirm             = loc.MustTrans("status.rmarrival.confirm")
	LRmArrivalConfirmNochange     = loc.MustTrans("status.rmarrival.confirm.nochange")
	LArrivethoughtConfirm         = loc.MustTrans("status.arrivethought.confirm")
	LDepartThoughtConfirm         = loc.MustTrans("status.departthought.confirm")
	LDepartThoughtConfirmNochange = loc.MustTrans("status.departthought.confirm.nochange")
	LReviseConfirm                = loc.MustTrans("status.revise.confirm")
	LReviseConfirmNochange        = loc.MustTrans("status.revise.confirm.nochange")
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
			l := loc.GetUserLanguage(m.Sender)

			everyoneDepart()
			bot.Send(m.Chat, LForceevictConfirm.Get(l))
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

	a, err := db.GetArrivals()
	if err != nil {
		bot.Send(m.Chat, FailGeneric.Getf(l, err))
		return
	}

	if len(a) == 0 {
		bot.Send(m.Chat, LQueryNoone.Get(l))
		return
	}

	b := &strings.Builder{}
	b.WriteString(LQueryList.Get(l))

	for i := 0; i < len(a); i++ {
		u, err := stalk.GetUserByID(a[i].User)
		if err != nil {
			panic(err)
		}

		if time.Unix(a[i].Time, 0).Equal(util.Today(0)) {
			// w/o time information
			fmt.Fprintf(b, "\n - %s %s", u.FirstName, LSometime.Get(l))
		} else {
			fmt.Fprintf(b, "\n - %s @ %s", u.FirstName, time.Unix(a[i].Time, 0).Format("15:04"))
		}
	}

	uas, err := ListUsersWithTagArrivingToday(TagHasKey)
	if uas == nil || len(uas) < 1 {
		b.WriteString("\n\n")
		b.WriteString(LQueryNoKey.Get(l))
	}

	bot.Send(m.Chat, b.String())
}

func handleSetArrival(m *tele.Message) {
	bot := nyu.GetBot()
	l := loc.GetUserLanguage(m.Sender)

	args := strings.Split(m.Text, " ")
	var t time.Time
	var err error

	if len(args) < 2 {
		t = util.Today(0)
	} else {
		t, err = util.ParseTime(args[1])
		if err != nil {
			bot.Send(m.Chat, ParsetimeInvalid.Get(l)+
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
				bot.Send(m.Chat, LKeyNoOne)
			} else {
				user, err := stalk.GetUserByID(s.User)
				if err != nil {
					bot.Sendf(m.Chat, LKeySomeoneID.Getf(l, s.User))
				} else {
					bot.Sendf(m.Chat, LKeySomeone.Getf(l, user.FirstName, s.Arrival.Format("15:03")))
				}
			}
		}
	}

	err = db.SetArrival(m.Sender.ID, t.Unix())
	if err != nil {
		bot.Send(m.Chat, FailGeneric.Getf(l, err))
	} else {
		if t.Equal(util.Today(0)) {
			bot.Send(m.Chat, LSetArrivalConfirm.Get(l))
		} else {
			bot.Send(m.Chat, LSetArrivalConfirmTime.Getf(l, t.Format("15:04")))
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

	list, err := db.WhoThere()
	if err != nil {
		bot.Send(m.Chat, FailGeneric.Getf(l, err))
	}

	if len(list) < 1 {
		bot.Send(m.Chat, LVisitingNoone.Get(l))
		return
	}

	var b = &strings.Builder{}
	b.WriteString(LVisitingList.Getf(l, len(list)))

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
		bot.RespondText(c, LDepartNochange.Get(l))
	} else {
		bot.RespondText(c, LDepart.Get(l))
	}
}

func handleBRBCallback(c *tele.Callback) {
	bot := nyu.GetBot()
	l := loc.GetUserLanguage(c.Sender)

	err := db.SetLocation(c.Sender.ID, time.Now().Unix(), "brb")
	if err != nil {
		bot.RespondText(c, FailGeneric.Getf(l, err))
	} else {
		bot.RespondText(c, LBRBConfirm.Get(l))

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
		bot.RespondText(c, LReturnConfirm.Getf(l, c.Sender.FirstName))
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
		bot.Send(m.Chat, LDepartConfirmNochange.Get(l))
	} else {
		bot.Send(m.Chat, LDepartConfirm.Get(l))
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
		bot.RespondText(m, LArriveConfirm.Getf(l, m.Sender.FirstName))
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
		bot.RespondText(m, LMoveArrivalConfirm.Get(l))
	} else {
		bot.RespondText(m, LMoveArrivalSchedule.Get(l))
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
		bot.RespondText(m, LRmArrivalConfirm.Get(l))
	} else {
		bot.RespondText(m, LRmArrivalConfirmNochange.Get(l))
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
		bot.Send(m.Chat, LArriveConfirm.Get(l))
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

	bot.Sendf(m.Chat, LArrivethoughtConfirm.Get(l))
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
		bot.Sendf(m.Chat, LDepartThoughtConfirm.Get(l))
	} else {
		bot.Sendf(m.Chat, LDepartThoughtConfirmNochange.Get(l))

	}
}

func handleBeRightBack(m *tele.Message) {
	bot := nyu.GetBot()
	l := loc.GetUserLanguage(m.Sender)

	err := db.SetLocation(m.Sender.ID, time.Now().Unix(), "brb")
	if err != nil {
		bot.Send(m.Chat, FailGeneric.Getf(l, err))
	} else {
		bot.Send(m.Chat, LBRBConfirm.Get(l))
	}
}

func handleReturn(m *tele.Message) {
	bot := nyu.GetBot()
	l := loc.GetUserLanguage(m.Sender)

	err := db.SetLocation(m.Sender.ID, time.Now().Unix(), "")
	if err != nil {
		bot.Send(m.Chat, FailGeneric.Getf(l, err))
	} else {
		bot.Send(m.Chat, LReturnConfirm.Get(l))
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
		bot.Send(m.Chat, LReviseConfirm.Get(l))
	} else {
		bot.Send(m.Chat, LReviseConfirmNochange.Get(l))
	}
}
