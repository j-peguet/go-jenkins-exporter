package exporter

import (
	"net/http"

	"github.com/goodbins/go-jenkins-exporter/config"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/sirupsen/logrus"
)

// Serve serves the metrics, helthcheck /ping and a redirection on /
func Serve() {
	// Print start message
	logrus.Info("Starting go-jenkins-exporter")

	// Launch metrics update go routine
	go SetGauges()

	// Handle routes: / /ping /metrics
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`<html>
		<head><title>Go Jenkins Exporter</title></head>
		<body>
		<h1>Go Jenkins Exporter</h1>
		<p><a href="` + config.Global.MetricsPath + `">Metrics</a></p>
		</body></html>`))
	})
	http.HandleFunc("/ping", Ping)
	http.Handle(config.Global.MetricsPath, promhttp.Handler())

	// Listen and serve
	logrus.Info("Listening on " + config.Global.ExporterHostPort + " ...")
	logrus.Fatal(http.ListenAndServe(config.Global.ExporterHostPort, nil))
}
