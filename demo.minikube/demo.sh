#!/bin/bash
NO="\e[0m"
BOLD="\e[1m"
RED="\e[91m"
GREEN="\e[92m"
YELLOW="\e[93m"

echo -e $YELLOW"Clean previous run if any..."$NO
kubectl delete deployment,rs,rc,configmap,service -l demo=kharvest

echo -e $YELLOW"Setup Docker env for minikube"$NO
eval $(minikube docker-env)

echo -e $YELLOW"Building and deploying kharvest server"$NO
cd kharvest
./docker.sh
kubectl create -f kharvest.deployment.yaml
kubectl create -f kharvest.service.yaml
cd ..

echo -e $YELLOW"Create files for demo using configmap"$NO
kubectl create configmap cmkharvest --from-file=files
kubectl label configmap cmkharvest demo=kharvest

echo -e $YELLOW"Configuring kharvest clients"$NO
kubectl create -f kharvestclient/clientConfig.yaml

echo -e $YELLOW"Building and deploying kharvest client"$NO
cd kharvestclient
./docker.sh
kubectl create -f kharvestclient.deployment.yaml
cd ..
echo -e $GREEN
read -n 1 -s -r -p "Wait for the pod to be running then press any key..."
sleep 2
echo -e $NO
echo -e $YELLOW"Here are the logs of the kharvest server:"$NO
kubectl logs $(kubectl get pod -l run=kharvest -ojsonpath='{.items[0].metadata.name}')
echo
echo -e $YELLOW"Here are the logs of the kharvest client:"$NO
kubectl logs $(kubectl get pod -l run=kharvestclient -ojsonpath='{.items[0].metadata.name}')
echo

echo -e $GREEN
read -n 1 -s -r -p "Now we are going to scale the number of replicas for the client to 3. They will have the same files (distributed by configmap). Press any key to proceed..."
echo -e $NO
echo -e $YELLOW"Scaling the clients to 3:"$NO
kubectl scale --replicas=3 deployment/kharvestclient
echo
echo -e $GREEN
read -n 1 -s -r -p "Wait for all the pod to be running then press any key..."
sleep 2
echo -e $NO
echo -e $YELLOW"Here are the logs of the kharvest server:"$NO
kubectl logs $(kubectl get pod -l run=kharvest -ojsonpath='{.items[0].metadata.name}')
echo
echo -e $BOLD"Note that the server has just acknowledged some Notifications; no store was performed since it recognized the files, only references are added."$NO
echo
read -n 1 -s -r -p "Press any key to continue the demo..."
###CLI
echo -e $YELLOW"Building the CLI"$NO
go build -o kharvestcli kharvestcli/main.go

KHARVESTURL=$(minikube ip):$(kubectl get svc kharvest -ojsonpath='{.spec.ports[?(@.name=="user")].nodePort}')
APOD=$(kubectl get pod -l run=kharvestclient -ojsonpath='{.items[0].metadata.name}')
echo -e $YELLOW"List all the files referenced by a pod:"$NO
echo -e $BOLD"./kharvestcli -k=$KHARVESTURL -cmd=pod -p=$APOD"$NO
./kharvestcli -k=$KHARVESTURL -cmd=pod -p=$APOD
echo
sleep 2
echo -e $YELLOW"List all the pod that reference the same file that was presented at index=0:"$NO
echo -e $BOLD"./kharvestcli -k=$KHARVESTURL -cmd=same -p=$APOD -i=0"$NO
./kharvestcli -k=$KHARVESTURL -cmd=same -p=$APOD -i=0
echo
sleep 2
# Let the user play
echo -e $GREEN"Remember the configmap sometimes take couple seconds to be distributed."$NO
echo -e $GREEN"Now try to edit the content of one file in the config map using the following command and monitor the logs of clients and servers."$NO
echo -e $GREEN"Remember the configmap sometimes take couple seconds to be distributed."$NO
echo -e $BOLD"kubectl edit configmap cmkharvest"$NO
echo -e $BOLD"kubectl logs $(kubectl get pod -l run=kharvest -ojsonpath='{.items[0].metadata.name}')"$NO
echo -e $BOLD"kubectl logs $(kubectl get pod -l run=kharvestclient -ojsonpath='{.items[0].metadata.name}')"$NO
echo -e $BOLD"kubectl logs $(kubectl get pod -l run=kharvestclient -ojsonpath='{.items[1].metadata.name}')"$NO
echo -e $BOLD"kubectl logs $(kubectl get pod -l run=kharvestclient -ojsonpath='{.items[2].metadata.name}')"$NO
echo
echo -e $GREEN"Or use the CLI:"$NO
echo -e $BOLD"./kharvestcli -k=$KHARVESTURL -cmd=[keys|pod|same] ..."$NO
./kharvestcli -cmd=help
echo
