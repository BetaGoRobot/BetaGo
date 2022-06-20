#!/bin/sh
set -e
current=$(date "+%Y-%m-%d %H:%M:%S") 
timeStamp=$(date -d "$current" +%s)  
#将current转换为时间戳，精确到毫秒  
currentTimeStamp=$((timeStamp*1000+$(date "+%N")/1000000))
registryURL=ccr.ccs.tencentyun.com

cd ../qqmusic-api-service
docker build . -f ../scripts/Dockerfile -t "$registryURL"/kevinmatt/qqmusic-api:latest

docker push "$registryURL"/kevinmatt/qqmusic-api
docker tag "$registryURL"/kevinmatt/qqmusic-api:latest "$registryURL"/kevinmatt/qqmusic-api:$currentTimeStamp
docker push "$registryURL"/kevinmatt/qqmusic-api:$currentTimeStamp