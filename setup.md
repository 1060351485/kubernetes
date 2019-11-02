# Kuberenetes with IPFS

 How to setup env on Ubuntu 18.04 LTS machines

## Manually compile kubernetes services on master node

1. Setup [Golang] environment on master node. 

2. Install [Docker] on both master and worker nodes.

    Enable docker auto start
    ```
    systemctl enable docker
    ```

3. Download [kubernetes] under go path, and compile binaries.

    ```
    mkdir -p $GOPATH/src/k8s.io
    cd $GOPATH/src/k8s.io
    git clone https://github.com/kubernetes/kubernetes
    cd kubernetes
    make
    ```
4. Binaries are under ``` $GOPATH/src/k8s.io/kubernetes/_output/bin/ ``` folder


## Install services on master node

!! copy kubectl to master node /usr/bin/

5. Setup [etcd] service on **master node**

    Download and install etcd service
    ```
    wget 'https://github.com/etcd-io/etcd/releases/download/v3.3.10/etcd-v3.3.10-linux-amd64.tar.gz'

    sudo tar -zxvf etcd-v3.3.10-linux-amd64.tar.gz -C /usr/local/src/

    sudo cp /usr/local/src/etcd-v3.3.10-linux-amd64/etc* /usr/bin
    ```

    Add a service config file:
    ```/usr/lib/systemd/system/etcd.service```
    ```
    [Unit]
    Description=Etcd Server
    After=network-manager.target
    [Service]
    Type=notify
    EnvironmentFile=-/etc/etcd/etcd.conf
    WorkingDirectory=/var/lib/etcd/
    ExecStart=/usr/bin/etcd
    Restart=on-failure
    [Install]
    WantedBy=multi-user.target
    ```

    Start etcd service on master node
    ```
    systemctl daemon-reload
    mkdir -p /var/lib/etcd/
    systemctl start etcd.service
    systemctl enable etcd.service
    ```

    Check etcd service status
    ```
    netstat -anop | grep etcd
    etcdctl cluster-health
    ```

6. Install kube-apiserver on **master node**
    
    Copy kube-apiserver binary to ```/usr/bin/```

    Add a service config file:
    ```/usr/lib/systemd/system/kube-apiserver.service```

    ```
    [Unit]
    Description=Kubernetes API Server
    Documentation=https://github.com/kubernetes/kubernetes
    After=etcd.service
    Wants=etcd.service
    [Service]
    EnvironmentFile=/etc/kubernetes/apiserver.conf
    ExecStart=/usr/bin/kube-apiserver $KUBE_API_ARGS
    Restart=on-failure
    Type=notify
    [Install]
    WantedBy=multi-user.target

    ```

    ```mkdir -p /etc/kubernetes```

    Add another config file for KUBE_API_ARGS:
    ```/etc/kubernetes/apiserver.conf```

    ```
    UBE_API_ARGS="--storage-backend=etcd3 --etcd-servers=http://127.0.0.1:2379 --insecure-bind-address=0.0.0.0 --insecure-port=8080 --service-cluster-ip-range=10.10.10.0/24 --service-node-port-range=1-65535 --admission-control=Namesp
    aceLifecycle,NamespaceExists,LimitRanger,ResourceQuota --logtostderr=true --log-dir=/var/log/kubernetes --v=4"
    ```

7. Install kube-controller-manager on **master node**

    Copy kube-controller-manager binary to ```/usr/bin```

    Add a service config file for this service

    ```
    /usr/lib/systemd/system/kube-controller-manager.service:
    
    [Unit]
    Description=Kubernetes Controller Manager
    Documentation=https://github.com/GoogleCloudPlatform/kubernetes
    After=kube-apiserver.service
    Requires=kube-apiserver.service

    [Service]
    EnvironmentFile=-/etc/kubernetes/controller-manager.conf
    ExecStart=/usr/bin/kube-controller-manager $KUBE_CONTROLLER_MANAGER_ARGS
    Restart=on-failure
    LimitNOFILE=65536

    [Install]
    WantedBy=multi-user.target
    ```

    Add a service config file for KUBE_CONTROLLER_MANAGER_ARGS:

    ```
    /etc/kubernetes/controller-manager.conf:

    UBE_CONTROLLER_MANAGER_ARGS="--master=http://{your master node IP}:8080 --logtostderr=true --log-dir=/var/log/kubernetes --v=4"
    ```
8. Install kube-scheduler on **master node**

    Copy kube-scheduler binary to ```/usr/bin```

    Add a service config file:

    ```
    /usr/lib/systemd/system/kube-scheduler.service:

    [Unit]
    Description=Kubernetes Scheduler
    Documentation=https://github.com/GoogleCloudPlatform/kubernetes
    After=kube-apiserver.service
    Requires=kube-apiserver.service

    [Service]
    EnvironmentFile=-/etc/kubernetes/scheduler.conf
    ExecStart=/usr/bin/kube-scheduler $KUBE_SCHEDULER_ARGS
    Restart=on-failure
    LimitNOFILE=65536

    [Install]
    WantedBy=multi-user.target
    ```

    Add a config file for KUBE_SCHEDULER_ARGS

    ```
    /etc/kubernetes/scheduler.conf

    KUBE_SCHEDULER_ARGS="--master=http://{your master node IP}:8080 --logtostderr=true --log-dir=/var/log/kubernetes --v=2"
    ```

9. Start services on master node and check service status

    ```
    sudo systemctl daemon-reload
    sudo systemctl enable kube-apiserver.service
    sudo systemctl start kube-apiserver.service
    sudo systemctl enable sudo kube-controller-manager.service
    sudo systemctl start kube-controller-manager.service
    sudo systemctl enable kube-scheduler.service
    sudo systemctl start kube-scheduler.service

    sudo systemctl status kube-apiserver.service
    sudo systemctl status kube-controller-manager.service
    sudo systemctl status kube-scheduler.service
    ```

## Install services on worker nodes

10. There are some services need to install before k8s services
    
    1). docker-ce

    2). [containerd.io]

    3). network-manager

    Enable docker auto start
    ```
    systemctl enable docker
    ```

11. Install kubelet on **worker nodes**

    Copy kubelet binary to ```/usr/bin/``` on worker nodes

    Add a service config file

    ```
    /usr/lib/systemd/system/kubelet.service

    [Unit]
    Description=Kubernetes Kubelet Server
    Documentation=https://github.com/GoogleCloudPlatform/kubernetes
    After=docker.service
    Requires=docker.service

    [Service]
    WorkingDirectory=/var/lib/kubelet
    EnvironmentFile=-/etc/kubernetes/kubelet.conf
    ExecStart=/usr/bin/kubelet $KUBELET_ARGS
    Restart=on-failure
    KillMode=process

    [Install]
    WantedBy=multi-user.target
    ```

    Add a config file for KUBELET_ARGS

    ```
    /etc/kubernetes/kubelet.conf

    UBELET_ARGS="--address={your worker node IP} --port=10250 --kubeconfig=/etc/kubernetes/kubelet.kubeconfig --cluster-dns=10.10.10.2 --cluster-domain=cluster.local --fail-swap-on=false --alsologtostderr=true --log-dir=/var/log/kubernetes --log-file=kubelet.log --v=4"
    ```


    Add a config file to connect with kube-apiserver

    ```
    /etc/kubernetes/kubelet.kubeconfig


    apiVersion: v1
    kind: Config
    clusters:
    - cluster:
        server: http://{your master node IP}:8080
        name: local
    contexts:
    - context:
        cluster: local
        name: local
    current-context: local
    ```

12. Install kube-proxy on **woker nodes**

    Add a service config file
    ```
    /usr/lib/systemd/system/kube-proxy.service

    [Unit]
    Description=Kubernetes Kube-proxy Server
    Documentation=https://github.com/GoogleCloudPlatform/kubernetes
    After=network-manager.service
    Requires=network-manager.service

    [Service]
    EnvironmentFile=/etc/kubernetes/proxy.conf
    ExecStart=/usr/bin/kube-proxy $KUBE_PROXY_ARGS
    Restart=on-failure
    LimitNOFILE=65536
    KillMode=process

    [Install]
    WantedBy=multi-user.target
    ```

    Add a config file for KUBE_PROXY_ARGS

    ```
    /etc/kubernetes/proxy.conf

    KUBE_PROXY_ARGS="--master=http://{your master node IP}:8080 --logtostderr=true --log-dir=/var/log/kubernetes --v=4"
    ```

13. Start kubelet and kube-proxy on **worker nodes**

    ```
    mkdir -p /var/lib/kubelet
    sudo systemctl daemon-reload
    sudo systemctl enable kubelet
    sudo systemctl start kubelet
    sudo systemctl enable kube-proxy
    sudo systemctl restart kube-proxy
    sudo systemctl status kubelet
    sudo systemctl status kube-proxy
    ```

## Check cluster status

14. Check cluster status on master node

    You should see all worker nodes listed below
    ```
    kubectl get nodes
    ```
    

[Golang]: https://golang.org/
[kubernetes]: https://github.com/kubernetes/kubernetes
[Docker]: https://www.docker.com/
[etcd]: https://github.com/etcd-io/etcd
[containerd.io]: https://containerd.io/
