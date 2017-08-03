FROM golang:1.7-alpine3.6

#Installing imagemagick
RUN apk add --no-cache imagemagick git

#Installing godeps
RUN go get github.com/tools/godep

RUN mkdir -p /usr/local/go/src/github.com/dooman87/
WORKDIR /usr/local/go/src/github.com/dooman87/
RUN git clone https://github.com/dooman87/transformimgs.git

WORKDIR /usr/local/go/src/github.com/dooman87/transformimgs/
RUN godep restore

WORKDIR /usr/local/go/src/github.com/dooman87/transformimgs/cmd
ENTRYPOINT ["go", "run", "main.go", "-imConvert=/usr/bin/convert", "-imIdentify=/usr/bin/identify"]
