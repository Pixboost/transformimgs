FROM golang:1.16.7-buster AS build

RUN mkdir -p /go/src/github.com/Pixboost/
WORKDIR /go/src/github.com/Pixboost/
RUN git clone https://github.com/Pixboost/transformimgs.git

WORKDIR /go/src/github.com/Pixboost/transformimgs/
RUN go mod vendor

WORKDIR /go/src/github.com/Pixboost/transformimgs/cmd

RUN go build -o /transformimgs

FROM dpokidov/imagemagick:7.1.0-10-buster

ENV IM_HOME /usr/local/bin

COPY --from=build /transformimgs /transformimgs

ENTRYPOINT ["/transformimgs", "-imConvert=/usr/local/bin/convert", "-imIdentify=/usr/local/bin/identify"]
