package main

import (
	"net/http"

	"github.com/julienschmidt/httprouter" //处理option请求和对json统一友好
)

func (app *application) routes() http.Handler {

	router := httprouter.New()

	router.HandlerFunc(http.MethodGet, "/v1/healthcheck", app.healthcheckHandler)

	router.HandlerFunc(http.MethodGet, "/v1/transactions", app.listTransactionsHandler)

	return router
}
