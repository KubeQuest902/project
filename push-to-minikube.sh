#!/bin/bash
set -e

### How to use this script
#
# 1) Copy this script to the workspace of any microservice you want to manually push to Minikube
# 2) Adapt the three variables "user", "host" and "image" to point to your Minikube VM and ythe image name you want to push
# 3) Run the script
# 4) Edit the deployment (or other controller) of the local microservice to use the "dev" tag of the image, via "kubectl edit deploy ..."

microservice=project
namespace=kubequest
user=docker #pass=tcuser
host=$(minikube ip -p zolara)
image=nexuszolara.me/kubequest/$microservice

echo "Push to minikube Started"
go mod vendor
GOOS=linux GARCH=amd64 CGO_ENABLED=0 go build -o app
chmod +x app
docker build . -t $image:KUBE_GIT_TAG -f Dockerfile
docker save -o image.tar $image:KUBE_GIT_TAG

# sshpass -p "tcuser" scp image.tar $user@$host:/tmp/image.tar
# sshpass -p "tcuser" ssh -t $user@$host "docker load -i /tmp/image.tar"

minikube -p zolara cp image.tar /tmp/image.tar
minikube -p zolara ssh "docker load -i /tmp/image.tar"

rm image.tar
rm app

sed -i 's/version: KUBE_GIT_TAG/version: v0-KUBE-GIT-TAG/' helm/$microservice/Chart.yaml
helm upgrade --install zolara-$microservice -n $namespace helm/$microservice --reset-values
sed -i 's/version: v0-KUBE-GIT-TAG/version: KUBE_GIT_TAG/' helm/$microservice/Chart.yaml

kubectl delete pods -n $namespace -l app=$microservice