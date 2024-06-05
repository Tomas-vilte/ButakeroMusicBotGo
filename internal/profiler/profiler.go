package profiler

import (
	"github.com/grafana/pyroscope-go"
	"log"
)

// StartProfiler inicia el profiler de Pyroscope
func StartProfiler() {
	_, err := pyroscope.Start(pyroscope.Config{
		ApplicationName: "GoMusicBot",
		ServerAddress:   "http://pyroscope:4040",
		ProfileTypes: []pyroscope.ProfileType{
			pyroscope.ProfileCPU,
			pyroscope.ProfileAllocObjects,
			pyroscope.ProfileAllocSpace,
			pyroscope.ProfileInuseObjects,
			pyroscope.ProfileInuseSpace,
			pyroscope.ProfileGoroutines,
		},
	})
	if err != nil {
		log.Fatalf("Error al iniciar Pyroscope: %v", err)
	}
}
