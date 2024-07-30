package Helpers

import (
	"log"
	"net/http"
	_ "net/http/pprof"
)

// http://localhost:6060/debug/pprof/
func StartPprof() {
	go func() {
		log.Println(http.ListenAndServe("localhost:6060", nil))
	}()
}
