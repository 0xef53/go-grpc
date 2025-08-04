package utils

import (
	"context"
	"math/rand"
	"net"
	"slices"
	"strings"
	"time"

	grpc_metadata "google.golang.org/grpc/metadata"
)

// ExtractRequestID tries to extract the request ID from the outgoing gRPC context.
// If no request ID is found in the outgoing metadata, the function generates a new one
// using [NewRequestID()].
func ExtractRequestID(ctx context.Context) string {
	reqID := NewRequestID()

	if md, ok := grpc_metadata.FromIncomingContext(ctx); ok {
		if v, ok := md["request-id"]; ok {
			reqID = v[0]
		}
	}

	return reqID
}

// NewRequestID generates a new request ID.
func NewRequestID() string {
	return RandString(6)
}

// RandString returns a random generated string with given length.
func RandString(length int) string {
	srcBytes := "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZ"
	seededRand := rand.New(rand.NewSource(time.Now().UnixNano()))
	strb := make([]byte, length)
	for i := range strb {
		strb[i] = srcBytes[seededRand.Intn(len(srcBytes))]
	}
	return string(strb)
}

// NormalizeHostport normalizes a hostport string by ensuring it has a port number.
func NormalizeHostport(hostport string) string {
	hostport = strings.TrimSpace(hostport)

	switch n := len(strings.Split(hostport, ":")); {
	case n == 1:
		// IPv4 without port
		return hostport + ":9191"
	case n == 2:
		// IPv4 with port
		return hostport
	}

	// IPv6 probably

	if strings.HasPrefix(hostport, "[") {
		switch parts := strings.Split(hostport, "]"); len(parts) {
		case 1:
			return hostport + ":9191"
		case 2:
			if len(parts[1]) == 0 {
				return hostport + ":9191"
			}
			return hostport
		}
	}

	return "[" + hostport + "]:9191"
}

// ParseBindings converts a string slice of IP addresses and interface names
// into a slice of [net.IP].
// Interface names are expanded into the set of IP addresses configured on the interface
// at the time the code is executed.
func ParseBindings(bindings ...string) ([]net.IP, error) {
	ifaces, err := net.Interfaces()
	if err != nil {
		return nil, err
	}

	m := make(map[string]net.IP)

	for _, v := range bindings {
		idx := slices.IndexFunc(ifaces, func(iface net.Interface) bool { return iface.Name == v })

		if idx < 0 {
			// Try to parse as IP address
			if ip := net.ParseIP(v); ip != nil {
				if _, ok := m[ip.String()]; !ok {
					m[ip.String()] = ip
				}

				continue
			}
		} else {
			// Perhaps this is a network interface name
			addrs, err := ifaces[idx].Addrs()
			if err != nil {
				return nil, err
			}

			for _, addr := range addrs {
				if ipnet, ok := addr.(*net.IPNet); ok {
					if ipnet.IP.IsLinkLocalUnicast() {
						continue
					}

					m[ipnet.IP.String()] = ipnet.IP
				}
			}
		}
	}

	addrs := make([]net.IP, 0, len(m))

	for _, ip := range m {
		addrs = append(addrs, ip)
	}

	return addrs, nil
}
