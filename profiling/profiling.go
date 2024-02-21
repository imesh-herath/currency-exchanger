package profiling

import (
	"assignment-imesh/configuration"
	"fmt"
	"log"
	"net/http"
)

func Profiling(appCfg configuration.AppConfig) {
	go func() {

		err := http.ListenAndServe(fmt.Sprintf(":%d", appCfg.Server.PprofPort), nil)
		if err != nil {
			log.Panicf("PProf server on %d cannot start!", appCfg.Server.PprofPort)
		}
	}()
}
