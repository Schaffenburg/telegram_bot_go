/*
this tries to keep an up2date version of all users
*/
package stalk

import (
	tele "gopkg.in/tucnak/telebot.v2"

	"git.schaffenburg.org/nyu/schaffenbot/nyu"
)

func init() {
	pp := nyu.Poller()

	ch := make(chan tele.Update, 100)
	pp.AddCh(ch)
	go func() {
		for {
			u, ok := <-ch
			if !ok {
				return
			}

			if u.Message != nil {
				stalkUsers(u.Message)
				stalkMemberships(u.Message)
			}
		}
	}()
}
