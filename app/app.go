package main

import (
	"encoding/hex"
	"fmt"
	"io"
	"net"
	"net/http"
	"strconv"
	"time"
)

func handleRequest(w http.ResponseWriter, r *http.Request) {
	// Combine URL query parameters and POST form values
	params := r.URL.Query()

	// Check if the request method is GET or POST
	if r.Method != http.MethodGet && r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Check if a valid hex string is present in the query parameters or form data
	var packetHex string
	for k, v := range params {
		if len(v) > 0 && hex.IsHex(v[0]) {
			packetHex = v[0]
			break
		}
	}

	if packetHex == "" {
		http.Error(w, "No valid hex packet found in request", http.StatusBadRequest)
		return
	}

	// Validate the remote address
	remoteAddr := net.JoinHostPort("engage.cloudflareclient.com", "2408")
	if _, err := net.ResolveUDPAddr("udp", remoteAddr); err != nil {
		http.Error(w, fmt.Sprintf("Invalid remote address: %v", err), http.StatusBadRequest)
		return
	}

	// Send the UDP packet and handle the response
	response, err := sendUdpPacket(remoteAddr, packetHex)
	if err != nil {
		http.Error(w, fmt.Sprintf("Error: %v", err), http.StatusInternalServerError)
	} else {
		fmt.Fprintf(w, "Response: %s\n", response)
	}
}

func sendUdpPacket(remoteAddr string, packetHex string) (string, error) {
	// Check if the packetHex string is a valid hex string
	if _, err := hex.DecodeString(packetHex); err != nil {
		return "", fmt.Errorf("invalid hex string: %w", err)
	}

	conn, err := net.Dial("udp", remoteAddr)
	if err != nil {
		return "", fmt.Errorf("dial error: %w", err)
	}
	defer conn.Close()

	// Set a read timeout based on the size of the packet
	conn.SetReadDeadline(time.Now().Add(time.Duration(len(packetHex)/2) * time.Millisecond))

	packet, err := hex.DecodeString(packetHex)
	if err != nil {
		return "", fmt.Errorf("invalid hex string: %w", err)
	}

	if _, err = conn.Write(packet); err != nil {
		return "", fmt.Errorf("write error: %w", err)
	}

	buffer := make([]byte, 1024)
	n, err := conn.Read(buffer)
	if err != nil {
		return "", fmt.Errorf("read error: %w", err)
	}

	return hex.EncodeToString(buffer[:n]), nil
}

func main() {
	http.HandleFunc("/", handleRequest)
	port, err := strconv.Atoi("8080")
	if err != nil {
		fmt.Println("Error starting server:", err)
		return
	}
	http.ListenAndServe(":"+strconv.Itoa(port), nil)
}
