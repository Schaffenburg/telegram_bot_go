package main

import (
	"github.com/Schaffenburg/telegram_bot_go/nyu"

	_ "github.com/Schaffenburg/telegram_bot_go/debug"
	_ "github.com/Schaffenburg/telegram_bot_go/help"
	_ "github.com/Schaffenburg/telegram_bot_go/localize"
	_ "github.com/Schaffenburg/telegram_bot_go/localize_cmd"
	_ "github.com/Schaffenburg/telegram_bot_go/misc"
	_ "github.com/Schaffenburg/telegram_bot_go/nyu"
	_ "github.com/Schaffenburg/telegram_bot_go/pager"
	_ "github.com/Schaffenburg/telegram_bot_go/spaceinteract"
	_ "github.com/Schaffenburg/telegram_bot_go/spacestatus"
	_ "github.com/Schaffenburg/telegram_bot_go/stalk"
	_ "github.com/Schaffenburg/telegram_bot_go/status"
)

func main() {
	// much code
	// do stuff
	// many do
	nyu.Run()
}
