# Spill

Utility to quickly scan over a signle IP / CIDR to search for OpenPrinting CVE 2024-47176 on UDP 631

This utility is quick and ugly - but could be useful to some people.

Exploit has been left out purposefully as this is more designed to scan and provide awareness to users.

## Build Project

```
go build .
```

## Usage

```
Usage of ./spill:
  -cidr string
    	CIDR block for network scanning
  -dest string
    	Destination IP address
  -destport string
    	Destination UDP port (default "12345")
  -ip string
    	Target IP address
  -port string
    	Target UDP port (default "631")
```

## Example (single IP)

```
go run main.go -ip <target-ip> -port 631 -dest <your listening ip> -destport <your listening port>
OR
./spill -ip <target-ip> -port 631 -dest <your listening ip> -destport <your listening port>
```

## Example (CIDR)

```
go run main.go -cidr <target-range> -port 631 -dest <your listening ip> -destport <your listening port>
OR
./spill -cidr <target-range> -port 631 -dest <your listening ip> -destport <your listening port>
```

## Example Output

```zsh
┌──(kali㉿kali-raspberry-pi)-[~/spill]
└─$ ./spill -ip 192.168.50.174 -port 631 -dest 192.168.50.175 -destport 12345
2024/09/27 03:28:12 Starting HTTP server on port 12345...
2024/09/27 03:28:12 Received POST request: 192.168.50.174:55580
2024/09/27 03:28:12 Received POST request: 192.168.50.174:55592
2024/09/27 03:28:12 Received POST request: 192.168.50.174:55614
2024/09/27 03:28:13 Received POST request: 192.168.50.174:55620
2024/09/27 03:28:13 Received POST request: 192.168.50.174:55636
2024/09/27 03:28:13 Received POST request: 192.168.50.174:55662
```
