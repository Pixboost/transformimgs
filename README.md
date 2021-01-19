# TransformImgs

[![Build Status](https://travis-ci.org/Pixboost/transformimgs.svg?branch=master)](https://travis-ci.org/Pixboost/transformimgs)
[![codecov](https://codecov.io/gh/Pixboost/transformimgs/branch/master/graph/badge.svg)](https://codecov.io/gh/Pixboost/transformimgs)
[![Docker Pulls](https://img.shields.io/docker/pulls/pixboost/transformimgs)](https://hub.docker.com/r/pixboost/transformimgs/)
[![Docker Automated build](https://img.shields.io/docker/automated/jrottenberg/ffmpeg.svg)](https://hub.docker.com/r/pixboost/transformimgs/)

Open Source [Image CDN](https://web.dev/image-cdns/) that provides image transformation API and supports 
the latest image formats, such as WebP, AVIF. 

There are two ways of using the service:

* Deploy on your own infrastructure using docker image
* Use as SaaS at [pixboost.com](https://pixboost.com?source=github)

## Table of Contents

- [Features](#features)
- [Install](#install)
- [Usage](#usage)
- [API](#api)
- [Contribute](#contribute)
- [License](#license)
- [TODO](#todo)

## Features

* Resize/optimises/crops raster (PNG and JPEG) images.
* [AVIF](https://en.wikipedia.org/wiki/AV1)/[WebP](https://developers.google.com/speed/webp/) support based on "Accept" header.
* Sets "[Cache-Control](www.w3.org/Protocols/rfc2616/rfc2616-sec14.html#sec14.9)" header in a response. 
    Cache TTL is configurable through command line flag "-cache".
* Execution queue that will create number of executors based on number of CPUs or can be configured through "-proc" flag.
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

Prerequisites:

* Go with [modules support](https://golang.org/ref/mod)
* Installed [imagemagick v7.0.25+](http://imagemagick.org) with AVIF support. Run script assumes that binaries are in `/usr/local/bin`

```
$ git clone git@github.com:Pixboost/transformimgs.git
$ cd transformimgs
$ ./run.sh 
```

Go modules have been introduced in v6.

### Performance tests

There is a [JMeter](https://jmeter.apache.org) performance test you can run against a service. To run tests:

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

* Run JMeter AVIF test:
```
$ jmeter -n -t perf-test-avif.jmx -l ./results-avif.jmx -e -o ./results-avif
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
* ~~Add JpegXR support~~ (IE supports WEBP)
* ~~Add Jpeg 2000 support~~ (Safari support WEBP)
* Client Hints
* [Save-Data header](https://developer.mozilla.org/en-US/docs/Web/HTTP/Headers/Save-Data)
* SVG support
* Consider using [Zopfli](https://github.com/google/zopfli) or [Brotli](https://en.wikipedia.org/wiki/Brotli) for PNGs
* ~~GIF support~~ (Added in version 6.1.0)
