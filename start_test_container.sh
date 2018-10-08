#!/bin/bash

# Start the proxied server
docker stop test-nginx
docker rm test-nginx
docker run --name test-nginx \
-v `pwd`/testdata:/usr/share/nginx/html:ro \
-p 8081:80 \
-d \
nginx