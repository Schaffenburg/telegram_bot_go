// this file contains a bunch of messages that have inline buttons
package status

import (
	tele "gopkg.in/tucnak/telebot.v2"

	"github.com/Schaffenburg/telegram_bot_go/nyu"
)

func SendReminderReturn(u int64) {
	nyu.GetBot().Send(nyu.Recipient(u), "Wider da?",
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

func SendArrivalMessage(u int64) {
	nyu.GetBot().Send(nyu.Recipient(u), "Willkommen im space.\nHier ein paar Kurzwahltasten:",
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

// Sens the user arrival confirmation inline button thingy
func AskUserIfArrived(u int64) {
	nyu.GetBot().Send(nyu.Recipient(u), "Du wolltest jetzt da sein. bist du es auch?",
		&tele.SendOptions{
			ReplyMarkup: &tele.ReplyMarkup{
				InlineKeyboard: [][]tele.InlineButton{
					[]tele.InlineButton{
						tele.InlineButton{
							Unique: CallbackAmHere,

							Text: "Ja, i bims 1 da",
						},
					},
					[]tele.InlineButton{
						tele.InlineButton{
							Unique: CallbackDelay15Mins,

							Text: "In 15 minuten",
						},
						tele.InlineButton{
							Unique: CallbackWontCome,

							Text: "Donnich",
						},
					},
				},
			},
		},
	)
}
