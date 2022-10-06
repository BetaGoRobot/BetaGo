set -xe
docker stop betago
docker rm betago
docker run --pull=always -d --name=betago --network=betago kevinmatt/betago:latest