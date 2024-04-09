// this file contains a bunch of messages that have inline buttons
package status

import (
	tele "gopkg.in/tucnak/telebot.v2"

	"github.com/Schaffenburg/telegram_bot_go/localize"
	"github.com/Schaffenburg/telegram_bot_go/nyu"
)

var (
	LReturn = loc.MustTrans("status.return")
)

func SendReminderReturn(u int64) {
	l := loc.MustGetUserLanguageID(u)

	nyu.GetBot().Send(nyu.Recipient(u), LReturn.Get(l),
		&tele.SendOptions{
			ReplyMarkup: &tele.ReplyMarkup{
				InlineKeyboard: [][]tele.InlineButton{
					[]tele.InlineButton{
						tele.InlineButton{
							Unique: CallbackReturn,

							Text: "/wiederda",
						},
					},
				},
			},
		},
	)
}

var (
	LWelcomeSpeedDial = loc.MustTrans("status.welcomespeeddial")
)

func SendArrivalMessage(u int64) {
	l := loc.MustGetUserLanguageID(u)

	nyu.GetBot().Send(nyu.Recipient(u), LWelcomeSpeedDial.Get(l),
		&tele.SendOptions{
			ReplyMarkup: &tele.ReplyMarkup{
				InlineKeyboard: [][]tele.InlineButton{
					[]tele.InlineButton{
						tele.InlineButton{
							Unique: CallbackDepart,

							Text: "/ichgehjetzt",
						},
						tele.InlineButton{
							Unique: CallbackBRB,

							Text: "/brb",
						},
					},
				},
			},
		},
	)
}

var (
	LAreYouHereNow      = loc.MustTrans("status.areyouherenow")
	LAreYouHereNowYes   = loc.MustTrans("status.areyouherenow.yes")
	LAreYouHereNow15min = loc.MustTrans("status.areyouherenow.15m")
	LAreYouHereNowNo    = loc.MustTrans("status.areyouherenow.no")
)

// Sens the user arrival confirmation inline button thingy
func AskUserIfArrived(u int64) {
	l := loc.MustGetUserLanguageID(u)

	nyu.GetBot().Send(nyu.Recipient(u), LAreYouHereNow.Get(l),
		&tele.SendOptions{
			ReplyMarkup: &tele.ReplyMarkup{
				InlineKeyboard: [][]tele.InlineButton{
					[]tele.InlineButton{
						tele.InlineButton{
							Unique: CallbackAmHere,

							Text: LAreYouHereNowYes.Get(l),
						},
					},
					[]tele.InlineButton{
						tele.InlineButton{
							Unique: CallbackDelay15Mins,

							Text: LAreYouHereNow15min.Get(l),
						},
						tele.InlineButton{
							Unique: CallbackWontCome,

							Text: LAreYouHereNowNo.Get(l),
						},
					},
				},
			},
		},
	)
}
