# TransformImgs

[![Build Status](https://travis-ci.org/Pixboost/transformimgs.svg?branch=master)](https://travis-ci.org/Pixboost/transformimgs)
[![Docker Automated build](https://img.shields.io/docker/automated/jrottenberg/ffmpeg.svg)](https://hub.docker.com/r/pixboost/transformimgs/)

Image transformations web service. Provides Http API to image 
manipulation operations backed by [Imagemagick](http://imagemagick.org) CLI.

## Table of Contents

- [Features](#features)
- [Install](#install)
- [Usage](#usage)
- [API](#api)
- [Contribute](#contribute)
- [License](#license)
- [TODO](#todo)

## Features

* Resizes and optimises raster (PNG and JPEG) images.
* Sets "[Cache-Control](www.w3.org/Protocols/rfc2616/rfc2616-sec14.html#sec14.9)" header in a response. 
    Cache TTL is configurable through command line flag "-cache".
* Execution queue that will create number of executors based number of CPUs or can be configured through "-proc" flag.
* Webp support based on "Accept" header.
* Supports "[Vary](www.w3.org/Protocols/rfc2616/rfc2616-sec14.html#sec14.44)" header to cache responses based on the output format.

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

At the moment application provides 4 HTTP endpoints:

* /img/{IMG_URL}/optimise - optimises image
* /img/{IMG_URL}/resize - resizes image
* /img/{IMG_URL}/fit - resize image to the exact size by resizing and cropping it
* /img/{IMG_URL}/asis - returns original image

Detailed API docs are here - https://pixboost.com/docs/api/

### Running the application locally from sources

```
$ docker-compose up
```

### Building and Running from sources 

Dependencies:

* Go 1.8+
* [Gorilla MUX](https://github.com/gorilla/mux) for HTTP routing
* [kolibri](https://github.com/dooman87/kolibri) for healthcheck and testing
* [glogi](https://github.com/dooman87/glogi) for logging interface
* Installed [imagemagick](http://imagemagick.org)

```
$ go get github.com/golang/dep/cmd/dep
$ go get github.com/Pixboost/transformimgs
$ cd $GOPATH/src/github.com/Pixboost/transformimgs
$ ./run.sh 
```

### Perfomance tests

There is a JMeter performance test that you can run against a service. To make it run
* Run performance test environment:
```
$ docker-compose -f docker-compose-perf.yml up
```
* Run JMeter test:
```
$ jmeter -n -t perf-test.jmx -l ./results.jmx -e -o ./results
```

* Run JMeter WebP test:
```
$ jmeter -n -t perf-test-webp.jmx -l ./results-webp.jmx -e -o ./results-webp
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
* Add JpegXR support
* Add Jpeg 2000 support
* Consider using [Zopfli](https://github.com/google/zopfli) for PNGs
* Add SVG support
