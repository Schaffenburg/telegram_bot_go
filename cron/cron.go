package cron

import (
	"log"
	"time"

	"git.schaffenburg.org/nyu/schaffenbot/util"
)

// tododo: cron stuffs

// executes f once every t
func Every(f func(), t time.Duration) {

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
