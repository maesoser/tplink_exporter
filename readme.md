# tplink_exporter

Prometheus exporter for cheap TP-Link routers, like the [TL-WR841N](https://www.tp-link.com/en/products/details/cat-9_TL-WR841N.html).

Inspired by [this repository](https://github.com/mkubicek/tpylink).

The tplink package created for this exporter can be used for another projects besides this exporter.

![grafana_image](https://raw.githubusercontent.com/maesoser/tp_link_exporter/master/images/grafana.jpg)

## Usage

Just create a systemd service like this one:

```
[Unit]
Description=Node Exporter
Wants=network-online.target
After=network-online.target

[Service]
User=node_exporter
Group=node_exporter
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

LAN metrics include
- IP address
- MAC Address
- Device name
