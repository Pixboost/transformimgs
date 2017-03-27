# TransformImgs

[![Build Status](https://travis-ci.org/dooman87/transformimgs.svg?branch=master)](https://travis-ci.org/dooman87/transformimgs)

Image transformations service.

The first iteration goal:

* Provide HTTP endpoints to predefined operations implemented using [imagemagick](http://imagemagick.org) CLI.

# Requirements

* Go 1.7+
* [Gorilla MUX](https://github.com/gorilla/mux) for HTTP routing
* [kolibri](https://github.com/dooman87/kolibri) for healthcheck and testing
* Installed [imagemagick](http://imagemagick.org)

# Running

```
$ docker run -p 8080:8080 dpokidov/transformimgs
```

To test that application started successfully:

`http://localhost:8080/health`

You should get 'OK' string in the response.

## Running the application locally ##

```
$ docker-compose up
```

## Building and running locally ##
```
$ go get github.com/tools/godep
$ go get github.com/dooman87/transformimgs
$ cd $GOPATH/src/github.com/dooman87/transformimgs
$ godep restore
$ go run cmd/main.go -logtostderr=true -imConvert=/usr/bin/convert
```