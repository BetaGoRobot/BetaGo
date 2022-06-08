docker rm -f netease-api
docker run -d --rm --name netease-api -p 3335:3335 --network betago kevinmatt/netease-api