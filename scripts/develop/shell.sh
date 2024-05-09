docker run -d --name=loki -p=3100:3100 --network=betago  grafana/loki 
docker run -d --name=grafana --network=betago -p 3000:3000 grafana/grafana-enterprise
docker run -d --network=betago --name=prometheus -v /root/.config/promethus/prometheus.yml:/etc/prometheus/prometheus.yml -p 9091:9090 prom/prometheus