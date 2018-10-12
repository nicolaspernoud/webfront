#!/bin/bash

# Start the proxied server
docker stop test-nginx
docker rm test-nginx
docker run --name test-nginx \
-v `pwd`/testdata:/usr/share/nginx/html:ro \
-v `pwd`/nginx.conf:/etc/nginx/nginx.conf:ro \
-v `pwd`/default.conf:/etc/nginx/conf.d/default.conf:ro \
-p 8081:80 \
-d \
nginx