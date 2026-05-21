package main

import (
	//"fmt"
	"net/http"
)

// 日志记录错误
func (app *application) logError(r *http.Request, err error) {
	var (
		method = r.Method
		uri    = r.URL.RequestURI()
	)

	app.logger.Error(err.Error(), "method", method, "uri", uri)
}

// 错误响应的格式
func (app *application) errorResponse(w http.ResponseWriter, r *http.Request, status int, message any) {
	env := envelope{"error": message}

	err := app.writeJSON(w, status, env, nil)
	if err != nil {
		app.logError(r, err)
		w.WriteHeader(status)
	}
}

// 处理未知的服务器错误
func (app *application) serverErrorResponse(w http.ResponseWriter, r *http.Request, err error) {
	app.logError(r, err)

	message := "the server encountered a problem and could not process your request"
	app.errorResponse(w, r, http.StatusInternalServerError, message)
}

// 未找到响应
func (app *application) notFoundResponse(w http.ResponseWriter, r *http.Request) {
	message := "the requested resource could not be found"
	app.errorResponse(w, r, http.StatusNotFound, message)
}

func (app *application) rateLimitExceededResponse(w http.ResponseWriter, r *http.Request) {
	message := "rate limit exceeded, please try again"
	app.errorResponse(w, r, http.StatusTooManyRequests, message)
}

// rateLimitExceededResponse 当用户请求过于频繁触发限流时，返回 429 Too Many Requests 状态码
func (app *application) ratelimitExceededResponse(w http.ResponseWriter, r *http.Request) {
	message := "rate limit exceeded, please try again"
	app.errorResponse(w, r, http.StatusTooManyRequests, message)
}
