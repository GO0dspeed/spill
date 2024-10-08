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
Usage:
  -dest string
    	IP address for callbacks
  -destport string
    	TCP port to listen on for HTTP callbacks (default "12345")
  -port string
    	Target UDP port (default "631")
  -target string
    	Target IP address, CIDR network, or filename to scan
```

## Example (single IP)

```
go run main.go -ip <target-ip> -port 631 -dest <your listening ip> -destport <your listening port>
OR
./spill -ip <target-ip> -port 631 -dest <your listening ip> -destport <your listening port>
```

## Example (CIDR)

```
go run main.go -target <target-range> -port 631 -dest <your listening ip> -destport <your listening port>
OR
./spill -target <target-range> -port 631 -dest <your listening ip> -destport <your listening port>
```

## Example (file with CIDR networks)

```
go run main.go -target <path/to/target/file> -port 631 -dest <your listening ip> -destport <your listening port>
OR
./spill -target <path/to/target/file> -port 631 -dest <your listening ip> -destport <your listening port>

```

## Example Output

```zsh
┌──(kali㉿kali-raspberry-pi)-[~/spill]
└─$ ./spill -target 10.1.80.0/24 -dest 10.110.123.74 -destport 9003

	  . .
	  .. . *.
- -_ _-__-0oOo
 _-_ -__ -||||)
    ______||||______
~~~~~~~~~~^""' Spill

2024/09/27 13:55:37 Starting HTTP server on port 9003...
Packet Progress 100% |████████████████████████████████████████|
2024/09/27 13:55:37 Received POST request: 10.1.80.89:56952
2024/09/27 13:55:37 Received POST request: 10.1.80.85:37606
```
