package main

import (
	"flag"
	"log"
	"net/http"
	"os"
	"strconv"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"

	"github.com/maesoser/tplink_exporter/macdb"
	"github.com/maesoser/tplink_exporter/tplink"
)

func GetEnvStr(name, value string) string {
	if os.Getenv(name) != "" {
		return os.Getenv(name)
	}
	return value
}

func main() {
	Address := flag.String(
		"a",
		GetEnvStr("TPLINK_ROUTER_ADDR", "192.168.0.1"),
		"Router's address",
	)
	Pass := flag.String(
		"w",
		GetEnvStr("TPLINK_ROUTER_PASSWD", "admin"),
		"Router's password",
	)
	User := flag.String(
		"u",
		GetEnvStr("TPLINK_ROUTER_USER", "admin"),
		"Router's username",
	)
	Port := flag.Int(
		"p",
		9300,
		"Prometheus port",
	)
	Verbose := flag.Bool(
		"v",
		false,
		"Verbose output",
	)
        Reboot := flag.Bool(
		"reboot",
		false,
		"Reboots the router",
	)
	Filename := flag.String(
		"f",
		GetEnvStr("TPLINK_ROUTER_MACS", "/etc/known_macs"),
		"MAC Database",
	)
	flag.Parse()

	macs := macdb.MACDB{}
	err := macs.Load(*Filename)
	if err != nil {
		log.Println("Unable to load MAC database:", err)
	} else {
		log.Printf("%d MACs loaded", macs.Size())
	}

	router := tplink.NewRouter(*Address, *User, *Pass)
	router.Verbose = *Verbose

	if (*Reboot){
		if err := router.Login(); err != nil {
                	log.Fatal("Error logging: %v", err)
        	}
		if err := router.Reboot(); err != nil {
			log.Fatal(err)
		}
		log.Println("Reboot command sent")
		os.Exit(1)
	}
	c := newRouterCollector(router, macs)
	prometheus.MustRegister(c)

	http.Handle("/metrics", promhttp.Handler())
	log.Fatal(http.ListenAndServe(":"+strconv.Itoa(*Port), nil))
}
