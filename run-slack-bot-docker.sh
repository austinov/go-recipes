#!/bin/bash

set -e

echo "Remove previouse images:"
sudo docker rm -v -f rocker || :
sudo docker rmi -f slack-bot || :
sudo docker rm -v -f pgdb || :
sudo docker rmi -f slack-bot-db || :
sudo docker images -f dangling=true -q --no-trunc | sudo xargs -r docker rmi -f

sudo docker build -f ./slack-bot/docker/db/Dockerfile -t slack-bot-db .
sudo docker run -t --name pgdb -d slack-bot-db

sudo docker build -f ./slack-bot/docker/bot/Dockerfile -t slack-bot .
sudo docker run -t --name rocker --link pgdb:pgdb -d slack-bot
