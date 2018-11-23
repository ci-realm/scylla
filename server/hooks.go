package server

import (
	"net/http"

	"github.com/k0kubun/pp"
	"gopkg.in/go-playground/webhooks.v5/github"
)

func handleHooks() {
	hook, _ := github.New()

	http.HandleFunc("/hooks/github", func(w http.ResponseWriter, r *http.Request) {
		payload, err := hook.Parse(r, github.ReleaseEvent)
		if err != nil {
			if err == github.ErrEventNotFound {
				logger.Printf("Wrong event received: %s\n", err)
			}
		}
		pp.Println(payload)
	})
}
