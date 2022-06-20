#!/bin/sh
registryURL=ccr.ccs.tencentyun.com
docker rm -f netease-api
docker run -d --rm --name netease-api -p 3335:3335 --network betago "$registryURL"/kevinmatt/netease-api