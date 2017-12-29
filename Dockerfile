FROM golang:1.8-alpine

#Installing imagemagick
RUN apk add --no-cache imagemagick git

#Installing godeps
RUN go get github.com/golang/dep/cmd/dep

RUN mkdir -p /usr/local/go/src/github.com/dooman87/
WORKDIR /usr/local/go/src/github.com/dooman87/
RUN git clone https://github.com/dooman87/transformimgs.git

WORKDIR /usr/local/go/src/github.com/dooman87/transformimgs/
RUN dep ensure

WORKDIR /usr/local/go/src/github.com/dooman87/transformimgs/cmd
ENTRYPOINT ["go", "run", "main.go", "-imConvert=/usr/bin/convert", "-imIdentify=/usr/bin/identify"]
