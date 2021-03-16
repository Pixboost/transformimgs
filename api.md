


# Images transformations API
The purpose of this API is to provide a set of
endpoints that will transform and optimise images.
Then it becomes easy to use the API with <img> and <picture> tags in web development.
  

## Informations

### Version

2

## Content negotiation

### URI Schemes
  * https

### Consumes
  * application/json

### Produces
  * image/jpeg
  * image/png

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
| GET | /api/2/img/{imgUrl}/fit | [fit image](#fit-image) | Resize image to the exact size and optimizes it. |
| GET | /api/2/img/{imgUrl}/optimise | [optimise image](#optimise-image) | Optimises image from the given url. |
| GET | /api/2/img/{imgUrl}/resize | [resize image](#resize-image) | Resize image with preserving aspect ratio and optimizes it. |
  


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
| imgUrl | `path` | string | `string` |  | ✓ |  | url of the image |

#### All responses
| Code | Status | Description | Has headers | Schema |
|------|--------|-------------|:-----------:|--------|
| [200](#asis-image-200) | OK | Requested image. |  | [schema](#asis-image-200-schema) |

#### Responses


##### <span id="asis-image-200"></span> 200 - Requested image.
Status: OK

###### <span id="asis-image-200-schema"></span> Schema

### <span id="fit-image"></span> Resize image to the exact size and optimizes it. (*fitImage*)

```
GET /api/2/img/{imgUrl}/fit
```

Will resize image and crop it to the size.
If you need to resize image with preserved aspect ratio then use /img/resize endpoint.

#### Produces
  * image/jpeg
  * image/png

#### Parameters

| Name | Source | Type | Go type | Separator | Required | Default | Description |
|------|--------|------|---------|-----------| :------: |---------|-------------|
| imgUrl | `path` | string | `string` |  | ✓ |  | url of the original image |
| size | `query` | string | `string` |  | ✓ |  | size of the image in the response. Should be in the format 'width'x'height', e.g. 200x300 |

#### All responses
| Code | Status | Description | Has headers | Schema |
|------|--------|-------------|:-----------:|--------|
| [200](#fit-image-200) | OK | Resized image in the same format as original. |  | [schema](#fit-image-200-schema) |

#### Responses


##### <span id="fit-image-200"></span> 200 - Resized image in the same format as original.
Status: OK

###### <span id="fit-image-200-schema"></span> Schema

### <span id="optimise-image"></span> Optimises image from the given url. (*optimiseImage*)

```
GET /api/2/img/{imgUrl}/optimise
```

#### Produces
  * image/jpeg
  * image/png

#### Parameters

| Name | Source | Type | Go type | Separator | Required | Default | Description |
|------|--------|------|---------|-----------| :------: |---------|-------------|
| imgUrl | `path` | string | `string` |  | ✓ |  | Url of the original image |
| save-data | `query` | string | `string` |  |  |  | Sets an optional behaviour when Save-Data header is "on".
When passing "off" value the result image won't use additional
compression when data saver mode is on.
When passing "hide" value the result image will be an empty image. |

#### All responses
| Code | Status | Description | Has headers | Schema |
|------|--------|-------------|:-----------:|--------|
| [200](#optimise-image-200) | OK | Optimised image in the same format as original. |  | [schema](#optimise-image-200-schema) |

#### Responses


##### <span id="optimise-image-200"></span> 200 - Optimised image in the same format as original.
Status: OK

###### <span id="optimise-image-200-schema"></span> Schema

### <span id="resize-image"></span> Resize image with preserving aspect ratio and optimizes it. (*resizeImage*)

```
GET /api/2/img/{imgUrl}/resize
```

If you need the exact size then use /fit operation.

#### Produces
  * image/jpeg
  * image/png

#### Parameters

| Name | Source | Type | Go type | Separator | Required | Default | Description |
|------|--------|------|---------|-----------| :------: |---------|-------------|
| imgUrl | `path` | string | `string` |  | ✓ |  | url of the original image |
| size | `query` | string | `string` |  | ✓ |  | size of the image in the response. Should be in format 'width'x'height', e.g. 200x300
Only width or height could be passed, e.g 200, x300. |

#### All responses
| Code | Status | Description | Has headers | Schema |
|------|--------|-------------|:-----------:|--------|
| [200](#resize-image-200) | OK | Resized image in the same format as original. |  | [schema](#resize-image-200-schema) |

#### Responses


##### <span id="resize-image-200"></span> 200 - Resized image in the same format as original.
Status: OK

###### <span id="resize-image-200-schema"></span> Schema

## Models
