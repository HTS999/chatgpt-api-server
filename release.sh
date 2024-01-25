#!/bin/bash

set -e

gf build main.go -a amd64 -s linux
gf docker main.go -t hts999/chatgpt-api-server:latest
# 修改镜像标签为当前日期时间
time=$(date "+%Y%m%d%H%M%S")
# 获取当前git commit id 
commit_id=$(git rev-parse HEAD)
docker tag hts999/chatgpt-api-server:latest hts999/chatgpt-api-server:$time-$commit_id
# 推送镜像到docker hub
docker push hts999/chatgpt-api-server:latest
docker push hts999/chatgpt-api-server:$time-$commit_id