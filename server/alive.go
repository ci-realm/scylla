package server

import "net/http"

func getAlive(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("ALIVE"))
}
