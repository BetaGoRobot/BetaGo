docker rm -f qqmusic-api
docker run -d --rm --name qqmusic-api -p 3300:3300 --network betago kevinmatt/qqmusic-api