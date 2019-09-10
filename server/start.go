package server


import (
	"fmt"
	"html/template"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
)


var templateFuncMap = []template.FuncMap{{
	"FormatTime": func(t time.Time) string {
		return t.Format("2006-01-02 15:04:05")
	},
	"ToClass": func(s string) string {
		return strings.ToLower(s)
	},
	"ShortSHA": func(s string) string {
		return s[0:7]
	},
	"FormatDuration": func(s time.Duration) string {
		return s.String()
	},
	"FormatTimeAgo": func(s time.Time) string {
		return time.Since(s).String()
	},
	"ScyllaVersionLink": func() string {
		return "https://source.xing.com/e-recruiting-api-team/scylla"
	},
	"ScyllaHostname": func() string {
		if host := os.Getenv("HOSTNAME"); host != "" {
			return host
		}
		return "localhost"
	},
}}

func Start() {
	ParseConfig()
	SetupDB()
	defer pgxpool.Close()

	go SetupQueue()

	go startLogDistributor(pgxpool)

	r := mux.NewRouter()
	setupRouting(r)

	recovery := handlers.RecoveryHandler(
		handlers.PrintRecoveryStack(true),
		handlers.RecoveryLogger(logger),
	)

  logger.Printf("Starting server at http://%s:%d\n", config.Host, config.Port)

	srv := &http.Server{
		Handler:      handlers.CombinedLoggingHandler(os.Stderr, recovery(r)),
		Addr:         fmt.Sprintf("%s:%d", config.Host, config.Port),
		WriteTimeout: 15 * time.Second,
		ReadTimeout:  15 * time.Second,
	}
	logger.Fatal(srv.ListenAndServe())
}
