package main

import (
	"flag"
	"fmt"
	"github.com/golang/glog"
	"github.com/mayuresh82/alert_manager"
	_ "github.com/mayuresh82/alert_manager/listener"
	_ "github.com/mayuresh82/alert_manager/listener/parsers"
	_ "github.com/mayuresh82/alert_manager/plugins/outputs"
	_ "github.com/mayuresh82/alert_manager/plugins/processors/all"
	_ "github.com/mayuresh82/alert_manager/plugins/transforms/all"
	"net/http"
	_ "net/http/pprof"
	"strings"
)

var (
	pprofAddr = flag.String("pprof-addr", "", "pprof address to listen on, dont activate pprof if empty")
	config    = flag.String("config", "config.toml", "Config file for alert_manager")
)

func init() {
	flag.Parse()
}

func main() {
	if *pprofAddr != "" {
		go func() {
			pprofHostPort := *pprofAddr
			parts := strings.Split(pprofHostPort, ":")
			if len(parts) == 2 && parts[0] == "" {
				pprofHostPort = fmt.Sprintf("localhost:%s", parts[1])
			}
			pprofHostPort = "http://" + pprofHostPort + "/debug/pprof"

			glog.Infof("Starting pprof HTTP server at: %s", pprofHostPort)

			if err := http.ListenAndServe(*pprofAddr, nil); err != nil {
				glog.Fatalf("error Starting pprof: %v", err)
			}
		}()
	}
	config := alert_manager.NewConfig(*config)
	alert_manager.Run(config)
}
