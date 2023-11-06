package nyu

import (
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

	FailText() string
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

func handlePermit(f func(*tele.Message), perms ...Permission) func(*tele.Message) {
	return func(m *tele.Message) {
		bot := GetBot()

		var ok bool
		var err error

		for _, perm := range perms {
			ok, err = perm.Check(m)
			if err != nil {
				log.Printf("Permission check failed: %s", err)

				bot.Send(m.Chat, "Rechte ueberpruefung fehlgeschlagen!")

				return
			}

			if !ok {
				custom, ok := perm.(CustomPermission)
				if !ok {
					bot.Send(m.Chat, "Folgende Anforderung wird nicht erfuellt: "+perm.String())
					return
				}

				bot.Send(m.Chat, custom.FailText())

				return
			}
		}

		f(m)
	}
}
