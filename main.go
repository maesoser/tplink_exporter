package main

import (
	"flag"
	"log"
	"net/http"
	"strconv"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"

	"github.com/maesoser/tplink_exporter/macdb"
	"github.com/maesoser/tplink_exporter/tplink"
)

func main() {
	Address := flag.String("a", "192.168.0.1", "Router's address")
	Pass := flag.String("w", "admin", "Router's password")
	User := flag.String("u", "admin", "Router's username")
	Port := flag.Int("p", 9300, "Prometheus port")
	Filename := flag.String("f", "/etc/known_macs", "MAC Database")

	macs, vendors, err := macdb.Open(*Filename)
	if err != nil {
		log.Println("Unable to load MAC database:", err)
	} else {
		log.Printf("%d custom MACs loaded\n", len(macs))
		log.Printf("%d vendor MACs loaded\n", len(vendors))
	}

	router := tplink.NewRouter(*Address, *User, *Pass)

	c := newRouterCollector(router, macs, vendors)
	prometheus.MustRegister(c)

	http.Handle("/metrics", promhttp.Handler())
	log.Fatal(http.ListenAndServe(":"+strconv.Itoa(*Port), nil))
}
