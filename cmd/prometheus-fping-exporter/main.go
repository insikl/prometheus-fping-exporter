package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	flags "github.com/jessevdk/go-flags"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var opts struct {
	Listen  string `short:"l" long:"listen" description:"Listen address" value-name:"[HOST]:PORT" default:":9605"`
	Period  uint   `short:"p" long:"period" description:"Period in seconds, should match Prometheus scrape interval" value-name:"SECS" default:"60"`
	Fping   string `short:"f" long:"fping"  description:"Fping binary path" value-name:"PATH" default:"/usr/bin/fping"`
	Count   uint   `short:"c" long:"count"  description:"Number of pings to send at each period" value-name:"N" default:"20"`
	Version bool   `long:"version" description:"Show version"`
}

// Build information.
const (
	BuildVersion = "0.1.2"
)

// Build information populated at build-time.
var (
	BuildName     string
	BuildCommit   string
	BuildBranch   string
	BuildUser     string
	BuildDate     string
	BuildGo       string
	BuildPlatform string
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
			period: time.Second * time.Duration(opts.Period),
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
		fmt.Printf("  platform:         %v\n", BuildPlatform)
		os.Exit(0)
	}

	if _, err := os.Stat(opts.Fping); os.IsNotExist(err) {
		fmt.Printf("could not find fping at %q\n", opts.Fping)
		os.Exit(1)
	}
	http.Handle("/metrics", promhttp.Handler())
	http.HandleFunc("/probe", probeHandler)
	log.Fatal(http.ListenAndServe(opts.Listen, nil))
}
