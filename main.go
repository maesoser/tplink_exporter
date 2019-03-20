package main

import (
	"bufio"
	"flag"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"

	"github.com/maesoser/tplink_exporter/tplink"
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

func main() {
	Address := flag.String("a", "192.168.0.1", "Router's address")
	Pass := flag.String("w", "admin", "Router's password")
	User := flag.String("u", "admin", "Router's username")
	Port := flag.Int("p", 9300, "Prometheus port")
	Filename := flag.String("f", "/etc/known_macs", "MAC Database")

	macs, vendors, err := ReadMACList(*Filename)
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
