## Build and Use
```sh
go build -o copy-machinesets ./pkg
./copy-machinesets --kubeconfig $KUBECONFIG --hosted-namespace ${NAMESPACE}\
  --create-replicas
```

Note, all machinesets are created in openshift-machine-api namespace, we just
use hosted-namespace to prefix/suffix the name of the new machinesets.

## Help
```sh
#help

Usage of ./copy-machinesets:
  -create-replicas
    	Create replicas for the copy for any machineset having > 0 replicas (default: false)
  -hosted-namespace string
    	namespace of the hosted cluster (eg: 'hosted')
  -kubeconfig string
    	absolute path to the kubeconfig file
```

## Useful things you can copy and paste
```
$OC get machinesets -n openshift-machine-api
$OC get machines -n openshift-machine-api
```

## Cleanup
```
$OC get machinesets -nopenshift-machine-api | grep ${NAMESPACE} | \
  awk '{print $1}' | xargs $OC delete machinesets -nopenshift-machine-api

```
