sudo systemctl daemon-reload
sudo systemctl enable kubelet
sudo systemctl start kubelet
sudo systemctl enable kube-proxy
sudo systemctl start kube-proxy
sudo systemctl status kubelet
sudo systemctl status kube-proxy
