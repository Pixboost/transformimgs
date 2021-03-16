#!/usr/bin/env bash

#URL="http://localhost:8080/img/https://pixboost.com/img/homepage/hero.jpg/resize?size=x600"
#URL="http://localhost:8080/img/https:%2F%2Fcdn.shopify.com%2Fs%2Ffiles%2F1%2F0030%2F1128%2F0994%2Ffiles%2Fsock_up_your_life_1000x.jpg%3Fv%3D1609967932/optimise?auth=MTkwMDI1ODQ5NQ__"
#URL="http://localhost:8080/img/https://images.unsplash.com/photo-1612691626803-08cdafc05fb9/resize?size=x600"

#<span>Photo by <a href="https://unsplash.com/@lh1me?utm_source=unsplash&amp;utm_medium=referral&amp;utm_content=creditCopyText">HONG LIN</a> on <a href="https://unsplash.com/?utm_source=unsplash&amp;utm_medium=referral&amp;utm_content=creditCopyText">Unsplash</a></span>
URL="http://localhost:8080/img/https://images.unsplash.com/photo-1529978755210-7f13333beb13/resize?size=x600"
curl -H "Accept: image/webp"  "${URL}" -o ./original.webp
curl -H "Accept: image/avif"  "${URL}" -o ./original.avif
curl -H "Accept: image/jpeg"  "${URL}" -o ./original.jpg

curl -H "Accept: image/webp" -H "Save-Data: on"  "${URL}" -o ./savedata.webp
curl -H "Accept: image/avif" -H "Save-Data: on"  "${URL}" -o ./savedata.avif
curl -H "Accept: image/jpg" -H "Save-Data: on"  "${URL}" -o ./savedata.jpg
