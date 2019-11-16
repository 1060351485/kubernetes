#!/bin/sh

systemctl start docker
ipfs daemon &
