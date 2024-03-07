package cron

import (
	"log"
	"time"

	"github.com/Schaffenburg/telegram_bot_go/util"
)

// executes f once every t ; uses timeTicker so is not aligned
func Every(f func(), t time.Duration) {
	go func() {
		last := time.NewTicker(t)

		for {
			<-last.C

			f()
		}
	}()
}

// executes f at hour:min o Clock
func Daily(f func(), daytime time.Duration) {
	go func() {
		for {
			now := time.Now()
			then := util.Today(daytime)

			if now.After(then) {
				then = then.AddDate(0, 0, 1) // next day
			}

			sl := then.Sub(now)
			log.Printf("sleeping %s\n", sl)
			time.Sleep(sl)

			f()
		}
	}()
}
