package helpers

import (
	"log"
	"net/http"
	_ "net/http/pprof"
	"runtime"
)

// http://localhost:6060/debug/pprof/
func StartPprof() {
	go func() {
		log.Println(http.ListenAndServe("localhost:6060", nil))
	}()
}

func HeapUsage() uint64 {
	var memStats runtime.MemStats
	runtime.ReadMemStats(&memStats)
	return memStats.HeapSys
}

func GoroutineCount() int {
	return runtime.NumGoroutine()
}
