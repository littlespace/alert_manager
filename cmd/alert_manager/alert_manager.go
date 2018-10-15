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
	"os"
	"strings"
)

var (
	pprofAddr   = flag.String("pprof-addr", "", "pprof address to listen on, dont activate pprof if empty")
	config      = flag.String("config", "", "Config file for alert_manager")
	fVersion    = flag.Bool("version", false, "display the version")
	nextVersion = "0.0.1"
	version     string
	commit      string
	branch      string
)

func init() {
	flag.Parse()
}

func getVersion() string {
	if version == "" {
		return fmt.Sprintf("v%s~%s", nextVersion, commit)
	}
	return "v" + version
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
				glog.Exitf("error Starting pprof: %v", err)
			}
		}()
	}
	if *fVersion {
		fmt.Printf("Alert Manager: %s , (git: %s, %s)\n", getVersion(), commit, branch)
		os.Exit(0)
	}
	if *config == "" {
		glog.Exit("A config file must be specified with -config")
	}
	glog.Infof("Starting Alert Manager %s", getVersion())
	config := alert_manager.NewConfig(*config)
	alert_manager.Run(config)
}
