# Packet sniffer

Packet sniffing tool is written by Go based on `gopacket/pcap`.

See detail documentation at:

- [MySQL error response sniffer](./docs/mysql-error-response/README.md)
- [Redis error response sniffer](./docs/mysql-error-response/README.md)
- And more ...


## Build

### Binary

Require: Golang ver >= 1.25

#### Linux build

```bash
> GOOS=linux GOARCH=amd64 go build .
```

#### Docker image

```bash
> docker build . -t packet-sniffer
```
