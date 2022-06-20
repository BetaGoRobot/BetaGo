#!/bin/bash
set -e
docker build . -t kevinmatt/betago-dev
docker push kevinmatt/betago-dev