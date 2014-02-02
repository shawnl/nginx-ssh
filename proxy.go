package main

import (
	"bytes"
	"fmt"
	"io"
	"net"
	"os"
	"syscall"
	"strconv"
	"strings"
)

func copyAndClose(w io.Writer, r io.ReadCloser) {
	io.Copy(w, r)
	r.Close()
}

func handleConnection(c net.Conn) {
	var err error
	var d net.Conn
	buf := make([]byte, 4096, 4096)
	oobbuf := make([]byte, 1024, 1024)

	defer c.Close()
	
	conn := c.(*net.TCPConn)
	f, _ := conn.File()
	n, _, _, from, err := syscall.Recvmsg(int(f.Fd()), buf, oobbuf, syscall.MSG_PEEK)
	if err != nil {
		fmt.Println(from, err)
		return
	}
	
	if bytes.Equal(buf, []byte{'S', 'S', 'H', '-'}) {
		d, err = net.Dial("tcp", "churchofgit.com:22")
	} else if bytes.Equal(buf[:2], []byte{0x16, 0x03}) && buf[3] >= 0x00 && buf[3] <= 0x03 {
		d, err = net.Dial("tcp", "churchofgit.com:443")
	} else if n == 5 {
	} else {
		fmt.Println(c.RemoteAddr, "Protocol not recognized")
		return
	}
	if err != nil {
		fmt.Println(c.RemoteAddr, err)
		return
	}
	go copyAndClose(c, d)
	io.Copy(d, c)
}

func parseHostPort(arg string) (host string, port int, err error) {
	if strings.Index(arg, ":") == -1 {
		host = "0.0.0.0"
		port, err = strconv.Atoi(arg)
		return
	}
	n, err := strconv.Atoi(arg[strings.Index(arg, ":") + 1:])
	return arg[:strings.Index(arg, ":")], n, err
}

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Program requires arguments")
		return
	}

	_, port, err := parseHostPort(os.Args[1])
	if err != nil {
		fmt.Println("bad Listen host:port", os.Args[1])
	}

	ln, err := net.Listen("tcp", fmt.Sprint(":", port))
	if err != nil {
		fmt.Println("Listen failed: ", err)
		os.Exit(1)
	}

	for {
		conn, err := ln.Accept()
		if err != nil {
			fmt.Println("Accept failed: ", err)
			continue
		}
		go handleConnection(conn)
	}
}
