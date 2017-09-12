# Demo of Kharvest

A Kharvest server will be running (from the demo folder) :


```
cd kharvest
go run *.go --deploy
```

This server will receive notifications and pushes from clients. The client will be configured via configMap (demo/kharvestclient/clientConfig.yaml)

```
kind: ConfigMap
apiVersion: v1
metadata:
  name: kharvest-client
  namespace: default
data:
  kharvest.cfg: |-
    /cfg/cmkharvest/file1
    /cfg/cmkharvest/file2
    /cfg/cmkharvest/file3
```

This config tells the client to monitor 3 figo run *.go --deploy --configMaps='["cmkharvest","kharvest-client"]'les. To create the associated configMap (from the demo folder) :

```
kubectl create configmap -f kharvestclient/clientConfig.yaml
```

To change the content of the 3 files in pods we will also use a configMap (from the demo folder):

```
kubectl create configmap cmkharvest --from-file=files
```

Now that the two config files are created we can run the client(from the demo folder):
```
cd kharvestclient
go run *.go --deploy --configMaps='["cmkharvest","kharvest-client"]'
```


