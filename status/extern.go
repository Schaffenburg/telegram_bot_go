package status

import (
	"encoding/json"
	"github.com/Schaffenburg/telegram_bot_go/config"
	"log"
	"net/http"
)

func init() {
	go acceptExternalStatus()
}

func acceptExternalStatus() {
	mux := http.NewServeMux()
	mux.HandleFunc("/setstatus", handleSetStatus)

	err := http.ListenAndServe(
		config.Get().SpaceStatusExtApi,
		mux,
	)

	log.Fatalf("Failed to Listen And Servie For Ext Spacestatus API: %s", err)
}

func handleSetStatus(rw http.ResponseWriter, r *http.Request) {
	conf := config.Get()
	rw.Header().Set("Content-Type", "application/json")

	if r.URL.Query().Get("key") != conf.SpaceStatusExtApiKey {
		rw.WriteHeader(405)
		rw.Write([]byte(`{"code":405,"error":"authentication required"}`))
		return
	}

	var status string
	dec := json.NewDecoder(r.Body)
	err := dec.Decode(&status)
	if err != nil {
		log.Printf("Failed to decode body: %s", err)
		rw.WriteHeader(400)
		rw.Write([]byte(`{"code":400,"error":"malformed requrest"}`))
		return
	}

	switch status {
	case "open":
		SetStatus(StatusOpen)

	case "close":
		SetStatus(StatusClosed)
	default:
		log.Printf("Failed to decode body: %s", err)
		rw.WriteHeader(405)
		rw.Write([]byte(`{"code":400,"error":"method not allowed"}`))
		return
	}

	rw.WriteHeader(200)
	rw.Write([]byte(`{"code":200}`))
}
