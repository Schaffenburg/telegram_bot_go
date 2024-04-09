package nyu

import (
	"github.com/Schaffenburg/telegram_bot_go/config"
	"github.com/Schaffenburg/telegram_bot_go/localize"

	tele "gopkg.in/tucnak/telebot.v2"

	"log"
)

// see perms for implementations
type Permission interface {
	Check(*tele.Message) (bool, error)

	String() string
}

// CustomPermission extens the Permission interface by a ErrorText function for custom text on Check returning false
type CustomPermission interface {
	Permission

	FailText(u *tele.User) string
}

// Wraps a Permission to include a ErrorText function returning Text
type PermissionFailText struct {
	Perm Permission

	Text loc.Translation
}

func (p *PermissionFailText) FailText(u *tele.User) string {
	return p.Text.Get(loc.GetUserLanguage(u))
}

func (p *PermissionFailText) Check(m *tele.Message) (bool, error) { return p.Perm.Check(m) }
func (p *PermissionFailText) String() string                      { return p.Perm.String() }

var (
	FailFollowing = loc.MustTrans("perms.FailFollowing")
	FailGeneric   = loc.MustTrans("perms.FailGeneric")
)

func handlePermit(f func(*tele.Message), perms ...Permission) func(*tele.Message) {
	return func(m *tele.Message) {
		bot := GetBot()
		lang := loc.GetUserLanguage(m.Sender)

		// admin super powers!1!!
		conf := config.Get()
		if conf.SetupAdmin == m.Sender.ID {
			f(m) // permit

			return
		}

		var ok bool
		var err error

		for _, perm := range perms {
			ok, err = perm.Check(m)
			if err != nil {
				log.Printf("Permission check failed: %s", err)

				bot.Send(m.Chat, FailGeneric.Get(lang))

				return
			}

			if !ok {
				custom, ok := perm.(CustomPermission)
				if !ok {
					bot.Send(m.Chat, FailFollowing.Get(lang)+perm.String())
				} else {
					bot.Send(m.Chat, custom.FailText(m.Sender))
				}

				return
			}
		}

		f(m)
	}
}
