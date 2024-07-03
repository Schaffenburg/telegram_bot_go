package status

import (
	"github.com/Schaffenburg/telegram_bot_go/config"
	"github.com/Schaffenburg/telegram_bot_go/status"

	"bytes"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"net/url"
)

type SetStatusRequest struct {
	Name    string             `json:"name"`
	Status  status.SpaceStatus `json:"status"`
	Comment string             `json:"comment"`
	ApiKey  string             `json:"apikey"`
}

func init() {
	updateCh := make(chan status.SpaceStatus)
	status.AddStatusCh(updateCh)

	conf := config.Get()

	go func() {
		var st status.SpaceStatus

		for {
			st = <-updateCh

			// api.aschaffenburg.digital
			func() {
				buf := &bytes.Buffer{}

				enc := json.NewEncoder(buf)
				enc.Encode(&SetStatusRequest{
					Name:   conf.SpaceStatusAPIName,
					ApiKey: conf.SpaceStatusAPIKey,

					Comment: "",
					Status:  st,
				})

				//
				resp, err := http.Post("https://api.aschaffenburg.digital/status", "application/json", buf)
				if err != nil {
					log.Printf("Failed to push update to aschaffenburg.digital: %s")
					return
				}

				d, err := io.ReadAll(resp.Body)
				if err != nil {
					log.Printf("Failed to read body from aschaffenburg.digital: %s", err)
					return
				}

				log.Printf("set status at aschaffenburg.digital: %s", d)
			}()

			// legacy status (auch der auf der website glaube ich lul)
			func() {
				const apiendpoint = "https://status.schaffenburg.org/api.php?"

				var act string
				switch st {
				case status.StatusOpen:
					act = "open"
				case status.StatusClosed:
					act = "close"
				}

				if act == "" {
					// spacestatus no supported :/
					log.Printf("spacetatus %s; act %s; is not supported by legacy spacestatus", st, act)
					return
				}

				v := url.Values{
					"user":   []string{conf.SpaceStatusLegacyUser},
					"pwd":    []string{conf.SpaceStatusLegacyKey},
					"action": []string{act},
				}

				log.Printf("%s", apiendpoint+v.Encode())

				resp, err := http.Get(apiendpoint + v.Encode())
				if err != nil {
					log.Printf("Failed to update legacy spacestatus: %s", err)
					return
				}

				d, err := io.ReadAll(resp.Body)
				if err != nil {
					log.Printf("Failed to read body from legacy spacestatus: %s", err)
					return
				}

				log.Printf("set status at legacy spacestatus: %s", d)
			}()
		}
	}()
}
