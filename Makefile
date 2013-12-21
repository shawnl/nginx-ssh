CFLAGS=-fstack-protector --param ssp-buffer-size=4 -fPIE -pie

spawn-ssh-suid:
	cc -Wall $(CFLAGS) -g spawn-ssh-suid.c -o spawn-ssh-suid

proxy:
	go build proxy.go

install: spawn-ssh-suid
	mkdir -p /usr/lib/nginx
	cp spawn-ssh-suid /usr/lib/nginx
	chown root:root /usr/lib/nginx/spawn-ssh-suid
	chmod +s /usr/lib/nginx/spawn-ssh-suid

clean:
	rm spawn-ssh-suid
