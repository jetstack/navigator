FROM alpine:3.5

ADD navigator_linux_amd64 /usr/bin/navigator

CMD ["/usr/bin/navigator"]
