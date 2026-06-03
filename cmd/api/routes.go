package main

import (
	"embed"
	"net/http"

	"github.com/julienschmidt/httprouter" //处理option请求和对json统一友好
)

//go:embed ui/index.html
var fs embed.FS

func (app *application) routes() http.Handler {

	router := httprouter.New()

	router.NotFound = http.HandlerFunc(app.notFoundResponse)

	router.HandlerFunc(http.MethodGet, "/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		htmlBytes, err := fs.ReadFile("ui/index.html")
		if err != nil {
			app.logger.Error("failed to read embedded index.html", "error", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		w.Write(htmlBytes)
	})

	router.HandlerFunc(http.MethodGet, "/v1/healthcheck", app.healthcheckHandler)

	router.HandlerFunc(http.MethodGet, "/v1/transactions", app.listTransactionsHandler)

	router.HandlerFunc(http.MethodGet, "/v1/events", app.broker.Handler)

	return app.recoverPanic(router)
}
