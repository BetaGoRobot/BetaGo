set -e
current=`date "+%Y-%m-%d %H:%M:%S"`  
timeStamp=`date -d "$current" +%s`   
#将current转换为时间戳，精确到毫秒  
currentTimeStamp=$((timeStamp*1000+`date "+%N"`/1000000))


cd ../netease-api-service
docker build . -f ../scripts/Dockerfile -t kevinmatt/netease-api:latest

docker push kevinmatt/netease-api
docker tag kevinmatt/netease-api:latest kevinmatt/netease-api:$currentTimeStamp
docker push kevinmatt/netease-api:$currentTimeStamp