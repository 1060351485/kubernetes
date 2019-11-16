#!/bin/sh

function install(){
# install docker

# install kubeproxy kubelet cri

# install ipfs

    ansible para -i hosts.txt -m shell -a "mkdir /var/tmp/k8sipfs" -b=true
}
