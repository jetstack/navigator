FROM alpine:3.5

ADD colonel_linux_amd64 /usr/bin/colonel

CMD ["/usr/bin/colonel"]
