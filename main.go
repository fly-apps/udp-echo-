// the simplest example of a Fly.io app that listens on both UDP and TCP;
// the envar ECHO_PORT configures the port (default 5000)

package main

import (
	"bufio"
	"errors"
	"fmt"
	"log"
	"net"
	"os"
	"strconv"
)

var (
	port int = 5000
)

func init() {
	if v := os.Getenv("ECHO_PORT"); v != "" {
		p, err := strconv.Atoi(v)
		if err != nil {
			log.Fatalf("can't parse ECHO_PORT: %s", err)
		}

		port = p
	}
}

func main() {
	// what you want to be able to do here is use the same listening address
	// for both TCP and UDP, but we live in a fallen world and Fly.io
	// discriminates UDP traffic on the `fly-global-services` address;
	// if you bind instead to all addresses, like we do for TCP, your
	// outgoing traffic will have the wrong source address.
	//
	// this catches me up every time I write a UDP anything on Fly.io,
	// and I wrote Fly.io's UDP feature. All I can do is apologize.

	udp, err := net.ListenPacket("udp", fmt.Sprintf("fly-global-services:%d", port))
	if err != nil {
		log.Fatalf("can't listen on %d/udp: %s", port, err)
	}

	tcp, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		log.Fatalf("can't listen on %d/tcp: %s", port, err)
	}

	go handleTCP(tcp)

	handleUDP(udp)
}

// everything below this point is textbook Go code.

func handleTCP(l net.Listener) {
	for {
		conn, err := l.Accept()
		if err != nil {
			if errors.Is(err, net.ErrClosed) {
				return
			}

			log.Printf("error accepting on %d/tcp: %s", port, err)
			continue
		}

		go handleConnection(conn)
	}
}

// see, just a TCP connection handler.
func handleConnection(c net.Conn) {
	defer c.Close()

	lines := bufio.NewReader(c)

	for {
		line, err := lines.ReadString('\n')
		if err != nil {
			return
		}

		c.Write([]byte(line))
	}
}

func handleUDP(c net.PacketConn) {
	packet := make([]byte, 2000)

	for {
		n, addr, err := c.ReadFrom(packet)
		if err != nil {
			if errors.Is(err, net.ErrClosed) {
				return
			}

			log.Printf("error reading on %d/udp: %s", port, err)
			continue
		}

		c.WriteTo(packet[:n], addr)
	}
}
