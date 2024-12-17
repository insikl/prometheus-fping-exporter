# prometheus-fping-exporter

prometheus-fping-exporter allows you to run measure network latency using fping and
prometheus. Compared to [blackbox-exporter](https://github.com/prometheus/blackbox_exporter),
it gives you additionally latency distribution and a packet loss statistics.
Also, it is probably better performing thanks to fping.

**WARNING: This is currently a work in progress, the code is not production-ready yet**

This graph shows the fping\_rtt summary as "[SmokePing](https://oss.oetiker.ch/smokeping/)"-like graph in Grafana:
![screenshot](README_screenshot.png)

## Build Steps (manual)

1. Clone repo

   ```
   git clone <this repo url>
   ```

2. Change into cloned repos directory

   ```
   cd <this repo name>
   ```

3. Build go binaries with the following command, this will place files in the
   `bin/*` directory.

   ```
   make build
   ```

   > NOTE The `Makefile` build will override the `GOBIN` environment to force
   > install of binaries into the `bin/*` directory. The `make` process will
   > look for `cmd/*/main.go` and create binary named after the directory.

4. Run script after configuration (detailed below)

## Usage

1. Start fping-exporter as follows:

   ```
     prometheus-fping-exporter [OPTIONS]
   
   Application Options:
     -l, --listen=[HOST]:PORT    Listen address (default: :9605)
     -p, --period=SECS           Period in seconds, should match Prometheus scrape interval (default: 60)
     -f, --fping=PATH            Fping binary path (default: /usr/bin/fping)
     -c, --count=N               Number of pings to send at each period (default: 20)
   
   Help Options:
     -h, --help                  Show this help message
   ```

2. Configure Prometheus to use this, as you would with blackbox-exporter. For example:

   ```
   global:
     scrape_interval: 60s
   
   scrape_configs:
     - job_name: test
       metrics_path: /probe
       static_configs:
       - targets:
         - "localhost:9605!8.8.4.4"
         - "localhost:9605!8.8.8.8"
    relabel_configs:
      # Set query param `?target=<fping_target>`
      - source_labels: [__address__]
        target_label: __param_target
        regex: '.*!(.*)'
        replacement: $1

      # set Prometheus instance to include target and source host pulled from
      # the `__address__` field.
      - source_labels: [__address__]
        target_label: instance
        regex: '(.*)!(.*)'
        replacement: '$2@$1'

      # get real target (eg `localhost:9605`) and replace `__address__`
      - source_labels: [__address__]
        target_label: __address__
        regex: '(.*)!.*'
        replacement: $1
   ```

## Metrics

prometheus-fping-exporter produces the following metrics:

- `fping_sent_count`: Number of sent probes
- `fping_lost_count`: Number of lost probes
- `fping_rtt_count`: Number of measured latencies (successful probes)
- `fping_rtt_sum`: Sum of measured latencies
- `fping_rtt`: Summary of measured latencies

