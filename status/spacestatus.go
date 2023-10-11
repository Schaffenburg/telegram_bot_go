package status

import (
	tele "gopkg.in/tucnak/telebot.v2"

	db "github.com/Schaffenburg/telegram_bot_go/database"
	"github.com/Schaffenburg/telegram_bot_go/help"
	"github.com/Schaffenburg/telegram_bot_go/nyu"
	"github.com/Schaffenburg/telegram_bot_go/perms"

	"log"
	"sync"
	"time"
)

const SpaceStatusSubTag = "status_info"

var (
	PermsSetStatus = &nyu.PermissionFailText{
		Perm: &perms.PermissionGroupTag{"perm_ev"},

		Text: "Um Befehle, die den Spacestatus setzen du benutzen, musst du Mitglied der e.V. Gruppe sein",
	}
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
	help.AddCommand(tele.Command{
		Text:        "openspace",
		Description: "oeffnet den space.",
	})

	bot.Command("spaceclose", handleClose, PermsSetStatus)
	bot.Command("closespace", handleClose, PermsSetStatus)
	help.AddCommand(tele.Command{
		Text:        "closespace",
		Description: "schliesst den space.",
	})

	bot.Command("spacestatus", handleGetStatus)
	help.AddCommand(tele.Command{
		Text:        "spacestatus",
		Description: "zeigt den status des spaces an.",
	})
	bot.Command("status", handleGetStatus, PermsEV)
	help.AddCommand(tele.Command{
		Text:        "status",
		Description: "zeigt den status des spaces an.",
	})

	// subscription services:
	bot.Command("abonnieren", handleSubscribe)
	help.AddCommand(tele.Command{
		Text:        "abonnieren",
		Description: "Abonniere den spacestatus als private push nachricht.",
	})
	bot.Command("abobeenden", handleUnsubscribe)
	help.AddCommand(tele.Command{
		Text:        "abobeenden",
		Description: "Kündige das Abonnement für private Benachrichtigungen über den Spacestatus.",
	})
}

func handleOpen(m *tele.Message) {
	bot := nyu.GetBot()
	// TODO: perms

	err := SetStatus("open")
	if err != nil {
		bot.Send(m.Chat, "Ohno, "+err.Error())
	} else {
		bot.Send(m.Chat, "Der space ist jetzt geoeffnet!")
	}
}

func handleClose(m *tele.Message) {
	bot := nyu.GetBot()
	// TODO: perms

	err := SetStatus("closed")
	if err != nil {
		bot.Send(m.Chat, "Ohno, "+err.Error())
	} else {
		bot.Send(m.Chat, "Der space ist jetzt geschlossen!")
	}
}

// returns newest status entry closest to when
func GetStatus(when time.Time) (status SpaceStatus, err error) {
	println("query")
	r, err := db.StmtQuery(`SELECT status FROM spacestatus
	WHERE time < ?
	ORDER BY time DESC
	LIMIT 1`,
		when.Unix(),
	)

	println("err != nil")
	if err != nil {
		return
	}

	if !r.Next() {
		return "<Nil>", nil
	}

	return status, r.Scan(&status)
}

type UserArrival struct {
	User int64
	When time.Time
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
			User: user,
			When: time.Unix(arrival, 0),
		})
	}

	return s, nil
}

type SpaceStatus string

func (s SpaceStatus) Text() string {
	switch string(s) {
	case "open":
		return "Der space ist jetzt geoeffnet!"
	case "closed":
		return "Der space ist jetzt geschlossen!"

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
		bot.Send(&tele.User{ID: u}, status.Text())
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

	updateStatus(now, status)

	_, err := db.StmtExec("INSERT INTO spacestatus (time, status) VALUES (?, ?);",
		now.Unix(), status,
	)

	return err
}

func handleGetStatus(m *tele.Message) {
	bot := nyu.GetBot()

	println("Get status now")
	status, err := GetStatus(time.Now())
	if err != nil {
		println("err")
		log.Printf("Failed to get status: %s", err)
		bot.Send(m.Chat, "Ohno, "+err.Error())
		return
	}

	println("send status")
	bot.Sendf(m.Chat, "Space status: %s", status)
}

func handleSubscribe(m *tele.Message) {
	bot := nyu.GetBot()

	err := db.SetUserTag(m.Sender.ID, SpaceStatusSubTag)
	if err != nil {
		log.Printf("Error adding user tag: %s", err)
		bot.Send(m.Chat, "Ohno, "+err.Error())
	} else {
		bot.Send(m.Chat, "Gern, du wirst jetzt ueber den space status auf dem laufenden gehalten.\nWenn du keine lust mehr auf diese Nachichten hast einfach mit /abobeenden dein abonnoment beenden :).")
	}
}

func handleUnsubscribe(m *tele.Message) {
	bot := nyu.GetBot()

	ch, err := db.RmUserTag(m.Sender.ID, SpaceStatusSubTag)
	if err != nil {
		log.Printf("Error adding user tag: %s", err)
		bot.Send(m.Chat, "Ohno, "+err.Error())
	} else {
		if ch {
			bot.Send(m.Chat, "Gern, du wirst jetzt nichtmehr ueber den space status informiert.\nWenn du wieder nachichten bekommen willst einfach mit /abonnieren dein abonnoment erneuern :).")
		} else {
			bot.Send(m.Chat, "o.O du warst gar nicht abonniert, naja jetzt bekommst du doppelt keine nachichten ;-).")
		}
	}
}
