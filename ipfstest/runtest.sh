#!/bin/bash

reset_env_before_run(){
    ansible para -i hosts.txt -m script -a "./clean_env.sh" -b=true;
}

start_service(){
    ansible para -i hosts.txt -m script -a "./start_service.sh" -b=true;
}

still_running(){
   count=`kubectl get deploy -o json | jq '.items[0].status.readyReplicas'`
   #echo $count, $1
   if [ $count == $1 ]
   then
       return 1
   else
       return 0
   fi    
}

run_test(){
    # reset_env_before_run();
    # start_service();

    # record time
    d1=`date +%s%N`
    echo $d1

    if [ $1 == 0 ]
    then
        echo $d1 > ./log/raw-$d1.txt
        kubectl apply -f k8s_raw.yaml
    else
        echo $d1 > ./log/ipfs-$d1.txt
        kubectl apply -f k8s_ipfs.yaml
    fi

    while (still_running $2)
    do true
    done

    d2=`date +%s%N`
    echo $d2
    d3=`expr $d2 - $d1`
    echo $d3 ns

    if [ $1 == 0 ]
    then
        echo $d2 >> ./log/raw-$d1.txt
        echo $d3 >> ./log/raw-$d1.txt
        kubectl delete deploy deploy-raw 
    else 
        echo $d2 >> ./log/ipfs-$d1.txt
        echo $d3 >> ./log/ipfs-$d1.txt
        kubectl delete deploy deploy-ipfs
    fi
}

# {raw/ipfs, #replicas}
run_test $1 $2

