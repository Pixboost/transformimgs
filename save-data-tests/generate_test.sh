#!/usr/bin/env bash

URL="http://localhost:8080/img/https://pixboost.com/img/homepage/hero.jpg/resize?size=x600"

curl -H "Accept: image/webp"  "${URL}" -o ./original.webp
curl -H "Accept: image/avif"  "${URL}" -o ./original.avif
curl -H "Accept: image/jpeg"  "${URL}" -o ./original.jpg

curl -H "Accept: image/webp" -H "Save-Data: on"  "${URL}" -o ./savedata.webp
curl -H "Accept: image/avif" -H "Save-Data: on"  "${URL}" -o ./savedata.avif
curl -H "Accept: image/jpg" -H "Save-Data: on"  "${URL}" -o ./savedata.jpg
