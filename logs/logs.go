package logs

import (
	"fmt"
	"log"
	"net/http"
	"time"
)

// https://ss64.com/nt/syntax-ansi.html
var (
	Reset  = "\033[0m"
	Red    = "\033[41m"
	Green  = "\033[42m"
	Yellow = "\033[103m"
	Blue   = "\033[44m"
	Purple = "\033[45m"
	Cyan   = "\033[46m"
	Gray   = "\033[100m"
	White  = "\033[107m"
)

// from https://dev.to/julienp/logging-the-status-code-of-a-http-handler-in-go-25aa
type StatusRecorder struct {
	http.ResponseWriter
	Status int
}

func (r *StatusRecorder) WriteHeader(status int) {
	r.Status = status
	r.ResponseWriter.WriteHeader(status)
}

func LoggerMiddleware(next http.Handler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		timeNow := time.Now()
		recorder := &StatusRecorder{
			ResponseWriter: w,
		}

		next.ServeHTTP(recorder, r)
		tolog := LogReqComposer(r.URL.Path, r.Host, r.Method, timeNow, recorder.Status)
		log.Print(tolog)
	}
}

// LogReqComposer
// A composer to create request logs with colors :) i'm blind so i need some colors on my logs
func LogReqComposer(url, host, method string, timetrigger time.Time, status int) string {
	//log.SetPrefix("[ entrance ] ")

	timeSpend := time.Since(timetrigger).Milliseconds()
	timeSpendStr := ""
	statusStr := ""
	switch {
	case status < 399:
		statusStr = fmt.Sprintf("%s %d %s", Green, status, Reset)
	case status >= 400 && status <= 499:
		statusStr = fmt.Sprintf("%s %d %s", Yellow, status, Reset)
	case status >= 500:
		statusStr = fmt.Sprintf("%s %d %s", Red, status, Reset)
	}
	switch {
	// Ill switch on the response time the server took to finish the request
	// For my services i expect
	// Query response time < 100
	// Mutations rsponse time  < 400
	// Auth response time < 1sec
	case timeSpend < 100:
		timeSpendStr = fmt.Sprintf("%s  %3d ms%s", Green, timeSpend, Reset)
	case timeSpend > 100 && timeSpend < 400:
		timeSpendStr = fmt.Sprintf("%s  %3d ms%s", Blue, timeSpend, Reset)
	case timeSpend > 400 && timeSpend < 600:
		timeSpendStr = fmt.Sprintf("%s  %3d ms%s", Yellow, timeSpend, Reset)
	case timeSpend > 600:
		timeSpendStr = fmt.Sprintf("%s  %3d ms%s", Red, timeSpend, Reset)
	}

	switch method {
	case "GET", "OPTIONS":
		method = fmt.Sprintf("%s  %s  %s", Green, method, Reset)
	case "POST":
		method = fmt.Sprintf("%s  %s  %s", Blue, method, Reset)
	case "PUT":
		method = fmt.Sprintf("%s  %s  %s", Yellow, method, Reset)
	case "DELETE":
		method = fmt.Sprintf("%s  %s  %s", Red, method, Reset)
	}

	tolog := fmt.Sprintf("| %3s |  %10s |\t%s  | %s |  \"%s\"", statusStr, timeSpendStr, host, method, url)
	return tolog

}
