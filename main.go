package main

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"log"
	"net"
	"os"
)

func lines(r io.Reader) <-chan string {
	out := make(chan string)

	go func() {
		br := bufio.NewReader(r)

		for {
			s, err := br.ReadString(('\n'))

			out <- s

			if err != nil {
				close(out)
				return
			}
		}
	}()

	return out
}

// Listens on address for the first connection, and returns it
func listen1(address string) (net.Conn, error) {
	listener, err := net.Listen("tcp", address)

	if err != nil {
		return nil, err
	}

	conn, err := listener.Accept()

	if err != nil {
		return nil, err
	}

	return conn, nil
}

type packetConnWithAddr struct {
	net.PacketConn
	net.Addr
	sendBuffer [][]byte // holds messages that get sent before we have an Addr
}

func (conn *packetConnWithAddr) Read(b []byte) (n int, err error) {
	n, addr, err := conn.ReadFrom(b)

	if conn.Addr == nil {
		conn.Addr = addr

		for _, b := range conn.sendBuffer {
			conn.WriteTo(b, addr)
		}
	}

	return n, err
}

func (conn *packetConnWithAddr) RemoteAddr() net.Addr {
	return conn.Addr
}

func (conn *packetConnWithAddr) Write(b []byte) (n int, err error) {
	if conn.Addr == nil {
		conn.sendBuffer = append(conn.sendBuffer, b)

		return len(b), nil
	}

	return conn.WriteTo(b, conn.Addr)
}

func newPacketConnWithAddr(conn net.PacketConn) packetConnWithAddr {
	sendBuffer := make([][]byte, 0, 10)

	return packetConnWithAddr{conn, nil, sendBuffer}
}

// Same as listen1 but for UDP
func listen1u(address string) (net.Conn, error) {
	pconn, err := net.ListenPacket("udp", address)

	if err != nil {
		return nil, err
	}

	conn := newPacketConnWithAddr(pconn)

	return &conn, nil
}

func main() {
	log.SetFlags(0)

	listen := flag.Bool("l", false, "Listen")
	udp := flag.Bool("u", false, "Use UDP instead of TCP")

	flag.Parse()

	if flag.NArg() < 1 {
		log.Fatalf("usage: %s hostname port\n", os.Args[0])
	}

	var host, port string

	if flag.NArg() == 1 {
		host = ""
		port = flag.Arg(0)
	} else {
		host = flag.Arg(0)
		port = flag.Arg(1)
	}

	address := net.JoinHostPort(host, port)

	var conn net.Conn
	var err error
	if *listen && *udp {
		conn, err = listen1u(address)
	} else if *listen {
		conn, err = listen1(address)
	} else {
		network := "tcp"
		if *udp {
			network = "udp"
		}

		conn, err = net.Dial(network, address)
	}

	if err != nil {
		log.Fatal(err)
	}

	remote := lines(conn)
	local := lines(os.Stdin)

	for {
		select {
		case s, ok := <-remote:
			if !ok {
				return
			}

			fmt.Print(s)
		case s := <-local:
			_, err := conn.Write([]byte(s))

			if err != nil {
				log.Fatal(err)
			}
		}
	}
}
