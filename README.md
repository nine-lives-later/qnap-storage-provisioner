# Kubernetes Storage Provisioner for QNAP storages

## Deployment

To deploy the Kubernetes Storage Provisioner for QNAP storages, see the [example deployment configuration](kubernetes.yml).

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
| `QNAP_MOUNTOPTIONS` | Colon (`:`) separated list of default NFS mount options to use with volumes. Default specified in the [Dockerfile](Dockerfile). See *Mount Options* below. | no | `hard:fg:udp` |

## Storage Class

The external storage provisioner provides the `qnap/filestation` *storage class*. The following parameters are supported:

| Name | Description | Required? | Example |
| --- | --- | --- | --- |
| shareName | The name of the file share to use; overrides the QNAP_SHARE env variable. | no | MyShare |
| noDefaultMountOptions | Do not apply default mount options. Has no value, keep empty. See below. | no | | 

Mount options specified within the storage class are *added* to the internal default mount options (see below).

## Mount Options

The following default mount options get applied to the NFS volume to maximize performance (see `QNAP_MOUNTOPTIONS` env variable above):

| Option | man page excerpt | Reason |
| --- | --- | --- |
| `hard` | NFS requests are retried indefinitely. | Make the process wait for any IO to complete. See `int` below, too. |
| `fg` | The fg option causes mount(8) to exit with an error status if any part of the mount request times out or fails outright.| Ensure the volume is mounted successfully. It is also the default on Linux. |
| `suid` | Allow set-user-identifier or set-group-identifier bits to take effect. | Allows accessing the volume as root user via sudo. This is required by some Docker images. |
| `nfsvers=3` | The NFS protocol version number used to contact the server's NFS service. | NFSv4 authentication is not supported in this scenario. QNAP supports NFS v3, which has some async and batching features, so use that one. |
| `proto=udp` | In addition to controlling how the NFS client transmits requests to the server, this mount option also controls how the mount(8) command communicates with the server's rpcbind and mountd services. Specifying a netid that uses UDP forces all traffic types to use UDP. | In a stable networking environment, the UDP protocol is way more faster than with NFS. ° |
| `intr` | If intr is specified, system calls return EINTR if an in-progress NFS operation is interrupted by a signal. Using the intr option is preferred to using the soft option because it is significantly less likely to result in data corruption. | When a process is killed (either exit or out-of-memory) interrupt any running NFS operation. The process does not have to wait for the current IO operations to complete before handling the signal. |
| `rsize=8192`<br>`wsize=8192` | The maximum number of bytes in each network READ request that the NFS client can receive when reading data from a file on an NFS server. The actual data payload size of each NFS READ request is equal to or smaller than the rsize setting. | Optimize for jumbo frames and UDP. ° |

> ° NFS man page excerpt: UDP can be quite effective in specialized settings where the networks MTU is large relative to NFSs data transfer size (such as network environments that enable jumbo Ethernet frames). In such environments, trimming the rsize and wsize settings so that each NFS read or write request fits in just a few network frames (or even in a single frame) is advised.

Custom options can be added via settings the mount options on the storage class (see above). 
To not apply any default options at all, set the *noDefaultMountOptions* parameter on the storage class (see above). 

## Authors

We thank all the authors who provided code to this library:

* Felix Kollmann
* marvin + konsorten GmbH (who sponsored this library in 2018)

## License

[Apache License 2.0](LICENSE)
