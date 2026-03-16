package ports

import (
	"fmt"
	"net"
)

// InUse returns true if the given port is already in use on localhost.
// Uses TCP listen to detect port occupancy (cross-platform).
func InUse(port int) bool {
	listener, err := net.Listen("tcp", net.JoinHostPort("127.0.0.1", fmt.Sprintf("%d", port)))
	if err != nil {
		return true // Assume in use on any error
	}
	listener.Close()
	return false
}
