# Spill

Utility to quickly scan over a signle IP / CIDR to search for OpenPrinting CVE 2024-47176 on UDP 631

This utility is quick and ugly - but could be useful to some people.

## Build Project

```
go build .
```

## Quick usage (single IP)

```
go run main.go -ip <target-ip> -port 631 -dest <your listening ip> -destport <your listening port>
OR
./spill -ip <target-ip> -port 631 -dest <your listening ip> -destport <your listening port>
```

## Quick usage (CIDR)

```
go run main.go -cidr <target-range> -port 631 -dest <your listening ip> -destport <your listening port>
OR
./spill -cidr <target-range> -port 631 -dest <your listening ip> -destport <your listening port>
```
