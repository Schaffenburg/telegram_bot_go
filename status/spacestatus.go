package status

import (
	tele "gopkg.in/tucnak/telebot.v2"

	db "github.com/Schaffenburg/telegram_bot_go/database"
	"github.com/Schaffenburg/telegram_bot_go/help"
	"github.com/Schaffenburg/telegram_bot_go/localize"
	"github.com/Schaffenburg/telegram_bot_go/nyu"
	"github.com/Schaffenburg/telegram_bot_go/perms"

	"log"
	"sync"
	"time"
)

const SpaceStatusSubTag = "status_info"
const DBGStatusTag = "dbg_status"

var (
	PermsSetStatus = &nyu.PermissionFailText{
		Perm: &perms.PermissionGroupTag{"perm_ev"},

		Text: loc.MustTrans("perms.FailSpaceStatusEV"),
	}
)

var (
	FailGeneric = loc.MustTrans("fail.generic")
)

func init() {
	err := db.StartDB()
	if err != nil {
		log.Fatalf("Failed to open DB: %s", err)
	}

	database := db.DB()

	_, err = database.Exec("CREATE TABLE IF NOT EXISTS `spacestatus` ( `time` BIGINT PRIMARY KEY, `status` TEXT );")
	if err != nil {
		log.Printf("error creating table arrivalTimes: %s", err)
		return
	}

	bot := nyu.GetBot()

	bot.Command("spaceopen", handleOpen, PermsSetStatus)
	bot.Command("openspace", handleOpen, PermsSetStatus)
	help.AddCommand("openspace")

	bot.Command("spaceclose", handleClose, PermsSetStatus)
	bot.Command("closespace", handleClose, PermsSetStatus)
	help.AddCommand("closespace")

	bot.Command("spacestatus", handleGetStatus)
	help.AddCommand("spacestatus")
	bot.Command("status", handleGetStatus, PermsEV)
	help.AddCommand("status")

	// subscription services:
	bot.Command("abonnieren", handleSubscribe)
	help.AddCommand("abonnieren")
	bot.Command("abobeenden", handleUnsubscribe)
	help.AddCommand("abobeenden")

	// super secret thingy
	bot.Command("dbg_abonnieren", handleSubscribeDBG, PermsEV)
	bot.Command("dbg_abobeenden", handleUnsubscribeDBG, PermsEV)
}

var (
	LSpaceChOpen  = loc.MustTrans("status.change.open")
	LSpaceChClose = loc.MustTrans("status.change.close")
)

func handleOpen(m *tele.Message) {
	l := loc.GetUserLanguage(m.Sender)

	bot := nyu.GetBot()

	err := SetStatus("open")
	if err != nil {
		bot.Send(m.Chat, FailGeneric.Getf(l, err))
	} else {
		bot.Send(m.Chat, LSpaceChOpen.Get(l))
	}
}

func handleClose(m *tele.Message) {
	l := loc.GetUserLanguage(m.Sender)

	bot := nyu.GetBot()

	err := SetStatus("closed")
	if err != nil {
		bot.Send(m.Chat, FailGeneric.Get(l)+err.Error())
	} else {
		bot.Send(m.Chat, LSpaceChClose.Get(l))
	}
}

// returns newest status entry closest to when
func GetStatus(when time.Time) (status SpaceStatus, err error) {
	r, err := db.StmtQuery(`SELECT status FROM spacestatus
	WHERE time < ?
	ORDER BY time DESC
	LIMIT 1`,
		when.Unix(),
	)

	if err != nil {
		return
	}

	if !r.Next() {
		return "<Nil>", nil
	}

	return status, r.Scan(&status)
}

type UserArrival struct {
	User    int64
	Arrival time.Time
}

// returns list of people who want to arrive today who have a Tag
func ListUsersWithTagArrivingToday(key string) (s []UserArrival, err error) {
	res, err := db.StmtQuery(`SELECT tags.user, at.time
	FROM tags as tags
	INNER JOIN arrivalTimes AS at ON tags.user = at.user
	WHERE tags.tag = ?;`, key)
	_ = res
	if err != nil {
		return nil, err
	}

	s = make([]UserArrival, 0)
	var user, arrival int64

	for res.Next() {
		err = res.Scan(&user, &arrival)
		if err != nil {
			return nil, err
		}

		s = append(s, UserArrival{
			User:    user,
			Arrival: time.Unix(arrival, 0),
		})
	}

	return s, nil
}

type SpaceStatus string

const (
	StatusOpen   SpaceStatus = "open"
	StatusClosed             = "closed"
)

var (
	LSpaceStatusOpen   = loc.MustTrans("status.spacestatus.open")
	LSpaceStatusClosed = loc.MustTrans("status.spacestatus.closed")
)

// TODO: localize
func (s SpaceStatus) Text(l *loc.Language) string {
	switch string(s) {
	case "open":
		return LSpaceStatusOpen.Get(l)
	case "closed":
		return LSpaceStatusClosed.Get(l)

	default:
		return "Space status: " + string(s)
	}
}

var (
	statusUpdateChsMu sync.RWMutex
	statusUpdateChs   []chan SpaceStatus
)

func AddStatusCh(ch chan SpaceStatus) {
	statusUpdateChsMu.Lock()
	defer statusUpdateChsMu.Unlock()

	statusUpdateChs = append(statusUpdateChs, ch)
}

func updateStatus(now time.Time, status SpaceStatus) {
	bot := nyu.GetBot()

	// send stuff (everyone with tag status_info gets the info)
	users, err := db.GetUsersWithTag(SpaceStatusSubTag)
	for _, u := range users {
		l := loc.MustGetUserLanguageID(u)

		bot.Send(&tele.User{ID: u}, status.Text(l))
	}
	if err != nil {
		log.Printf("Failed to broadcast spacestatus update: %s", err)
	}

	// broadcast to groups
	groups, err := db.GetTaggedGroups(SpaceStatusSubTag)
	for _, u := range groups {
		l := loc.DefaultLang()

		bot.Send(&tele.User{ID: u}, status.Text(&l))
	}
	if err != nil {
		log.Printf("Failed to broadcast spacestatus update: %s", err)
	}

	// broadcast to channels
	statusUpdateChsMu.RLock()
	for i := 0; i < len(statusUpdateChs); i++ {
		statusUpdateChs[i] <- status
	}
	statusUpdateChsMu.RUnlock()

	// boadcast to groups
}

func SetStatus(status SpaceStatus) error {
	now := time.Now()

	r, err := db.StmtExec(`INSERT INTO spacestatus (time, status) 
SELECT ?, ? 
FROM dual
WHERE NOT EXISTS (
    SELECT 1 
    FROM spacestatus 
    WHERE time = (SELECT MAX(time) FROM spacestatus) 
    AND status = ?
);
`,
		now.Unix(), status, status,
	)
	if err != nil {
		log.Printf("Failed to SetStatus: %s", err)
		return err
	}

	c, err := r.RowsAffected()
	if err != nil {
		log.Printf("Failed to SetStatus: %s", err)
	}

	if err != nil || c > 0 {
		updateStatus(now, status)
	}

	return err
}

func handleGetStatus(m *tele.Message) {
	l := loc.GetUserLanguage(m.Sender)
	bot := nyu.GetBot()

	status, err := GetStatus(time.Now())
	if err != nil {
		log.Printf("Failed to get status: %s", err)
		bot.Send(m.Chat, FailGeneric.Getf(l, err))
		return
	}

	bot.Sendf(m.Chat, "Space status: %s", status)
}

var (
	LSpaceStatusSubscribe = loc.MustTrans("status.subscribe.subscribe")
)

func handleSubscribeDBG(m *tele.Message) {
	l := loc.GetUserLanguage(m.Sender)
	bot := nyu.GetBot()

	err := db.SetUserTag(m.Sender.ID, DBGStatusTag)
	if err != nil {
		log.Printf("Error adding user tag: %s", err)
		bot.Send(m.Chat, FailGeneric.Getf(l, err))
	} else {
		bot.Send(m.Chat, LSpaceStatusSubscribe.Get(l))
	}
}

func handleSubscribe(m *tele.Message) {
	l := loc.GetUserLanguage(m.Sender)
	bot := nyu.GetBot()

	err := db.SetUserTag(m.Sender.ID, SpaceStatusSubTag)
	if err != nil {
		log.Printf("Error adding user tag: %s", err)
		bot.Send(m.Chat, FailGeneric.Getf(l, err))
	} else {
		bot.Send(m.Chat, LSpaceStatusSubscribe.Get(l))
	}
}

var (
	LSpaceStatusUnsubscribe          = loc.MustTrans("status.subscribe.unsubscribe")
	LSpaceStatusUnsubscribeNotChange = loc.MustTrans("status.subscribe.unsubscribe.nochange")
)

func handleUnsubscribe(m *tele.Message) {
	l := loc.GetUserLanguage(m.Sender)
	bot := nyu.GetBot()

	ch, err := db.RmUserTag(m.Sender.ID, SpaceStatusSubTag)
	if err != nil {
		log.Printf("Error removing user tag: %s", err)
		bot.Send(m.Chat, FailGeneric.Getf(l, err))
	} else {
		if ch {
			bot.Send(m.Chat, LSpaceStatusUnsubscribe.Get(l))
		} else {
			bot.Send(m.Chat, LSpaceStatusUnsubscribeNotChange.Get(l))
		}
	}
}

func handleUnsubscribeDBG(m *tele.Message) {
	l := loc.GetUserLanguage(m.Sender)
	bot := nyu.GetBot()

	ch, err := db.RmUserTag(m.Sender.ID, DBGStatusTag)
	if err != nil {
		log.Printf("Error removing user tag: %s", err)
		bot.Send(m.Chat, FailGeneric.Getf(l, err))
	} else {
		if ch {
			bot.Send(m.Chat, LSpaceStatusUnsubscribe.Get(l))
		} else {
			bot.Send(m.Chat, LSpaceStatusUnsubscribeNotChange.Get(l))
		}
	}
}
