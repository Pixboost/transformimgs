


# Image transformations API
The main purpose of this is to help Web Developers to serve
images in the best possible way meaning balance between
quality and speed.

Each endpoint could be used directly in `<img>` and `<picture>` HTML tags
  

## Informations

### Version

2.1

## Content negotiation

### URI Schemes
  * https

### Consumes
  * application/json

### Produces
  * image/avif
  * image/jpeg
  * image/png
  * image/webp

## Access control

### Security Schemes

#### api_key (query: auth)



> **Type**: apikey

### Security Requirements
  * api_key

## All endpoints

###  images

| Method  | URI     | Name   | Summary |
|---------|---------|--------|---------|
| GET | /api/2/img/{imgUrl}/asis | [asis image](#asis-image) |  |
| GET | /api/2/img/{imgUrl}/fit | [fit image](#fit-image) | Resizes, crops, and optimises an image to the exact size. |
| GET | /api/2/img/{imgUrl}/optimise | [optimise image](#optimise-image) | Optimises image from the given url. |
| GET | /api/2/img/{imgUrl}/resize | [resize image](#resize-image) | Resizes, optimises image and preserve aspect ratio. |
  


## Paths

### <span id="asis-image"></span> asis image (*asisImage*)

```
GET /api/2/img/{imgUrl}/asis
```

Respond with original image without any modifications

#### Produces
  * image/jpeg
  * image/png

#### Parameters

| Name | Source | Type | Go type | Separator | Required | Default | Description |
|------|--------|------|---------|-----------| :------: |---------|-------------|
| imgUrl | `path` | string | `string` |  | ✓ |  | Url of the original image including schema. Note that query parameters need to be properly encoded |

#### All responses
| Code | Status | Description | Has headers | Schema |
|------|--------|-------------|:-----------:|--------|
| [200](#asis-image-200) | OK | Requested image. |  | [schema](#asis-image-200-schema) |

#### Responses


##### <span id="asis-image-200"></span> 200 - Requested image.
Status: OK

###### <span id="asis-image-200-schema"></span> Schema

### <span id="fit-image"></span> Resizes, crops, and optimises an image to the exact size. (*fitImage*)

```
GET /api/2/img/{imgUrl}/fit
```

If you need to resize image with preserved aspect ratio then use /resize endpoint.

#### Produces
  * image/avif
  * image/jpeg
  * image/png
  * image/webp

#### Parameters

| Name | Source | Type | Go type | Separator | Required | Default | Description |
|------|--------|------|---------|-----------| :------: |---------|-------------|
| imgUrl | `path` | string | `string` |  | ✓ |  | Url of the original image including schema. Note that query parameters need to be properly encoded |
| dppx | `query` | float (formatted number) | `float32` |  |  | `1` | Number of dots per pixel defines the ratio between device and CSS pixels. The query parameter is a hint that enables extra optimisations for high density screens. The format is a float number that's the same format as window.devicePixelRatio. |
| save-data | `query` | string | `string` |  |  |  | Sets an optional behaviour when Save-Data header is "on". When passing "off" value the result image won't use extra compression when data saver mode is on. When passing "hide" value the result image will be an empty 1x1 image. When absent the API will use reduced quality for result images. |
| size | `query` | string | `string` |  | ✓ |  | size of the image in the response. Should be in the format 'width'x'height', e.g. 200x300 |

#### All responses
| Code | Status | Description | Has headers | Schema |
|------|--------|-------------|:-----------:|--------|
| [200](#fit-image-200) | OK | Resized image |  | [schema](#fit-image-200-schema) |

#### Responses


##### <span id="fit-image-200"></span> 200 - Resized image
Status: OK

###### <span id="fit-image-200-schema"></span> Schema

### <span id="optimise-image"></span> Optimises image from the given url. (*optimiseImage*)

```
GET /api/2/img/{imgUrl}/optimise
```

#### Produces
  * image/avif
  * image/jpeg
  * image/png
  * image/webp

#### Parameters

| Name | Source | Type | Go type | Separator | Required | Default | Description |
|------|--------|------|---------|-----------| :------: |---------|-------------|
| imgUrl | `path` | string | `string` |  | ✓ |  | Url of the original image including schema. Note that query parameters need to be properly encoded |
| dppx | `query` | float (formatted number) | `float32` |  |  | `1` | Number of dots per pixel defines the ratio between device and CSS pixels. The query parameter is a hint that enables extra optimisations for high density screens. The format is a float number that's the same format as window.devicePixelRatio. |
| save-data | `query` | string | `string` |  |  |  | Sets an optional behaviour when Save-Data header is "on". When passing "off" value the result image won't use extra compression when data saver mode is on. When passing "hide" value the result image will be an empty 1x1 image. When absent the API will use reduced quality for result images. |

#### All responses
| Code | Status | Description | Has headers | Schema |
|------|--------|-------------|:-----------:|--------|
| [200](#optimise-image-200) | OK | Optimised image. |  | [schema](#optimise-image-200-schema) |

#### Responses


##### <span id="optimise-image-200"></span> 200 - Optimised image.
Status: OK

###### <span id="optimise-image-200-schema"></span> Schema

### <span id="resize-image"></span> Resizes, optimises image and preserve aspect ratio. (*resizeImage*)

```
GET /api/2/img/{imgUrl}/resize
```

Use /fit operation for resizing to the exact size.

#### Produces
  * image/avif
  * image/jpeg
  * image/png
  * image/webp

#### Parameters

| Name | Source | Type | Go type | Separator | Required | Default | Description |
|------|--------|------|---------|-----------| :------: |---------|-------------|
| imgUrl | `path` | string | `string` |  | ✓ |  | Url of the original image including schema. Note that query parameters need to be properly encoded |
| dppx | `query` | float (formatted number) | `float32` |  |  | `1` | Number of dots per pixel defines the ratio between device and CSS pixels. The query parameter is a hint that enables extra optimisations for high density screens. The format is a float number that's the same format as window.devicePixelRatio. |
| save-data | `query` | string | `string` |  |  |  | Sets an optional behaviour when Save-Data header is "on". When passing "off" value the result image won't use extra compression when data saver mode is on. When passing "hide" value the result image will be an empty 1x1 image. When absent the API will use reduced quality for result images. |
| size | `query` | string | `string` |  | ✓ |  | Size of the result image. Should be in the format 'width'x'height', e.g. 200x300
Only width or height could be passed, e.g 200, x300. |

#### All responses
| Code | Status | Description | Has headers | Schema |
|------|--------|-------------|:-----------:|--------|
| [200](#resize-image-200) | OK | Resized image. |  | [schema](#resize-image-200-schema) |

#### Responses


##### <span id="resize-image-200"></span> 200 - Resized image.
Status: OK

###### <span id="resize-image-200-schema"></span> Schema

## Models
