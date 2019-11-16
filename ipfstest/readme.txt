1. install ansible
2. generate a ssh pubkey on master node
2. sendkey.sh use ansible to send ssh pub key to hosts, hosts list is ./hosts.txt
3. install.sh install env
4. run runtest.sh with {0 for raw, 1 for ipfs} {# of replicas} for a single test. log is in log/ dir, filename is time in ns, and 3 lines in each log file is: start time, end time, running time in ns. 
