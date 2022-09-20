FROM golang:1.13-alpine

WORKDIR /tmp

RUN apk add git --no-cache \
    && go get -v github.com/avulor/nginx-reloader


FROM nginx:1.23-alpine

COPY --from=0 /go/bin/nginx-reloader /usr/local/bin/

ENTRYPOINT ["nginx-reloader"]

CMD ["--cooldown", "3", "--watch", "/etc/nginx/conf.d", "--nginx-command", "nginx", "-g", "daemon off;"]
