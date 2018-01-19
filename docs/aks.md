## Stemcell

CPI is configured to use GCR with heavy warden stemcells.

## Director

See `deployments/aks/`

Nothing special.

- LB creation failed initially

## CF

See `deployments/aks-cf/`

Requirements:

- must reboot nodes with `cgroup_enable=memory swapaccount=1` in `/boot/grub/grub.cfg`
  - verify that `cat /proc/cmdline` contains above configuration

Notes:

- AKS seems to be running 16.04.3 with 4.11 kernel

## Azure CLI

Install CLI:

```
echo "deb [arch=amd64] https://packages.microsoft.com/repos/azure-cli/ wheezy main" |      sudo tee /etc/apt/sources.list.d/azure-cli.list
sudo apt-key adv --keyserver packages.microsoft.com --recv-keys 52E16F86FEE04B979B07E28DB02C46DF417A0893
sudo apt-get install apt-transport-https
sudo apt-get update && sudo apt-get install azure-cli

az login
az provider register -n Microsoft.ContainerService
az provider show -n Microsoft.ContainerService
watch 'az provider show -n Microsoft.ContainerService|head -20'
```

Create cluster:

```
az group create --name myResourceGroup --location eastus
az aks create --resource-group myResourceGroup --name myK8sCluster --node-count 1 --generate-ssh-keys
az aks install-cli
az aks get-credentials --resource-group myResourceGroup --name myK8sCluster

kubectl version
kubectl get nodes
```

Scale up cluster:

```
az aks -h
az aks scale -n myK8sCluster -c 3 -g myResourceGroup
```

Fix kernel boot config:

```
eval `ssh-agent`
ssh-add ~/.ssh/id_rsa
ssh azureuser@<private-ip> -i /root/.ssh/id_rsa
```
