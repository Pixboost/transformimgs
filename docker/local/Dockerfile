FROM golang:1.5.3-alpine

#Installing imagemagick
RUN apk add --no-cache imagemagick git

#Installing godeps
ENV GO15VENDOREXPERIMENT=1
RUN go get github.com/tools/godep

VOLUME /usr/local/go/src/github.com/dooman87/transformimgs/
WORKDIR /usr/local/go/src/github.com/dooman87/transformimgs/
ENTRYPOINT ["sh", "/usr/local/go/src/github.com/dooman87/transformimgs/run.sh"]
