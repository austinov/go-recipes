#!/bin/bash

set -e

echo "Remove previouse images:"
sudo docker rm -v -f rocker || :
sudo docker rmi -f rocker-bot || :
sudo docker rm -v -f pgdb || :
sudo docker rmi -f rocker-db || :
sudo docker images -f dangling=true -q --no-trunc | sudo xargs -r docker rmi -f

pushd "$(dirname "$(readlink -f "$BASH_SOURCE[0]")")" > /dev/null && {
  cd ..

  sudo docker build -f ./rocker-bot/docker/db/Dockerfile -t rocker-db .
  sudo docker run -t --name pgdb -d rocker-db

  sudo docker build -f ./rocker-bot/docker/bot/Dockerfile -t rocker-bot .
  sudo docker run -t --name rocker --link pgdb:pgdb -d rocker-bot

  popd > /dev/null
}
