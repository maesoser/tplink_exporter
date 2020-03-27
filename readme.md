# tplink_exporter [![Build Status](https://travis-ci.org/maesoser/tplink_exporter.svg?branch=master)](https://travis-ci.org/maesoser/tplink_exporter) [![License](https://img.shields.io/badge/License-Apache%202.0-blue.svg)](https://opensource.org/licenses/Apache-2.0)

Prometheus exporter for cheap TP-Link routers, like the [TL-WR841N](https://www.tp-link.com/en/products/details/cat-9_TL-WR841N.html).

Inspired by [this repository](https://github.com/mkubicek/tpylink).

The tplink package created for this exporter can be used for another projects besides this one.

![grafana_image](https://github.com/maesoser/tplink_exporter/raw/master/images/grafana.jpg)

## Usage

First compile it, of course. Then, you just need create a systemd service like this one:

```
[Unit]
Description=TP-Link Exporter
Wants=network-online.target
After=network-online.target

[Service]
User=tplink_exporter
Group=tplink_exporter
Type=simple
ExecStart=/usr/local/bin/tplink_exporter

[Install]
WantedBy=multi-user.target
```

Configure it and launch it:

```
sudo systemctl enable tplink_exporter
sudo systemctl start tplink_exporter
```

## Command line flags

- **-a**: Router's IP address
- **-w**: Router's password
- **-u**: Router's User
- **-p**: Prometheus port

## Metrics exposed

- **tplink_wan_rx_kbytes:** Total kbytes received
- **tplink_wan_tx_kbytes:** Total kbytes transmitted
- **tplink_lan_traffic_kbytes:** KBytes sent/received per device
- **tplink_lan_traffic_packets:** Packets sent/received per device
- **tplink_lan_leases_seconds:** Lease time left per device

LAN metrics include IP and MAC addresses and device name as labels. 


```
229 #  tplink_exporter:
230 #    container_name: tplink_exporter
231 #    restart: unless-stopped
232 #    cpu_count: 1
233 #    mem_limit: 16m
234 #    build:
235 #      context: ./containers/tplink_exporter
236 #      dockerfile: Dockerfile
237 #    ports:
238 #     - "9300:9300"
239 #    volumes:
240 #     - "./config/tplink_exporter/known_macs:/etc/known_macs"
```
