#include <unistd.h>
#include <sys/types.h>
#include <sys/stat.h>
#include <fcntl.h>
#include <stdlib.h>
#include <sys/socket.h>
#include <arpa/inet.h>
#include <sys/un.h>
#include <errno.h>
#include <pwd.h>

#define USER "www-data"
#define PORT 443

#ifndef static_assert
# define static_assert(expr, foo) do { char x[(expr) ? 0 : -1]; } while (0);
#endif

int main(int argc, char *argv[]) {
    int fd, r;
    socklen_t l;
    union {
            struct sockaddr sa;
            struct sockaddr_un un;
            struct sockaddr_in in;
            struct sockaddr_in6 in6;
            struct sockaddr_storage storage;
    } local, remote;
    pid_t ppid;
    struct stat st;
    struct passwd *pw;

    static_assert(PORT < 1024, "Allowing local port to be >= 1024 allows non-root"
                               "programs to spawn sshd thorugh this suid helper");

    errno = 0;
    pw = getpwnam(USER);
    if (errno > 0 || !pw)
        goto fail;

    if (pw->pw_uid != getuid())
        goto fail;

    fd = open("/dev/null", 0);
    if (fd < 0)
        goto fail;

    if (dup2(fd, STDERR_FILENO) < 0)
        goto fail;

    if (clearenv() < 0)
        goto fail;

    fd = 3;
    l = sizeof(local);
    if (getsockname(fd, &local.sa, &l) < 0)
        goto fail;

    l = sizeof(remote);
    if (getpeername(fd, &remote.sa, &l) < 0)
        goto fail;

    switch (local.sa.sa_family) {
    case AF_INET:
        if (local.in.sin_port != PORT)
            goto fail;
        break;
    case AF_INET6:
        if (local.in6.sin6_port != PORT)
            goto fail;
        break;
    case AF_UNIX:
        break;
    default:
        goto fail;
    }

    if (dup2(fd, STDOUT_FILENO) < 0)
        goto fail;
    if (dup2(fd, STDIN_FILENO) < 0)
        goto fail;
    r = close(fd);
    if (r < 0 && errno != EINTR)
        goto fail;

    execve("/usr/sbin/sshd", (char *[]){"sshd", "-i", NULL}, (char *[]){NULL});
fail:
    return EXIT_FAILURE;
}
