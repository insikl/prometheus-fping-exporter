package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	flags "github.com/jessevdk/go-flags"
	"github.com/prometheus/client_golang/prometheus/promhttp"

	"github.com/insikl/prometheus-fping-exporter/internal/logger"
)

var opts struct {
	Listen         string `short:"l" long:"listen" description:"Listen address" value-name:"[HOST]:PORT" default:":9605"`
	Period         uint   `short:"p" long:"period" description:"Period in seconds, should match Prometheus scrape interval" value-name:"SECS" default:"60"`
	Fping          string `short:"f" long:"fping"  description:"Fping binary path" value-name:"PATH" default:"/usr/bin/fping"`
	Count          uint   `short:"c" long:"count"  description:"Number of pings to send at each period" value-name:"N" default:"20"`
	StaleThreshold uint   `short:"s" long:"stale-threshold" description:"Stale target threshold in seconds" value-name:"SECS" default:"300"`
	Debug          bool   `long:"debug" description:"Enable debug logging"`
	Version        bool   `long:"version" description:"Show version"`
}

// Build information.
const (
	BuildVersion = "0.2.0"
)

// Build information populated at build-time.
var (
	BuildName   string
	BuildCommit string
	BuildBranch string
	BuildUser   string
	BuildDate   string
	BuildGo     string
	BuildOs     string
	BuildArch   string
)

func probeHandler(w http.ResponseWriter, r *http.Request) {
	targetParam := r.URL.Query().Get("target")
	if targetParam == "" {
		w.Header().Set("Content-Type", "text/html")
		_, _ = w.Write([]byte(`<html>
		    <head><title>Fping Exporter</title></head>
			<body>
			<b>ERROR: missing target parameter</b>
			</body>`))
		return
	}

	target := GetTarget(
		WorkerSpec{
			period:         time.Second * time.Duration(opts.Period),
			staleThreshold: uint(opts.StaleThreshold),
		},
		TargetSpec{
			host: targetParam,
		},
	)

	h := promhttp.HandlerFor(target.registry, promhttp.HandlerOpts{})
	h.ServeHTTP(w, r)
}

func main() {
	if _, err := flags.Parse(&opts); err != nil {
		os.Exit(0)
	}

	// NOTE: Disable all default log flags, custom logger handles formatting.
	log.SetFlags(0)

	// Set debug logging if set, init with WARN first as default
	logger.SetLogLevel(logger.WARN)
	if opts.Debug {
		logger.SetLogLevel(logger.DEBUG)
	}

	if opts.Version {
		fmt.Printf("%v, version %v (branch: %v, revision: %v)\n",
			BuildName,
			BuildVersion,
			BuildBranch,
			BuildCommit,
		)
		fmt.Printf("  build user:       %v\n", BuildUser)
		fmt.Printf("  build date:       %v\n", BuildDate)
		fmt.Printf("  go version:       %v\n", BuildGo)
		fmt.Printf("  platform:         %v/%v\n", BuildOs, BuildArch)
		os.Exit(0)
	}

	if _, err := os.Stat(opts.Fping); os.IsNotExist(err) {
		logger.Fatal("could not find fping at %q", opts.Fping)
	}
	http.Handle("/metrics", promhttp.Handler())
	http.HandleFunc("/probe", probeHandler)
	logger.Fatal("%v", http.ListenAndServe(opts.Listen, nil))
}
