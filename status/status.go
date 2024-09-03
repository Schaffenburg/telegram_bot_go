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

	PermsVerein = &nyu.PermissionFailText{
		Perm: perms.MemberSpaceGroup,

		Text: loc.MustTrans("perms.FailAnyGroup"),
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
	LKeySomeoneNoTimeID           = loc.MustTrans("status.key.notime.someoneid")
	LKeySomeoneNoTime             = loc.MustTrans("status.key.notime.someone")
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

	bot.Command("heikomaas", handleSetArrival, PermsVerein)
	bot.Command("eta", handleSetArrival, PermsVerein)
	bot.Command("ichkommeheute", handleSetArrival, PermsVerein)
	bot.Command("ichkommeheut", handleSetArrival, PermsVerein)
	bot.Command("ichkommheut", handleSetArrival, PermsVerein)
	bot.Command("ichkommheute", handleSetArrival, PermsVerein)
	bot.Command("ichkomme", handleSetArrival, PermsVerein)
	bot.Command("ichkomm", handleSetArrival, PermsVerein)

	help.AddCommand("ichkommeheute")
	bot.Command("ichkommdochnicht", handleReviseArrival, PermsVerein)
	bot.Command("ichkommedochnicht", handleReviseArrival, PermsVerein)
	bot.Command("ichkommedochnich", handleReviseArrival, PermsVerein)
	bot.Command("ichkommdochnich", handleReviseArrival, PermsVerein)
	help.AddCommand("ichkommdochnicht")

	bot.Command("werkommtheute", handleListArrival, PermsVerein)
	bot.Command("werkommtheut", handleListArrival, PermsVerein)
	bot.Command("werkommheut", handleListArrival, PermsVerein)
	bot.Command("werkommheute", handleListArrival, PermsVerein)
	help.AddCommand("werkommtheute")

	bot.Command("weristda", handleWhoThere, PermsVerein)
	bot.Command("werisda", handleWhoThere, PermsVerein)
	help.AddCommand("weristda")

	bot.Command("ichbinda", handleArrival, PermsVerein)
	bot.Command("icame", handleArrival, PermsVerein)
	bot.Command("ichda", handleArrival, PermsVerein)
	bot.Command("binda", handleArrival, PermsVerein)
	help.AddCommand("ichbinda")

	bot.Command("ichwaeregernda", handleWantArrival)
	help.AddCommand("ichwaeregernda")

	bot.Command("ichwaeredochnichtgernda", handleDontWantArrival)
	help.AddCommand("ichwaeredochnichtgernda")

	bot.Command("ichbinweg", handleDepart, PermsVerein)
	bot.Command("binweg", handleDepart, PermsVerein)
	bot.Command("nixda", handleDepart, PermsVerein)
	bot.Command("ichverziehmich", handleDepart, PermsVerein)
	bot.Command("ichgehjetzt", handleDepart, PermsVerein)

	help.AddCommand("ichbinweg")
	help.AddCommand("ichgehjetzt")

	bot.Command("afk", handleBeRightBack, PermsVerein) // alias
	help.AddCommand("afk")

	bot.Command("brb", handleBeRightBack, PermsVerein)
	help.AddCommand("brb")

	bot.Command("wiederda", handleReturn, PermsVerein)
	bot.Command("ichbinwiederda", handleReturn, PermsVerein)
	help.AddCommand("wiederda")

	bot.Command("forceclean", handleClean, PermsEV)
	help.AddCommand("forceclean")

	bot.Command("forceevict",
		func(m *tele.Message) {
			bot := nyu.GetBot()
			l := loc.GetUserLanguage(m.Sender)

			forceOff()
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

		everyoneDepart()
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

		if util.IsUnknown(time.Unix(a[i].Time, 0)) {
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
		t = util.TodayUnknown()
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

	// dbg abo
	msg := fmt.Sprintf("dbg_abo: user %s %s (%d) wants to arrive @ %v (haskey: %t)",
		m.Sender.FirstName, m.Sender.LastName, m.Sender.ID,
		args, haskey,
	)

	dbgu, err := db.GetUsersWithTag(DBGStatusTag)
	if err != nil {
		log.Printf("failed to get users with tag %s: %s", DBGStatusTag, err)
	} else {
		for _, u := range dbgu {
			bot.Send(nyu.Recipient(u), msg)
		}
	}

	if !haskey {
		uas, err := ListUsersWithTagArrivingToday(TagHasKey)
		if err != nil {
			log.Printf("Failed getting users with tag arriving today: %s", err)
		} else {
			var unknowntime *UserArrival
			var knowntime *UserArrival

			for _, ua := range uas {
				if util.IsUnknown(ua.Arrival) {
					unknowntime = &ua
				}

				// earliest w/ key;; now hav one or is before now
				if knowntime == nil || ua.Arrival.Before(knowntime.Arrival) {
					knowntime = &ua
				}
			}

			if knowntime == nil {
				if unknowntime == nil {
					bot.Send(m.Chat, LKeyNoOne)
				} else {
					user, err := stalk.GetUserByID(knowntime.User)
					if err != nil {
						bot.Sendf(m.Chat, LKeySomeoneNoTimeID.Getf(l, knowntime.User))
					} else {
						bot.Sendf(m.Chat, LKeySomeoneNoTime.Getf(l, user.FirstName))
					}
				}
			} else {
				user, err := stalk.GetUserByID(knowntime.User)
				if err != nil {
					bot.Sendf(m.Chat, LKeySomeoneID.Getf(l, knowntime.User))
				} else {
					bot.Sendf(m.Chat, LKeySomeone.Getf(l, user.FirstName, knowntime.Arrival.Format("15:03")))
				}
			}
		}
	}

	err = db.SetArrival(m.Sender.ID, t.Unix())
	if err != nil {
		bot.Send(m.Chat, FailGeneric.Getf(l, err))
	} else {
		if util.IsUnknown(t) {
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
		t := time.Unix(a.Time, 0)

		if time.Now().Unix() >= a.Time &&
			time.Now().Add(-time.Minute*5).Unix() <= a.Time && // TODO: fix jank for single note
			(t.Second() == 0 || t.Minute() == 0 || t.Hour() == 0) { // if time is in past
			// check if user arrived:
			if there, _, _ := db.IsUserThere(a.User); !there {
				AskUserIfArrived(a.User)
			}
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

func forceOff() {
	everyoneDepart()

	SetStatus(StatusClosed)
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

func Depart(id int64) (ok bool, err error) {
	ok, err = db.SetLocationDepart(id)
	if err != nil {
		return
	}

	list, err := db.WhoThere()
	if err != nil {
		return
	}

	if len(list) <= 0 {
		// everyone gone
		SetStatus(StatusClosed)
	}

	return ok, err
}

func handleDepartCallback(c *tele.Callback) {
	bot := nyu.GetBot()
	l := loc.GetUserLanguage(c.Sender)

	ok, err := Depart(c.Sender.ID)
	if err != nil {
		bot.RespondText(c, FailGeneric.Getf(l, err))
	}

	// dbg abo
	msg := fmt.Sprintf("dbg_abo: user %s %s (%d) is gone",
		c.Sender.FirstName, c.Sender.LastName, c.Sender.ID,
	)

	dbgu, err := db.GetUsersWithTag(DBGStatusTag)
	if err != nil {
		log.Printf("failed to get users with tag %s: %s", DBGStatusTag, err)
	} else {
		for _, u := range dbgu {
			bot.Send(nyu.Recipient(u), msg)
		}
	}

	// delete arrival time
	db.RmArrival(c.Sender.ID)

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

	ok, err := Depart(m.Sender.ID)
	if err != nil {
		bot.Send(m.Chat, FailGeneric.Getf(l, err))
	}

	// dbg abo
	msg := fmt.Sprintf("dbg_abo: user %s %s (%d) is gone",
		m.Sender.FirstName, m.Sender.LastName, m.Sender.ID,
	)

	dbgu, err := db.GetUsersWithTag(DBGStatusTag)
	if err != nil {
		log.Printf("failed to get users with tag %s: %s", DBGStatusTag, err)
	} else {
		for _, u := range dbgu {
			bot.Send(nyu.Recipient(u), msg)
		}
	}

	// delete arrival time
	db.RmArrival(m.Sender.ID)

	if !ok {
		bot.Send(m.Chat, LDepartConfirmNochange.Get(l))
	} else {
		bot.Send(m.Chat, LDepartConfirm.Get(l))
	}
}

func Arrive(u int64, note string) error {
	SendArrivalMessage(u)
	SetStatus(StatusOpen)

	return db.SetLocation(u, time.Now().Unix(), note)
}

func handleArrivalCallback(m *tele.Callback) {
	bot := nyu.GetBot()
	l := loc.GetUserLanguage(m.Sender)

	// dbg abo
	msg := fmt.Sprintf("dbg_abo: user %s %s (%d) arrived",
		m.Sender.FirstName, m.Sender.LastName, m.Sender.ID,
	)

	dbgu, err := db.GetUsersWithTag(DBGStatusTag)
	if err != nil {
		log.Printf("failed to get users with tag %s: %s", DBGStatusTag, err)
	} else {
		for _, u := range dbgu {
			bot.Send(nyu.Recipient(u), msg)
		}
	}

	err = Arrive(m.Sender.ID, "")
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

	// dbg abo
	msg := fmt.Sprintf("dbg_abo: user %s %s (%d) arrived",
		m.Sender.FirstName, m.Sender.LastName, m.Sender.ID,
	)

	dbgu, err := db.GetUsersWithTag(DBGStatusTag)
	if err != nil {
		log.Printf("failed to get users with tag %s: %s", DBGStatusTag, err)
	} else {
		for _, u := range dbgu {
			bot.Send(nyu.Recipient(u), msg)
		}
	}

	err = Arrive(m.Sender.ID, note)
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

	bot.Sendf(m.Chat, LArrivethoughtConfirm.Getf(l, m.Sender.Username))
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
		bot.Send(m.Chat, LReturnConfirm.Getf(l, m.Sender.Username))
	}
}

func handleReviseArrival(m *tele.Message) {
	bot := nyu.GetBot()
	l := loc.GetUserLanguage(m.Sender)

	// dbg abo
	msg := fmt.Sprintf("dbg_abo: user %s %s (%d) doesnt want to arrive anymore",
		m.Sender.FirstName, m.Sender.LastName, m.Sender.ID,
	)

	dbgu, err := db.GetUsersWithTag(DBGStatusTag)
	if err != nil {
		log.Printf("failed to get users with tag %s: %s", DBGStatusTag, err)
	} else {
		for _, u := range dbgu {
			bot.Send(nyu.Recipient(u), msg)
		}
	}

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
