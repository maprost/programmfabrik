package internal

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"
)

type JsonTag struct {
	Type        string            `json:"type"`
	Writeable   bool              `json:"writeable"`
	Path        string            `json:"path"`
	Group       string            `json:"group"`
	Description map[string]string `json:"description"`
}

type JsonTable struct {
	Tags []JsonTag `json:"tags"`
}

func MainHandler(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path == "/tags" {
		runHandler(w, r, "")
	} else {
		filter, _ := strings.CutPrefix(r.URL.Path, "/")
		runHandler(w, r, filter)
	}

}

func runHandler(w http.ResponseWriter, r *http.Request, filter string) {
	c := make(chan JsonTable, 1)
	quitFlush := make(chan struct{})
	quitCheck := make(chan struct{})
	realQuit := make(chan struct{})
	canceled := false
	go flushJson(w, c, quitFlush, realQuit)
	go checkCancelStatus(r, quitCheck, &canceled)

	err := callExiftool(c, &canceled, filter)

	quitFlush <- struct{}{}
	quitCheck <- struct{}{}

	<-realQuit
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	fmt.Println("done")
}

func flushJson(w http.ResponseWriter, c chan JsonTable, quit chan struct{}, realQuit chan struct{}) {
	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "Streaming unsupported!", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/stream+json; charset=utf-8")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")

	fmt.Fprint(w, "[") // simulating a json list
	sendOne := false
	flush := func(t JsonTable) {
		if sendOne {
			fmt.Fprint(w, ",")
		} else {
			sendOne = true
		}
		err := json.NewEncoder(w).Encode(t)
		if err != nil {
			fmt.Println("json error: ", err)
			sendOne = false // so there are no two comma in line
		} else {
			flusher.Flush()
		}
	}

	var t JsonTable
	for {
		select {
		case <-quit:
			timer := time.NewTimer(1 * time.Second) // if quit channel is faster than c channel, wait a second before finish it.
			for {
				select {
				case <-timer.C:
					fmt.Fprint(w, "]")
					flusher.Flush()
					realQuit <- struct{}{}
					return

				case t = <-c:
					flush(t)
				}
			}

		case t = <-c:
			flush(t)
		}
	}
}

func checkCancelStatus(r *http.Request, quit chan struct{}, canceled *bool) {
	ctx := r.Context()

	for {
		select {
		case <-quit:
			return
		case <-ctx.Done():
			*canceled = true
			<-quit
			return
		}
	}
}
