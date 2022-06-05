set -e
current=`date "+%Y-%m-%d %H:%M:%S"`  
timeStamp=`date -d "$current" +%s`   
#将current转换为时间戳，精确到毫秒  
currentTimeStamp=$((timeStamp*1000+`date "+%N"`/1000000))

cd ../qqmusic-api-service
docker build . -f ../scripts/Dockerfile -t kevinmatt/qqmusic-api:latest

docker push kevinmatt/qqmusic-api
docker tag kevinmatt/qqmusic-api:latest kevinmatt/qqmusic-api:$currentTimeStamp
docker push kevinmatt/qqmusic-api:$currentTimeStamp