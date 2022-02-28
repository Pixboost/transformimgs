# TransformImgs

[![Build Status](https://travis-ci.org/Pixboost/transformimgs.svg?branch=master)](https://travis-ci.org/Pixboost/transformimgs)
[![codecov](https://codecov.io/gh/Pixboost/transformimgs/branch/master/graph/badge.svg)](https://codecov.io/gh/Pixboost/transformimgs)
[![Docker Pulls](https://img.shields.io/docker/pulls/pixboost/transformimgs)](https://hub.docker.com/r/pixboost/transformimgs/)
[![Docker Automated build](https://img.shields.io/docker/automated/jrottenberg/ffmpeg.svg)](https://hub.docker.com/r/pixboost/transformimgs/)

Open Source [Image CDN](https://web.dev/image-cdns/) that provides image transformation API and supports 
the latest image formats, such as WebP, AVIF and network client hints. 

## Table of Contents

<!-- TOC start -->
- [Why?](#why)
- [Features](#features)
- [Quickstart](#quickstart)
- [API](#api)
- [Running](#running)
  * [Options](#options)
- [SaaS](#saas)
- [Building from sources ](#building-from-sources)
- [Performance tests](#performance-tests)
- [Opened tickets for images related features](#opened-tickets-for-images-related-features)
- [Contribute](#contribute)
- [License](#license)
- [Todo](#todo)
<!-- TOC end -->

## Why?

[We wrote a big blog on this](https://pixboost.com/blog/why-pixboost-is-the-best-image-cdn/), and here is TLDR:

Transformimgs is an image CDN for Web, so API must cover typical use cases, like
thumbnails, zoom in product images, etc. Any new API endpoints must 
solve the above problems.

The goal is to have zero-config API that makes decisions based on the input, so you don't need to provide additional parameters like quality, output format, type of compression, etc.

Therefore, this allows you to configure the integration once and new features, like new image formats, will work
with your front end without any changes.

## Features

* Resize/optimises/crops raster (PNG and JPEG) images.
* [AVIF](https://en.wikipedia.org/wiki/AV1) / [WebP](https://developers.google.com/speed/webp/) support based on "Accept" header.
* [Vary](www.w3.org/Protocols/rfc2616/rfc2616-sec14.html#sec14.44) header support - ready to deploy behind any CDN.
* Responsive images support including high DPI (retina) displays 

## Quickstart

There is an example of running API behind reverse proxy with integration example in `quickstart/` folder.

To run:

```
cd quickstart
docker-compose up
open https://localhost
```

## API

The API has 4 HTTP endpoints:

* /img/{IMG_URL}/optimise - optimises image
* /img/{IMG_URL}/resize - resizes image
* /img/{IMG_URL}/fit - resize image to the exact size by resizing and cropping it
* /img/{IMG_URL}/asis - returns original image

Docs:
* [Swagger-UI](https://pixboost.com/docs/api/) - use API key `MTg4MjMxMzM3MA__` which allows to transform any images from pixabay.com
* [Markdown API docs](api.md)

## Running

The latest docker image published on [Docker hub](https://hub.docker.com/r/pixboost/transformimgs)

Running the server:

```
$ docker run -p 8080:8080 pixboost/transformimgs [OPTIONS]
```

Test it:

* Health check: `curl http://localhost:8080/health`
* Transformation: `open http://localhost:8080/img/https://images.unsplash.com/photo-1591769225440-811ad7d6eab3/resize?size=600`

### Options

Everything below is optional and have sensible defaults.

| Option | Description |
|--------|-------------|
| cache  | Number of seconds to cache image(0 to disable cache). Default value is 2592000 (30 days) |
| proc   | Number of images processors to run. Defaults to number of CPUs |
| disableSaveData | If set to true then will disable Save-Data client hint. Could be useful for CDNs that don't support Save-Data header in Vary. |


## SaaS

We run SaaS version at [pixboost.com](https://pixboost.com?source=github) with generous free tier.

Perks of SaaS version:
* CDN with HTTP/3 support included
* Dashboard with usage monitor
* API Key support with domains allow list
* AWS S3 integration
* API workflows for cache busting and warmup
* Version upgrades

## Building from sources 

Prerequisites:

* Go 1.15+ with [modules support](https://golang.org/ref/mod)
* Installed [imagemagick v7.0.25+](http://imagemagick.org) with AVIF support in `/usr/local/bin`

```
$ git clone git@github.com:Pixboost/transformimgs.git
$ cd transformimgs
$ ./run.sh 
```

Go modules have been introduced in v6.

## Performance tests

There is a [JMeter](https://jmeter.apache.org) performance test you can run against a service. To run tests:

* Start a performance test environment:
```
$ docker-compose -f docker-compose-perf.yml up
```
* Run JMeter tests:
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


## Opened tickets for images related features

* [Safari to support Save-Data](https://bugs.webkit.org/show_bug.cgi?id=199101)
* [Safari to support AVIF](https://bugs.webkit.org/show_bug.cgi?id=207750)
* [Firefox to support JPEG XL](https://bugzilla.mozilla.org/show_bug.cgi?id=1539075)
* [Chrome to support JPEG XL](https://bugs.chromium.org/p/chromium/issues/detail?id=1178058)
* [Safari to support JPEG XL](https://bugs.webkit.org/show_bug.cgi?id=208235)
* Safari to support native lazy loading
  * [Implementation](https://bugs.webkit.org/show_bug.cgi?id=196698)
  * [Enabled by default](https://bugs.webkit.org/show_bug.cgi?id=208094)

## Contribute

Shout out with any ideas. PRs are more than welcome.

## License

[MIT](./LICENSE)

## Todo
* ~~Add JpegXR support~~ (IE supports WEBP)
* ~~Add Jpeg 2000 support~~ (Safari support WEBP)
* [Client Hints](https://github.com/Pixboost/transformimgs/issues/26) - on hold due to browsers adoption
* ~~[Save-Data header](https://github.com/Pixboost/transformimgs/issues/27)~~ (Added in version 7.0.0)
* [SVG support](https://github.com/Pixboost/transformimgs/issues/12)
* Consider using [Zopfli](https://github.com/google/zopfli) or [Brotli](https://en.wikipedia.org/wiki/Brotli) for PNGs
* ~~GIF support~~ (Added in version 6.1.0)
