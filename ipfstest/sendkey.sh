#!/bin/sh


ansible para -i ./hosts.txt -m authorized_key -a "user=root key='{{ lookup('file', '~/.ssh/id_rsa.pub')}}' path='/root/.ssh/authorized_keys' manage_dir=no" --ask-pass
