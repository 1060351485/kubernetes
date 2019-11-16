#!/bin/sh

docker stop $(docker ps -a -q)
docker rm -f $(docker ps -a -q)
docker rmi -f $(docker images -a -q)

systemctl stop docker

rm -rf /var/tmp/k8sipfs

killall ipfs

# ipfs clear cache?


