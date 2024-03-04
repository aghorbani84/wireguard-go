package psiphon

import (
	"encoding/hex"
	"fmt"
	"io"
	"net"
	"net/http"
	"time"
)

func handleRequest(w http.ResponseWriter, r *http.Request) {
	// Combine URL query parameters and POST form values
	params := r.URL.Query()
	if err := r.ParseForm(); err == nil {
		for k, v := range r.PostForm {
			params[k] = v
		}
	}

	// Find the first valid hex string in parameters
	var packetHex string
	for _, v := range params {
		if len(v) > 0 && isHex(v[0]) {
			packetHex = v[0]
			break
		}
	}

	if packetHex == "" {
		http.Error(w, "No valid hex packet found in request", http.StatusBadRequest)
		return
	}

	// Assuming remote host and port are provided
	remoteAddr := net.JoinHostPort("engage.cloudflareclient.com", "2408")

	response, err := sendUdpPacket(remoteAddr, packetHex)
	if err != nil {
		http.Error(w, fmt.Sprintf("Error: %v", err), http.StatusInternalServerError)
	} else {
		fmt.Fprintf(w, "Response: %s\n", response)
	}
}

func sendUdpPacket(remoteAddr string, packetHex string) (string, error) {
	packet, err := hex.DecodeString(packetHex)
	if err != nil {
		return "", fmt.Errorf("invalid hex string: %w", err)
	}

	conn, err := net.Dial("udp", remoteAddr)
	if err != nil {
		return "", fmt.Errorf("dial error: %w", err)
	}
	defer conn.Close()

	if _, err = conn.Write(packet); err != nil {
		return "", fmt.Errorf("write error: %w", err)
	}

	conn.SetReadDeadline(time.Now().Add(5 * time.Second))

	buffer := make([]byte, 1024)
	n, err := conn.Read(buffer)
	if err != nil {
		return "", fmt.Errorf("read error: %w", err)
	}

	return hex.EncodeToString(buffer[:n]), nil
}

func isHex(s string) bool {
	_, err := hex.DecodeString(s)
	return err == nil
}

func main() {
	http.HandleFunc("/", handleRequest)
	http.ListenAndServe(":8080", nil)
}
