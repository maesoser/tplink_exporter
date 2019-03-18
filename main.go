package main

import (
	"bufio"
	"flag"
	"fmt"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
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

type MACList map[string]string

func ReadMACList(filename string) (MACList, MACList, error) {
	var customMACs = make(map[string]string)
	var vendorMACs = make(map[string]string)
	if len(filename) == 0 {
		return customMACs, vendorMACs, nil
	}
	file, err := os.Open(filename)
	if err != nil {
		return customMACs, vendorMACs, err
	}
	defer file.Close()
	reader := bufio.NewReader(file)
	for {
		line, err := reader.ReadString('\n')
		line = strings.Replace(line, "\n", "", -1)
		elements := strings.Split(line, "=")
		if len(elements) == 2 {
			if key := strings.TrimSpace(elements[0]); len(key) > 0 {
				mac := strings.TrimSpace(elements[1])
				if len(key) == 8 {
					vendorMACs[key] = mac
				} else {
					customMACs[key] = mac
				}
			}
		}
		if err == io.EOF {
			break
		}
		if err != nil {
			return customMACs, vendorMACs, err
		}
	}
	return customMACs, vendorMACs, nil
}

func MACLookup(mac string, custom, vendor MACList) string {
	result := custom[mac]
	if len(result) != 0 {
		return result
	}
	return vendor[mac[:8]]
}

// Refresh refreshes the data to be exposed on /metrics
func Refresh(router *tplink.Router, macs, vendors MACList) error {
	err := router.Login()
	if err != nil {
		return fmt.Errorf("Error logging: %v", err)
	}
	rx, tx, err := router.GetWANTraffic()
	if err != nil {
		return fmt.Errorf("Error getting WAN metrics: %v", err)
	}
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
		name := MACLookup(client.MAC, macs, vendors)
		if len(name) == 0{
			name = client.Name
		}
		LANTraffic.With(prometheus.Labels{
			"mac":  client.MAC,
			"addr": client.Addr,
			"name": name,
		}).Set(client.Bytes)
		LANLeases.With(prometheus.Labels{
			"mac":  client.MAC,
			"addr": client.Addr,
			"name": name,
		}).Set(client.Lease)
		LANPackets.With(prometheus.Labels{
			"mac":  client.MAC,
			"addr": client.Addr,
			"name": name,
		}).Set(client.Packets)
	}
	//router.Logout()
	return nil

}

func PeriodicRefresh(router *tplink.Router, interval time.Duration, macs, vendors MACList) {
	for {
		start := time.Now()
		err := Refresh(router, macs, vendors)
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
	Filename := flag.String("f", "/etc/known_macs", "MAC Database")

	macs, vendors, err := ReadMACList(*Filename)
	if err != nil {
		log.Println("Unable to load MAC database:", err)
	}
	log.Printf("%d custom MACs loaded\n", len(macs))
	log.Printf("%d vendor MACs loaded\n", len(vendors))

	router := tplink.NewRouter(*Address, *User, *Pass)
	go PeriodicRefresh(router, 60*time.Second, macs, vendors)

	http.Handle("/metrics", promhttp.Handler())
	log.Fatal(http.ListenAndServe(":"+strconv.Itoa(*Port), nil))
}
