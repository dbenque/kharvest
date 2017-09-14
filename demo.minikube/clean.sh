#!/bin/bash
kubectl delete deployment,rs,rc,configmap,service -l demo=kharvest
