package internal

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
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
	quit := make(chan struct{})
	done := make(chan struct{})
	go flushJson(w, c, quit)

	err := callExiftool(c, done, filter)

	quit <- struct{}{}
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func flushJson(w http.ResponseWriter, c chan JsonTable, quit chan struct{}) {
	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "Streaming unsupported!", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/stream+json; charset=utf-8")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")

	var t JsonTable
	for {
		select {
		case <-quit:
			//fmt.Println("quit")
			return

		case t = <-c:
			err := json.NewEncoder(w).Encode(t)
			if err != nil {
				fmt.Println("json error: ", err)
				continue
			} else {
				// TODO: flush should also work with a list
				flusher.Flush()
			}
		}
	}
}

//func checkCancelStatus(r *http.Request, quit chan struct{}, canceled *bool) {
//	timer := time.NewTimer(time.Second)
//
//	for {
//		select {
//		case <-quit:
//			return
//		case <-r.Cancel:
//			*canceled = true
//			return
//		case <-timer.C:
//			if r.Close {
//				*canceled = true
//				fmt.Println("closed")
//			}
//		}
//	}
//}
