# tplink_exporter [![Build Status](https://travis-ci.org/maesoser/tplink_exporter.svg?branch=master)](https://travis-ci.org/maesoser/tplink_exporter) [![License](https://img.shields.io/github/license/maesoser/tpink_exporter)](https://www.gnu.org/licenses/gpl-3.0.html)

Prometheus exporter for cheap TP-Link routers, like the [TL-WR841N](https://www.tp-link.com/en/products/details/cat-9_TL-WR841N.html).

Inspired by [this repository](https://github.com/mkubicek/tpylink).

The tplink package created for this exporter can be used for another projects besides this one.

![grafana_image](https://github.com/maesoser/tplink_exporter/raw/master/images/grafana.jpg)

## Usage

tplink exporter has a few CLI arguments that you can configure like the following ones.

- **-a**: Router's IP address, you can also configure it by using the `TPLINK_ROUTER_ADDR` environment variable.
- **-w**: Router's password, you can also configure it by using the `TPLINK_ROUTER_PASSWD` environment variable.
- **-u**: Router's username, you can also configure it by using the `TPLINK_ROUTER_USER` environment variable.
- **-p**: Prometheus port.
- **-v**: Verbose output.
- **-f**: MAC Database, you can also configure it by using the `TPLINK_ROUTER_MACS` environment variable. By default should be located on `/etc/known_macs`.

## Metrics exposed

- **tplink_wan_rx_bytes:** Total bytes received
- **tplink_wan_tx_bytes:** Total bytes transmitted
- **tplink_lan_traffic_bytes:** bytes sent/received per device
- **tplink_lan_traffic_packets:** Packets sent/received per device
- **tplink_lan_leases_seconds:** Lease time left per device

LAN metrics include IP and MAC addresses and device name as labels. 

## Using systemd

You can run this agent as a service. First you need to compile it and then you just need create a systemd service like this one:

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

## Using docker compose

```
tplink_exporter:
  container_name: tplink_exporter
  restart: unless-stopped
  cpu_count: 1
  mem_limit: 16m
  build:
    context: ./containers/tplink_exporter
    dockerfile: Dockerfile
  ports:
   - "9300:9300"
  volumes:
   - "./config/tplink_exporter/known_macs:/etc/known_macs"
```
