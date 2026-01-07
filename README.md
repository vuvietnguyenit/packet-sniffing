# Packet sniffer

Packet sniffing tool is written by Go based on `gopacket/pcap`.

See detail documentation at:

- [MySQL error response sniffer](./docs/mysql-error-response/README.md)
- [Redis error response sniffer](./docs/redis-error-response/README.md)
- And more ...

## Usage

```sh
./packet-sniffer -h
Usage:
  packet-sniffer [command]

Available Commands:
  completion           Generate the autocompletion script for the specified shell
  help                 Help about any command
  mysql-error-response Record error responses are sent to mysql-client and export it as Prometheus exporter base on libpcap
  redis-error-response Record error reponses on Redis instance to clients base and export to Prometheus metrics base on libpcap

Flags:
  -h, --help   help for packet-sniffer

Use "packet-sniffer [command] --help" for more information about a command.
```

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
