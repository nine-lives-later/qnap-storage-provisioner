# QNAP Storage Provisioner for QNAP storages

## Deployment

To deploy the QNAP Storage Provisioner for QNAP storages, see the [example deployment configuration](kubernetes.yml).

The storage provisioner will be deployed into a new `qnap-storage` namespace. (For Rancher we recommend deploying to `System` project.)

Make sure to change the following properties:

* Image tag for `nineliveslater/qnap-storage-provisioner:master` image by a specific version, e.g. `v1.0.0.k8s-1-18`
* Value of `QNAP_*` environment variables (see below for details)
* Change or remove `shareName` of the *StorageClass* resource
* Create secret `qnap-storage-provisioner` with the following keys: `password`

## Example Persistent Volume Claim

```yaml
kind: PersistentVolumeClaim
apiVersion: v1
metadata:
  name: manual-test
  namespace: default
spec:
  storageClassName: qnap-docker-data1
  accessModes:
    - ReadWriteMany
  resources:
    requests:
      storage: 1Mi
```

This will result in the following directory to be created on the QNAP storage: `/{QNAP_SHARE}/{NAMESPACE}_{NAME}_{PVNAME}/`

In this case it might look sth like this: `/DockerData/default_manual-test_pvc-309624a8-d084-444a-b40a-5a40e6f6e441/`

## Environment Variables

This storage provisioner supports the following environment variables:

| Name | Description | Required? | Example |
| --- | --- | --- | --- |
| `QNAP_URL` | The URL to the QNAP file station API (for setting up NFS share folders). | yes | https://myqnap:443 |
| `QNAP_NFSHOST` | The hostname used for accessing NFS shares by Kubernetes. | yes | myqnap-san |
| `QNAP_SHARE` | The file share name/root folder for accessing NFS shares by Kubernetes. Sub-folders work, if all parent folders exist. Can be overriden via *storage class* parameter. | yes | DockerData |
| `QNAP_USER` | The username used for accessing QNAP file station API. | yes | kubernetes-controller |
| `QNAP_PWD` | The password used for accessing QNAP file station API. | yes | **** |

## Storage Class

The external storage provisioner provides the `qnap/filestation` *storage class*. The following parameters are supported:

| Name | Description | Required? | Example |
| --- | --- | --- | --- |
| shareName | The name of the file share to use; overrides the QNAP_SHARE env variable. | no | MyShare |

## Authors

The library is sponsored by the [marvin + konsorten GmbH](http://www.konsorten.de).

We thank all the authors who provided code to this library:

* Felix Kollmann

## License

[Apache License 2.0](LICENSE)
