#!/bin/bash

# Start the proxied server
docker stop test-nginx
docker rm test-nginx
docker run --name test-nginx \
-v `pwd`/testdata:/usr/share/nginx/html:ro \
-p 8081:80 \
-d \
nginx

# Allow reverse proxy to listen on ports 80 and 443
go build

# Start the reverse proxy
./webfront -rules=./rules.json