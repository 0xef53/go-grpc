package server

import (
	"crypto/tls"
	"fmt"
	"net"
	"os"
	"path/filepath"

	"github.com/0xef53/go-grpc/utils"
)

// Config represents a gRPC server / gRPC Gateway server config
type Config struct {
	// Bindings specifies the network addresses to listen on.
	// Network interface names are allowed here and will be expanded to
	// a list of [net.IP] addresses configured on the interface at
	// the time the code is executed.
	Bindings []string `gcfg:"listen" ini:"listen,,allowshadow" json:"listen"`

	// Port is a number of gRPC server port
	Port uint16 `gcfg:"port" ini:"port" json:"port"`

	// GatewayPort is a number of gRPC Gateway server port
	GatewayPort uint16 `gcfg:"port-gw" ini:"port-gw" json:"port_gw"`

	// GRPCSocketPath specifies the path to a Unix socket
	// on which the gRPC server will listen, in addition to the bindings
	// defined above.
	// This socket is also used by the gRPC gateway.
	GRPCSocketPath   string `gcfg:"-" ini:"-" json:"-"`
	GRPCSecureSocket bool   `gcfg:"-" ini:"-" json:"-"`

	// TLSConfig is used to configure TLS encryption for the connection.
	TLSConfig *tls.Config `gcfg:"-" ini:"-" json:"-"`
}

// Defaults sets default values for unpopulated fields.
func (c *Config) Defaults() {
	if len(c.Bindings) == 0 {
		c.Bindings = []string{"127.0.0.1"}
	}

	if c.Port == 0 {
		c.Port = 9191
	}

	if c.GatewayPort == 0 {
		c.GatewayPort = 9090
	}

	if len(c.GRPCSocketPath) == 0 {
		c.GRPCSocketPath = filepath.Join("/run", fmt.Sprintf("%s_%d.sock", filepath.Base(os.Args[0]), os.Getpid()))
	}
}

// Validate checks that all struct parameters are filled correctly.
func (c *Config) Validate() error {
	if len(c.Bindings) == 0 {
		return fmt.Errorf("no one listener defined")
	}

	if c.Port == 0 {
		return fmt.Errorf("gRPC port is not set")
	}

	if c.GatewayPort == 0 {
		return fmt.Errorf("gRPC Gateway port is not set")
	}

	if c.Port == c.GatewayPort {
		return fmt.Errorf("gRPC port cannot be the same as gRPC Gateway port")
	}

	if len(c.GRPCSocketPath) == 0 {
		return fmt.Errorf("gRPC unix socket path is not set")
	}

	return nil
}

// listeners creates TCP listeners for each provided IP address on the given port.
func (c *Config) listeners(addrs []net.IP, port uint16) (_ []net.Listener, err error) {
	listeners := make([]net.Listener, 0, len(addrs))

	defer func() {
		if err != nil {
			for _, l := range listeners {
				l.Close()
			}
		}
	}()

	for _, ipaddr := range addrs {
		var hostport string

		if ipaddr.To4() != nil {
			hostport = fmt.Sprintf("%s:%d", ipaddr.String(), port)
		} else {
			hostport = fmt.Sprintf("[%s]:%d", ipaddr.String(), port)
		}

		l, err := net.Listen("tcp", hostport)
		if err != nil {
			return nil, err
		}

		listeners = append(listeners, l)
	}

	return listeners, nil
}

// GetListeners returns a list of TCP listeners for the gRPC server
// obtained from the "Bindings" field.
func (c *Config) GetListeners() ([]net.Listener, error) {
	addrs, err := utils.ParseBindings(c.Bindings...)
	if err != nil {
		return nil, err
	}

	return c.listeners(addrs, c.Port)
}

// GetGatewayListeners returns a list of TCP listeners for the gRPC Gateway server
// obtained from the "Bindings" field.
func (c *Config) GetGatewayListeners() ([]net.Listener, error) {
	addrs, err := utils.ParseBindings(c.Bindings...)
	if err != nil {
		return nil, err
	}

	return c.listeners(addrs, c.GatewayPort)
}
