# TransformImgs

[![Build Status](https://travis-ci.org/dooman87/transformimgs.svg?branch=master)](https://travis-ci.org/dooman87/transformimgs)

Image transformations service.

The first iteration goal:

* Provide HTTP endpoints to predefined operations implemented using [imagemagick](http://imagemagick.org) CLI.

# Usage

You can explore [API docs](http://docs.pixboost.com/api/index.html). Authorize using 
API key `MTg4MjMxMzM3MA__` and then you'll be able to use any images from pixabay.com.

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

## Running the application locally from sources##

```
$ docker-compose up
```

## Building and Running from sources ##
```
$ go get github.com/tools/godep
$ go get github.com/dooman87/transformimgs
$ cd $GOPATH/src/github.com/dooman87/transformimgs
$ godep restore
$ go run cmd/main.go -logtostderr=true -imConvert=/usr/bin/convert
```

# Swagger docs generation #

[Go-swagger](https://goswagger.io) is used to generate swagger.json file from comments. To generate:

```
$ go get -u github.com/go-swagger/go-swagger/cmd/swagger
$ cd cmd/
$ swagger generate spec -o ../swagger.json
```