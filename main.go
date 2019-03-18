package main

import (
	"flag"
	"fmt"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/maesoser/tplink_exporter/tplink"
)

var (
	txWANTraffic = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "tplink_wan_tx_kbytes",
		Help: "Total kbytes transmitted",
	})
	rxWANTraffic = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "tplink_wan_rx_kbytes",
		Help: "Total kbytes received ",
	})
	LANTraffic = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "tplink_lan_traffic_kbytes",
		Help: "KBytes sent/received per device",
	},
		[]string{"mac", "addr", "name"},
	)
	LANPackets = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "tplink_lan_traffic_packets",
		Help: "Packets sent/received per device",
	},
		[]string{"mac", "addr", "name"},
	)
	LANLeases = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "tplink_lan_lease_seconds",
		Help: "Lease seconds left",
	},
		[]string{"mac", "addr", "name"},
	)
	scrapTime = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "tplink_scrap_duration_seconds",
		Help: "Time that took the scrapping process",
	})
)

// Refresh refreshes the data to be exposed on /metrics
func Refresh(router *tplink.Router) error {
	err := router.Login()
	if err != nil {
		return fmt.Errorf("Error logging: %v", err)
	}
	rx, tx, err := router.GetWANTraffic()
	if err != nil {
		return fmt.Errorf("Error getting WAN metrics: %v", err)
	}

	//log.Printf("WAN: TX: %f\tRX: %f\n", tx, rx)
	rxWANTraffic.Set(tx)
	txWANTraffic.Set(rx)

	clients, err := router.GetClients()
	if err != nil {
		return fmt.Errorf("Error getting WAN metrics: %v", err)
	}
	clients, err = router.GetLANTraffic(clients)
	if err != nil {
		return fmt.Errorf("Error getting LAN metrics: %v", err)
	}
	for _, client := range clients {
		//log.Printf("LAN: %v\n", client)
		LANTraffic.With(prometheus.Labels{
			"mac":  client.MAC,
			"addr": client.Addr,
			"name": client.Name,
		}).Set(client.Bytes)
		LANLeases.With(prometheus.Labels{
			"mac":  client.MAC,
			"addr": client.Addr,
			"name": client.Name,
		}).Set(client.Lease)
		LANPackets.With(prometheus.Labels{
			"mac":  client.MAC,
			"addr": client.Addr,
			"name": client.Name,
		}).Set(client.Packets)
	}
	//router.Logout()
	return nil

}

func PeriodicRefresh(router *tplink.Router, interval time.Duration) {
	for {
		start := time.Now()
		//log.Println("Updating Router stats")
		err := Refresh(router)
		elapsed := time.Since(start)
		if err != nil {
			log.Println("Error updating Router stats")
			log.Println(err)
		} else {
			scrapTime.Set(float64(elapsed.Seconds()))
			time.Sleep(interval)
		}
	}
}

func init() {
	prometheus.MustRegister(txWANTraffic)
	prometheus.MustRegister(rxWANTraffic)
	prometheus.MustRegister(LANLeases)
	prometheus.MustRegister(LANPackets)
	prometheus.MustRegister(LANTraffic)
	prometheus.MustRegister(scrapTime)
}

func main() {
	Address := flag.String("a", "192.168.0.1", "Router's address")
	Pass := flag.String("w", "admin", "Router's password")
	User := flag.String("u", "admin", "Router's username")
	Port := flag.Int("p", 9300, "Prometheus port")

	router := NewRouter(*Address, *User, *Pass)
	go PeriodicRefresh(router, 60*time.Second)

	http.Handle("/metrics", promhttp.Handler())
	log.Fatal(http.ListenAndServe(":"+strconv.Itoa(*Port), nil))
}
