FROM golang:1.13-alpine as builder

ENV VERSION 1.0

COPY . /galaxy/app/reptile

RUN sed -i 's/dl-cdn.alpinelinux.org/mirrors.ustc.edu.cn/g' /etc/apk/repositories

RUN apk add --no-cache git gcc ca-certificates musl-dev linux-headers


ENV GOPROXY https://goproxy.cn,direct


FROM alpine:latest

RUN apk add --no-cache ca-certificates

RUN apk add --no-cache tzdata

WORKDIR /galaxy/app/reptile

#RUN go build  -ldflags "-w -s" -o /galaxy/app/reptile/build/spider  /galaxy/app/reptile/cmd/app/main.go
