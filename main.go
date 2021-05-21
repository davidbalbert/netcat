package main

import (
	"flag"
	"io"
	"log"
	"net"
	"os"
)

// Listens on address for the first connection, and returns it
func listen1(address string) (net.Conn, error) {
	listener, err := net.Listen("tcp", address)

	if err != nil {
		return nil, err
	}

	conn, err := listener.Accept()
	defer listener.Close()

	if err != nil {
		return nil, err
	}

	return conn, nil
}

type boundPacketConn struct {
	net.PacketConn
	net.Addr
	sendBuffer [][]byte // holds messages that get sent before we have an Addr
}

func (conn *boundPacketConn) Read(b []byte) (n int, err error) {
	n, addr, err := conn.ReadFrom(b)

	if conn.Addr == nil {
		conn.Addr = addr

		for _, b := range conn.sendBuffer {
			conn.WriteTo(b, addr)
		}
	}

	return n, err
}

func (conn *boundPacketConn) RemoteAddr() net.Addr {
	return conn.Addr
}

func (conn *boundPacketConn) Write(b []byte) (n int, err error) {
	if conn.Addr == nil {
		conn.sendBuffer = append(conn.sendBuffer, b)

		return len(b), nil
	}

	return conn.WriteTo(b, conn.Addr)
}

func newBoundPacketConn(conn net.PacketConn) boundPacketConn {
	sendBuffer := make([][]byte, 0, 10)

	return boundPacketConn{conn, nil, sendBuffer}
}

// Same as listen1 but for UDP
func listen1u(address string) (net.Conn, error) {
	pconn, err := net.ListenPacket("udp", address)

	if err != nil {
		return nil, err
	}

	conn := newBoundPacketConn(pconn)

	return &conn, nil
}

func main() {
	log.SetFlags(0)

	listen := flag.Bool("l", false, "Listen")
	udp := flag.Bool("u", false, "Use UDP instead of TCP")

	flag.Parse()

	if flag.NArg() < 1 || flag.NArg() > 2 {
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

	go io.Copy(conn, os.Stdin)
	io.Copy(os.Stdout, conn)
}
