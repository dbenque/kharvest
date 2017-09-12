#!/bin/bash
kubectl delete deployment,rs,rc,configmap -l demo=kharvest
kubectl delete deployment,rs,rc,configmap -l run=kharvest