# TransformImgs

[![Build Status](https://travis-ci.org/Pixboost/transformimgs.svg?branch=master)](https://travis-ci.org/Pixboost/transformimgs)
[![Docker Automated build](https://img.shields.io/docker/automated/jrottenberg/ffmpeg.svg)](https://hub.docker.com/r/dpokidov/transformimgs/)

Image transformations web service. Provides Http API to image 
manipulation operations backed by [Imagemagick](http://imagemagick.org) CLI.

## Table of Contents

- [Install](#install)
- [Usage](#usage)
- [API](#api)
- [Contribute](#contribute)
- [License](#license)
- [TODO](#todo)

## Install

Using docker:

```
$ docker pull pixboost/transformimgs
```

## Usage

```
$ docker run -p 8080:8080 pixboost/transformimgs
```

To test that application started successfully:

`$ curl http://localhost:8080/health`

You should get 'OK' string in the response.

At the moment application provides 4 operations that accessible through HTTP endpoints:

* /img/{IMG_URL}/optimise - optimises image
* /img/{IMG_URL}/resize - resizes image
* /img/{IMG_URL}/fit - resize image to the exact size by resizing and cropping it
* /img/{IMG_URL}/asis - returns original image

Detailed API docs are here - http://docs.pixboost.com/api/index.html

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

### Perfomance tests

There is a JMeter performance test that you can run against a service. To make it run
* Run performance test environment:
```
$ docker-compose -f docker-compose-perf-test.yml up
```
* Run JMeter test:
```
$ jmeter -n -t perf-test.jmx -l ./results.jmx -e -o ./results
```

## API

You can go through [API docs](https://pixboost.com/docs/api/index.html) and try it out there as well. Use 
API key `MTg4MjMxMzM3MA__` to transform any images from pixabay.com.

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

## Todo

* Consider using [Zopfli](https://github.com/google/zopfli) for PNGs
