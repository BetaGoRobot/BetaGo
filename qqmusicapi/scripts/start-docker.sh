#!/bin/sh
registryURL=ccr.ccs.tencentyun.com
docker rm -f qqmusic-api
docker run -d --rm --name qqmusic-api -p 3300:3300 --network betago "$registryURL"/kevinmatt/qqmusic-api