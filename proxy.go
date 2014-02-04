package main

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"net"
	"os"
	"regexp"
	"strconv"
	"strings"
	"syscall"
)

type matcher func(packet []byte, length int) (host string, port int)

func copyAndClose(w io.Writer, r io.ReadCloser) {
	io.Copy(w, r)
	r.Close()
}

func handleConnection(c net.Conn, patterns []matcher) {
	var err error
	var d net.Conn
	buf := make([]byte, 4096, 4096)
	oobbuf := make([]byte, 512, 512)

	defer c.Close()

	conn := c.(*net.TCPConn)
	f, _ := conn.File()
	length, _, _, from, err := syscall.Recvmsg(int(f.Fd()), buf, oobbuf, syscall.MSG_PEEK)
	if err != nil {
		fmt.Println(from, err)
		return
	}
	f.Close()

	for n := 0;;n += 1 {
		var host string
		var port int

		if len(patterns) == n {
			fmt.Println(c.RemoteAddr, "Protocol not recognized")
			return
		}

		host, port = patterns[n](buf, length)
		if port > 0 {
			d, err = net.Dial("tcp", fmt.Sprint(host, ":", port))
			break
		}
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
		host = "localhost"
		port, err = strconv.Atoi(arg)
		return
	}
	n, err := strconv.Atoi(arg[strings.Index(arg, ":")+1:])
	return arg[:strings.Index(arg, ":")], n, err
}

func main() {
	var patterns []matcher

	if len(os.Args) < 2 {
		fmt.Println("multiplexd [listenhost:]port [--ssl [host:]port|--ssh [host:]port|--openvpn [host:]port|--regex regex [host:]port..]")
		return
	}

	host, port, err := parseHostPort(os.Args[1])
	if err != nil {
		fmt.Println("Bad Listen host:port:", os.Args[1])
	}

	if bytes.Compare([]byte(host), []byte("localhost")) == 0 {
		host = "0.0.0.0"
	}

	ln, err := net.Listen("tcp", fmt.Sprint(host, ":", port))
	if err != nil {
		fmt.Println("Listen failed:", err)
		os.Exit(1)
	}

	for n := 2; n < len(os.Args)-1; n += 2 {
		if bytes.Equal([]byte(os.Args[n]), []byte("--regex")) {
			if len(os.Args) < n+2 {
				return
			}
			host, port, err := parseHostPort(os.Args[n+2])
			if err != nil {
				fmt.Println("Bad host:port specification:", os.Args[n+2], host, port, err)
				return
			}

			r, err := regexp.Compile(os.Args[n+1])
			if err != nil {
				fmt.Println("Failed to compile regular expression:", os.Args[n+1], err)
				return
			}

			patterns = append(patterns, (func(packet []byte, length int) (h string, p int) {
				h = host
				if r.Match(packet) {
					p = port
				}
				return
			}))

			n += 1
			continue
		}

		host, port, err := parseHostPort(os.Args[n+1])
		if err != nil {
			fmt.Println("Bad host:port specification:", os.Args[n+1], host, port, err)
			return
		}
		if bytes.Equal([]byte(os.Args[n]), []byte("--ssh")) {
			patterns = append(patterns, (func(packet []byte, length int) (h string, p int) {
				h = host
				if bytes.Equal(packet[:4], []byte("SSH-")) {
					p = port
				}
				return
			}))
		} else if bytes.Equal([]byte(os.Args[n]), []byte("--ssl")) {
			patterns = append(patterns, (func(pack []byte, length int) (h string, p int) {
				h = host
				if bytes.Equal(pack[:2], []byte{0x16, 0x03}) && pack[3] >= 0x00 && pack[3] <= 0x03 {
					p = port
				}
				return
			}))
		} else if bytes.Equal([]byte(os.Args[n]), []byte("--openvpn")) {
			patterns = append(patterns, (func(pack []byte, length int) (h string, p int) {
				var l uint16
				h = host
				binary.Read(bytes.NewReader(pack), binary.BigEndian, &l)
				if l == uint16(length-2) {
					p = port
				}
				return
			}))
		}
	}

	for {
		conn, err := ln.Accept()
		if err != nil {
			fmt.Println("Accept failed:", err)
			continue
		}
		go handleConnection(conn, patterns)
	}
}
