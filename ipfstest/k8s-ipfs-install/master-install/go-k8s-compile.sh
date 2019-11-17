# install go and setup env variables
wget https://dl.google.com/go/go1.12.10.linux-amd64.tar.gz  -P ~/
sudo tar -C /usr/local -xzf ~/go1.12.10.linux-amd64.tar.gz
echo "export GOPATH=~/" >> ~/.profile
echo "sudo swapoff -a" >> ~/.profile
echo "export PATH=\$PATH:/usr/local/go/bin" >> ~/.profile
echo "TZ='America/New_York'; export TZ" >> ~/.profile
echo "export PATH=\"/users/mkunal/src/k8s.io/kubernetes/third_party/etcd=\${PATH}\"" >> ~/.profile
source  ~/.profile

# install ipfs and generate swarm key
wget https://dist.ipfs.io/go-ipfs/v0.4.22/go-ipfs_v0.4.22_linux-amd64.tar.gz -P ~/
tar xvfz ~/go-ipfs_v0.4.22_linux-amd64.tar.gz
sudo mv go-ipfs/ipfs /usr/local/bin/ipfs
rm -R ~/go-ipfs

IPFS_PATH=~/.ipfs ipfs init
# generate swarm key on master node
echo -e "/key/swarm/psk/1.0.0/\n/base16/\n$(tr -dc 'a-f0-9' < /dev/urandom | head -c64)" > ~/.ipfs/swarm.key

# clone kubernetes repo
mkdir -p $GOPATH/src/k8s.io
cd $GOPATH/src/k8s.io
git clone https://github.com/1060351485/kubernetes
cd $GOPATH/src/k8s.io/kubernetes
git checkout myb
make -j 4

# install etcd
wget 'https://github.com/etcd-io/etcd/releases/download/v3.3.10/etcd-v3.3.10-linux-amd64.tar.gz'
sudo tar -zxvf etcd-v3.3.10-linux-amd64.tar.gz -C /usr/local/src/
sudo cp /usr/local/src/etcd-v3.3.10-linux-amd64/etc* /usr/bin
rm etcd-v3.3.10-linux-amd64.tar.gz
sudo cp etcd.service /etc/systemd/system/
sudo mkdir /var/lib/etcd
# start etcd
sudo systemctl daemon-reload
mkdir -p /var/lib/etcd/
sudo systemctl start etcd.service
sudo systemctl enable etcd.service
# check status
sudo netstat -anop | grep etcd
etcdctl cluster-health

# copy kubctl bin
sudo cp $GOPATH/src/k8s.io/kubernetes/_output/bin/kubectl /usr/bin/

# copy kube-apiserver bin
sudo cp $GOPATH/src/k8s.io/kubernetes/_output/bin/kube-apiserver /usr/bin/
sudo cp kube-apiserver.service /etc/systemd/system/
sudo mkdir -p /etc/kubernetes
sudo cp apiserver.conf /etc/kubernetes/

# copy kube-controller-manager bin
sudo cp $GOPATH/src/k8s.io/kubernetes/_output/bin/kube-controller-manager /usr/bin/
sudo cp kube-controller-manager.service /etc/systemd/system/
# need fill in master ip
sudo cp controller-manager.conf /etc/kubernetes/

# copy kube-scheduler bin
sudo cp $GOPATH/src/k8s.io/kubernetes/_output/bin/kube-scheduler /usr/bin/
sudo cp kube-scheduler.service /etc/systemd/system/
# need fill in master ip
sudo cp scheduler.conf /etc/kubernetes/

# start k8s services on master node
sudo systemctl daemon-reload
sudo systemctl enable kube-apiserver.service
sudo systemctl start kube-apiserver.service
sudo systemctl enable kube-controller-manager.service
sudo systemctl start kube-controller-manager.service
sudo systemctl enable kube-scheduler.service
sudo systemctl start kube-scheduler.service

sudo systemctl status kube-apiserver.service
sudo systemctl status kube-controller-manager.service
sudo systemctl status kube-scheduler.service

sudo apt-get install ansible
# !!! fill in hosts in /etc/ansible/hosts
# [ipfs-nodes]
# 123.32.3.2
