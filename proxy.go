package main

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"net"
	"os"
)

func copyAndClose(w io.Writer, r io.ReadCloser) {
	io.Copy(w, r)
	r.Close()
}

func handleConnection(c net.Conn) {
	defer c.Close()
	r := bufio.NewReader(c)
	buf, err := r.Peek(4)
	var d net.Conn
	if bytes.Equal(buf, []byte{'S', 'S', 'H', '-'}) {
		d, err = net.Dial("tcp", "localhost:22")
	} else if bytes.Equal(buf[:2], []byte{0x16, 0x03}) && buf[3] >= 0x00 && buf[3] <= 0x03 {
		d, err = net.Dial("tcp", "localhost:443")
	} else {
		fmt.Println(c.RemoteAddr, "Protocol not recognized")
		return
	}
	if err != nil {
		fmt.Println(c.RemoteAddr, err)
		return
	}
	go copyAndClose(c, d)
	io.Copy(d, r)
}

func main() {
	ln, err := net.Listen("tcp", ":443")
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
