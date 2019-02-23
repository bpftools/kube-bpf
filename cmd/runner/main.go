package main

import (
	"log"
	"net/http"

	"github.com/leodido/bpf-operator/mapcollector"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func main() {
	reg := prometheus.NewRegistry()

	c := mapcollector.New("example/dummy.o")
	if err := c.Setup(); err != nil {
		log.Fatal(err)
	}

	log.Println("Starting ...")

	if err := reg.Register(c); err != nil {
		log.Fatalf("Error during collector registration: %s", err)
	}

	promHandler := promhttp.HandlerFor(reg, promhttp.HandlerOpts{})
	http.Handle("/metrics", promHandler)

	listenAddress := ":9387"
	log.Printf("Listening on %s ...", listenAddress)
	if err := http.ListenAndServe(listenAddress, nil); err != http.ErrServerClosed {
		log.Fatalf("Error while starting bpfrunner: %v\n", err)
	}
}
