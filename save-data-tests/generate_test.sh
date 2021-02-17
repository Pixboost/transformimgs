#!/usr/bin/env bash

curl -H "Accept: image/webp"  "http://localhost:8080/img/https://pixboost.com/img/homepage/hero.jpg/resize?size=x600" -o ./original.webp
curl -H "Accept: image/avif"  "http://localhost:8080/img/https://pixboost.com/img/homepage/hero.jpg/resize?size=x600" -o ./original.avif
curl -H "Accept: image/jpeg"  "http://localhost:8080/img/https://pixboost.com/img/homepage/hero.jpg/resize?size=x600" -o ./original.jpg

curl -H "Accept: image/webp" -H "Save-Data: on"  "http://localhost:8080/img/https://pixboost.com/img/homepage/hero.jpg/resize?size=x600" -o ./savedata.webp
curl -H "Accept: image/avif" -H "Save-Data: on"  "http://localhost:8080/img/https://pixboost.com/img/homepage/hero.jpg/resize?size=x600" -o ./savedata.avif
curl -H "Accept: image/jpg" -H "Save-Data: on"  "http://localhost:8080/img/https://pixboost.com/img/homepage/hero.jpg/resize?size=x600" -o ./savedata.jpg
