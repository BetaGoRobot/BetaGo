#!/bin/sh
set -e
current=$(date "+%Y-%m-%d %H:%M:%S")
timeStamp=$(date -d "$current" +%s)
#将current转换为时间戳，精确到毫秒  
currentTimeStamp=$((timeStamp*1000+$(date "+%N")/1000000))
registryURL=ccr.ccs.tencentyun.com

cd ../netease-api-service
docker build . -f ../scripts/Dockerfile -t "$registryURL"/kevinmatt/netease-api:latest

docker push "$registryURL"/kevinmatt/netease-api
docker tag "$registryURL"/kevinmatt/netease-api:latest "$registryURL"/kevinmatt/netease-api:$currentTimeStamp
docker push "$registryURL"/kevinmatt/netease-api:$currentTimeStamp