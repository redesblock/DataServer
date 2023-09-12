ARG GO_VER=1.20
ARG ALPINE_VER=3.17

FROM alpine:${ALPINE_VER} as base
RUN apk add --no-cache tzdata \
    && cp /usr/share/zoneinfo/Asia/Shanghai /etc/localtime \
    && echo "Asia/Shanghai" > /etc/timezone \
    && apk del tzdata

FROM golang:${GO_VER}-alpine${ALPINE_VER} as golang
RUN apk add --no-cache \
	bash \
	gcc \
	g++ \
	make

WORKDIR /go/src/github.com/redesblock/DataServer
COPY . .
RUN go env -w GOPROXY=https://goproxy.cn,direct
RUN go get -u github.com/swaggo/swag/cmd/swag
RUN make

FROM base
WORKDIR /root
COPY --from=golang /go/src/github.com/redesblock/DataServer/build /usr/local/bin
COPY --from=golang /go/src/github.com/redesblock/DataServer/.dataserver.yaml .
CMD ["dataserver"]