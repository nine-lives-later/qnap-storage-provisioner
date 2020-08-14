# QNAP Storage Provisioner for QNAP storages

## Environment Variables

This storage provisioner supports the following environment variables:

| Name | Description | Required? | Example |
| --- | --- | --- | --- |
| `QNAP_URL` | The URL to the QNAP file station API (for setting up NFS share folders). | yes | https://myqnap:443 |
| `QNAP_NFSHOST` | The hostname used for accessing NFS shares by Kubernetes. | yes | myqnap-san |
| `QNAP_SHARE` | The file share name/root folder for accessing NFS shares by Kubernetes. Can be overriden via *storage class* parameter. | yes | DockerData |
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
