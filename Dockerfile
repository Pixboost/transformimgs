FROM golang:1.14-buster AS build

#Installing godeps
RUN go get github.com/golang/dep/cmd/dep

RUN mkdir -p /go/src/github.com/Pixboost/
WORKDIR /go/src/github.com/Pixboost/
RUN git clone https://github.com/Pixboost/transformimgs.git

WORKDIR /go/src/github.com/Pixboost/transformimgs/
RUN dep ensure

WORKDIR /go/src/github.com/Pixboost/transformimgs/cmd

RUN go build -o /transformimgs

FROM dpokidov/imagemagick:latest-buster

ENV IM_HOME /usr/local/bin

COPY --from=build /transformimgs /transformimgs

ENTRYPOINT ["/transformimgs", "-imConvert=/usr/local/bin/convert", "-imIdentify=/usr/local/bin/identify"]
