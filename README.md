Multiplex's SSH onto nginx's SSL ports.

This allows you to listen for SSH on port 443
to evade firewalls, and is similar to [sslh](https://github.com/yrutschle/sslh)
except it is more performant.

For a tutorial see https://www.churchofgit.com/wordpress/ssh-and-https-on-the-same-port/

TODO
----

The nginx patch, as long as my helper binary could be much better integrated
with nginx, including using a proper feature macro with configure-time
selection, having a directive to listen in the nginx config file (like spdy),
and integrating the helper binary into the nginx build system. Patches welcome!
