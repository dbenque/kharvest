# Demo of Kharvest in minikube

Prerequisit:
- minikube running
- minikube binary in the PATH
- kubectl binary in the PATH
- docker binary in the PATH
- GOPATH set correctly

Be sure that your GOPATH variable is correctly set because the demo does build code!

You can run the demo in any namespace.Enter folder **demo.minikube** and run demo script **demo.sh**

To clean the objects created by the script, used **clean.sh** located in the same folder.

## Code

Thanks to that demo you can see:
- How to run the kharvest server inside the folder **kharvest**. There are 2 servers that can be run independently:
  - The "internal server" that receive notification and store 
  data files and references: 
  ```
  server.RunKharvestServer(...)
  ```
  - The "API server" that allow to retrieve information stored (consumed by the CLI): 
  ```
  server.RunKharvestServerUserAPI(...)
  ```

- How to run the kharvest client in each pod inside the folder **kharvestclient**, build the config, watch the change of configuration and launch the client part:
```
    conf := client.NewConfig(os.Getenv("PODNAME"), os.Getenv("NAMESPACE"))
    conf.ConfigPath = "/cfg/kharvest-client/kharvest.cfg"

    conf.ReadAndWatch();

    client.RunKharvestClient(conf)
```

- How to consume the grpc user API, with an example of a very dumb CLI in folder **kharvestcli**