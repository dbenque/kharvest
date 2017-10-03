#!/bin/bash
NO="\e[0m"
BOLD="\e[1m"
RED="\e[91m"
GREEN="\e[92m"
YELLOW="\e[93m"

## waitendpointcount waits for the endpoint ($1) to contains a given number of ip ($2)
## $1=servicename $2=expected count
function waitendpointcount()
{
  EPC="-1"
  while [[ ! EPC -eq $2 ]]
  do
    IPS=$(kubectl get endpoints $1 --output=jsonpath={.subsets[*].addresses[*].ip})
    DOT=$(grep -o "\." <<< "$IPS " | wc -l)
    EPC=$(expr $DOT / 3)
    echo "Waiting for endpoints $1: $EPC/$2"
    sleep 1
  done
}

function waitReplicasOnDeployment()
{
  RUNNINGC="-1"
  while [[ ! RUNNINGC -eq $2 ]]
  do
    RUNNINGC=$(kubectl get deployment $1 -ojsonpath='{.status.readyReplicas}')
    if [[ RUNNINGC -eq "" ]]; then
        RUNNINGC="0"
    fi
    echo "Waiting pods $1 for being ready: $RUNNINGC/$2"
    sleep 1
  done
}

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

waitendpointcount kharvest 1

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
echo -e $YELLOW"Waiting for the client to be scheduled and ready ..."$NO
waitReplicasOnDeployment kharvestclient 1
echo -e $YELLOW"Opening log windows...wait for connectivity to happen and more logs in the server..."$NO
xterm -geometry 155x40 -sb -sl 1000 -e "kubectl logs $(kubectl get pod -l run=kharvest -ojsonpath='{.items[0].metadata.name}') -f" &
xterm -geometry 155x40 -sb -sl 1000 -e "kubectl logs $(kubectl get pod -l run=kharvestclient -ojsonpath='{.items[0].metadata.name}') -f" &
echo -e $GREEN
read -n 1 -s -r -p "Now we are going to scale the number of replicas for the client to 3. They will have the same files (distributed by configmap). Press any key to proceed..."
echo -e $NO
echo -e $YELLOW"Scaling the clients to 3:"$NO
kubectl scale --replicas=3 deployment/kharvestclient
echo
waitReplicasOnDeployment kharvestclient 3
echo -e $NO
echo -e $YELLOW"Look at the server logs, they should have been updated by notifications from the added pods."$NO
echo -e $BOLD"Note that the server has just acknowledged some Notifications; no store was performed since it recognized the files, only references are added."$NO
echo
read -n 1 -s -r -p "Press any key to continue the demo..."
echo
echo -e $GREEN"An edit window will pop-up, edit the content of file3 in the configmap and save (ESC+:wq). This will modify the content of some files in the pods."$NO
xterm -geometry 120x40 -sb -sl 1000 -e "kubectl edit cm cmkharvest"
echo -e $YELLOW"Look at the server logs, they should have been updated by notifications from the added pods."$NO
echo -e $YELLOW"Remark: the AAAAAAAAAAAAAAAAAAAAAA== and 0 size happens because configmaps are using symlinks and not plain files.A link switch is seen as a file REMOVE+CREATE"$NO
echo
read -n 1 -s -r -p "Press any key to continue the demo..."
###CLI
echo
echo -e $YELLOW"Building the CLI"$NO
go build -o democli kharvestcli/main.go

KHARVESTURL=$(minikube ip):$(kubectl get svc kharvest -ojsonpath='{.spec.ports[?(@.name=="user")].nodePort}')
APOD=$(kubectl get pod -l run=kharvestclient -ojsonpath='{.items[0].metadata.name}')
NAMESPACE=$(kubectl get pod -l run=kharvestclient -ojsonpath='{.items[0].metadata.namespace}')
echo -e $YELLOW"List all the files referenced by a pod:"$NO
echo -e $BOLD"./democli -k=$KHARVESTURL -cmd=pod -p=$APOD -n=$NAMESPACE"$NO
./democli -k=$KHARVESTURL -cmd=pod -n=$NAMESPACE -p=$APOD
echo
sleep 2
echo -e $YELLOW"List all the pods that reference the same file that was presented at index=0:"$NO
echo -e $BOLD"./democli -k=$KHARVESTURL -cmd=same -p=$APOD -n=$NAMESPACE -i=0"$NO
./democli -k=$KHARVESTURL -cmd=same -p=$APOD -n=$NAMESPACE -i=0
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
echo -e $GREEN"Or use the CLI (see example above):"$NO
echo -e $BOLD"./democli -k=$KHARVESTURL -cmd=[pod|same] -n=$NAMESPACE ..."$NO

echo
