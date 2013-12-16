spawn-ssh-suid:
	cc -g spawn-ssh-suid.c -o spawn-ssh-suid

install: spawn-ssh-suid
	mkdir -p /usr/lib/nginx
	cp spawn-ssh-suid /usr/lib/nginx
	chown root:root /usr/lib/nginx/spawn-ssh-suid
	chmod +s /usr/lib/nginx/spawn-ssh-suid

clean:
	rm spawn-ssh-suid
