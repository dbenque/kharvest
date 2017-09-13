# kharvest
Harvest files in kubernetes cluster.

In a kubernetes cluster running thousand of pods, you may need to collect files produced locally in the pod. Pods being controlled by replication controller, the chance that they all produce the same file is high. Kharvest helps you to collect these files in an efficient way, transfering the file only once in the network for a set of pod producing the same file.

Kharvest comes with a server that index and store the file and a client.

The server is link to a store to persist harvested files. The storage backend can be whatever database layer as soon as it implements the Store interface (see file store/interface.go).

The client can be directly be embedded in the application if it is written in Go by calling the **client.RunKharvestClient** in your main. Also it can be in as a sidecar container, in that case pay attention to mount correctly the volume where the files to harvest are located.


