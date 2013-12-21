package main

import (
	"os"
	"os/user"
	"strconv"
	"syscall"
)

const (
	from_user string = "www-data"
	fd        int    = 3
)

func main() {
	var e error
	u := new(user.User)
	u, e = user.Lookup(from_user)
	if e != nil {
		os.Exit(1)
	}

	uid, e := strconv.Atoi(u.Uid)
	if e != nil {
		os.Exit(1)
	}

	if os.Getuid() != uid {
		os.Exit(1)
	}

	// doesn't work cause syscall.Sockaddr type
	// can't be used in any way.
	//syscall.Getsockaddr(fd)
	e = syscall.Dup2(fd, syscall.Stdin)
	if e != nil {
		os.Exit(1)
	}
	e = syscall.Dup2(fd, syscall.Stdout)
	if e != nil {
		os.Exit(1)
	}
	e = syscall.Dup2(fd, syscall.Stderr)
	if e != nil {
		os.Exit(1)
	}
	e = syscall.Close(fd)
	if e != nil {
		os.Exit(1)
	}

	syscall.Setuid(0)
	_ = syscall.Exec("/usr/sbin/sshd", []string{"sshd", "-i"}, []string{})
	os.Exit(1)
}
