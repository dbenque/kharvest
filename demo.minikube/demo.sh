#!/bin/bash
NO="\e[0m"
BOLD="\e[1m"
RED="\e[91m"
GREEN="\e[92m"
YELLOW="\e[93m"

echo -e $YELLOW"Clean previous run if any..."$NO
kubectl delete deployment,rs,rc,configmap,service -l demo=kharvest
kubectl delete deployment,rs,rc,configmap,service -l run=kharvest
kubectl delete deployment,rs,rc,configmap,service -l run=kharvestclient

echo $YELLOW"Deploy minifileserver dependency..."$NO
kubectl apply -f minifileserver.yaml

echo $YELLOW"Building and deploying kharvest server"$NO
cd kharvest
go run *.go --deploy --labels='{"demo":"kharvest"}'
cd ..

echo $YELLOW"Create files for demo using configmap"$NO
kubectl create configmap cmkharvest --from-file=files
kubectl label configmap cmkharvest demo=kharvest

echo $YELLOW"Configuring kharvest clients"$NO
kubectl create -f kharvestclient/clientConfig.yaml

echo $YELLOW"Building and deploying kharvest client"$NO
cd kharvestclient
go run *.go --deploy --configMaps='["cmkharvest","kharvest-client"]' --labels='{"demo":"kharvest"}'
cd ..

read -n 1 -s -r -p $GREEN"Wait for the pod to be running then press any key..."$NO
echo $YELLOW"Here are the logs of the kharvest server:"$NO
kubectl logs $(kubectl get pod -l run=kharvest -ojsonpath='{.items[0].metadata.name}')
echo
echo $YELLOW"Here are the logs of the kharvest client:"$NO
kubectl logs $(kubectl get pod -l run=kharvestclient -ojsonpath='{.items[0].metadata.name}')
echo

read -n 1 -s -r -p $GREEN"Now we are going to scale the number of replicas for the clientto 3. They will ahve the same files (distributed by configmap). Press any key to proceed..."$NO
echo $YELLOW"Scaling the clients to 3:"$NO
kubectl scale --replicas=3 rs/kharvestclient
echo
read -n 1 -s -r -p $GREEN"Wait for all the pod to be running then press any key..."$NO
echo $YELLOW"Here are the logs of the kharvest server:"$NO
kubectl logs $(kubectl get pod -l run=kharvest -ojsonpath='{.items[0].metadata.name}')
echo
echo $YELLOW"Here are the logs of the kharvest client:"$NO
kubectl logs $(kubectl get pod -l run=kharvestclient -ojsonpath='{.items[0].metadata.name}')
echo
echo $GREEN"Now try to edit the content of one file in the config map using the following command and monitor the logs of clients and servers."$NO
echo $GREEN"Remember the configmap sometimes take couple seconds to be distributed."$NO
echo $BOLD"kubectl edit configmap cmkharvest"$NO
echo $BOLD"kubectl logs $(kubectl get pod -l run=kharvest -ojsonpath='{.items[0].metadata.name}')"$NO
echo $BOLD"kubectl logs $(kubectl get pod -l run=kharvestclient -ojsonpath='{.items[0].metadata.name}')"$NO
echo $BOLD"kubectl logs $(kubectl get pod -l run=kharvestclient -ojsonpath='{.items[1].metadata.name}')"$NO
echo $BOLD"kubectl logs $(kubectl get pod -l run=kharvestclient -ojsonpath='{.items[2].metadata.name}')"$NO