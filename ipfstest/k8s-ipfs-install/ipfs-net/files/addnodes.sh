#!/bin/sh

ipfs bootstrap rm --all
ipfs pin ls --type recursive | cut -d' ' -f1 | xargs -n1 ipfs pin rm
ipfs repo gc
ipfs bootstrap add /ip4/128.110.154.137/tcp/4001/ipfs/QmZ6Er4EJXjNFUp72dinzS12os4Q18DPUhEpHx35xeNzPP
ipfs bootstrap add /ip4/128.110.154.150/tcp/4001/ipfs/QmYy1LyHvTcRAfgmi9SJRSZh2iUzf9uvyx6T2K4zVm8SwN
