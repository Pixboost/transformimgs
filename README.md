# TransformImgs

[![Build Status](https://travis-ci.org/dooman87/transformimgs.svg?branch=master)](https://travis-ci.org/dooman87/transformimgs)
[![Docker Automated build](https://img.shields.io/docker/automated/jrottenberg/ffmpeg.svg)](https://hub.docker.com/r/dpokidov/transformimgs/)

Image transformations service.

The first iteration goal:

* Provide API to predefined operations implemented using [imagemagick](http://imagemagick.org) CLI.

## Table of Contents

- [Install](#install)
- [Usage](#usage)
- [API](#api)
- [Contribute](#contribute)
- [License](#license)

## Install

Using docker:

```
$ docker pull dpokidov/transformimgs
```

## Usage

```
$ docker run -p 8080:8080 dpokidov/transformimgs
```

To test that application started successfully:

`$ curl http://localhost:8080/health`

You should get 'OK' string in the response.

At the moment application provides 3 operations that accessible through HTTP:

* /img - optimises image
* /img/resize - resizes image
* /img/fit - resize image to the exact size by resizing and cropping it

Detailed API docs is here - http://docs.pixboost.com/api/index.html

### Running the application locally from sources

```
$ docker-compose up
```

### Building and Running from sources 

Dependencies:

* Go 1.7+
* [Gorilla MUX](https://github.com/gorilla/mux) for HTTP routing
* [kolibri](https://github.com/dooman87/kolibri) for healthcheck and testing
* Installed [imagemagick](http://imagemagick.org)

```
$ go get github.com/tools/godep
$ go get github.com/dooman87/transformimgs
$ cd $GOPATH/src/github.com/dooman87/transformimgs
$ godep restore
$ go run cmd/main.go -imConvert=/usr/bin/convert
```

## API

You can go through [API docs](http://docs.pixboost.com/api/index.html) and try it out there as well. Use 
API key `MTg4MjMxMzM3MA__` to use with any images from pixabay.com.

[Go-swagger](https://goswagger.io) is used to generate swagger.json file from sources. To generate:

```
$ go get -u github.com/go-swagger/go-swagger/cmd/swagger
$ cd cmd/
$ swagger generate spec -o ../swagger.json
```

## Contribute

Shout out with any ideas. PRs are more than welcome.

## License

[MIT](./LICENSE)