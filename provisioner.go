/*
Copyright 2018 The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"strings"
	"syscall"

	"github.com/nine-lives-later/go-qnap-filestation"
	"k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/klog"
	"sigs.k8s.io/sig-storage-lib-external-provisioner/v6/controller"
)

const (
	provisionerName = "qnap/filestation"
)

type qnapStorageProvisioner struct {
	StorageURL         string
	StorageNFSHostname string
	StorageUser        string
	StoragePassword    string
	ShareName          string
	MountOptions       []string
}

// ensure interface compatibility
var (
	_ controller.Provisioner      = &qnapStorageProvisioner{}
	_ controller.Qualifier        = &qnapStorageProvisioner{}
	_ controller.DeletionGuard    = &qnapStorageProvisioner{}
	_ controller.BlockProvisioner = &qnapStorageProvisioner{}
)

// NewQnapStorageProvisioner creates a new hostpath provisioner
func NewQnapStorageProvisioner() controller.Provisioner {
	qnapURL := os.Getenv("QNAP_URL")
	if qnapURL == "" {
		klog.Fatal("Failed to retrieve qnap URL from environment variable QNAP_URL")
	}

	qnapNFSHost := os.Getenv("QNAP_NFSHOST")
	if qnapNFSHost == "" {
		klog.Fatal("Failed to retrieve qnap NFS hostname from environment variable QNAP_NFSHOST")
	}

	qnapShare := os.Getenv("QNAP_SHARE")
	if qnapShare == "" {
		klog.Fatal("Failed to retrieve qnap share name from environment variable QNAP_SHARE")
	}

	qnapUser := os.Getenv("QNAP_USER")
	if qnapUser == "" {
		klog.Fatal("Failed to retrieve qnap username from environment variable QNAP_USER")
	}

	qnapPwd := os.Getenv("QNAP_PWD")
	if qnapPwd == "" {
		klog.Fatal("Failed to retrieve qnap user password from environment variable QNAP_PWD")
	}

	qnapMountOptionsStr := os.Getenv("QNAP_MOUNTOPTIONS")
	var qnapMountOptions []string
	if qnapMountOptionsStr != "" {
		qnapMountOptions = strings.SplitN(qnapMountOptionsStr, ":", -1)
	}

	return &qnapStorageProvisioner{
		StorageURL:         qnapURL,
		StorageNFSHostname: qnapNFSHost,
		ShareName:          qnapShare,
		StorageUser:        qnapUser,
		StoragePassword:    qnapPwd,
		MountOptions:       qnapMountOptions,
	}
}

// Provision creates a storage asset and returns a PV object representing it.
func (p *qnapStorageProvisioner) Provision(ctx context.Context, options controller.ProvisionOptions) (*v1.PersistentVolume, controller.ProvisioningState, error) {
	// retrieve config
	shareName := options.StorageClass.Parameters["shareName"]
	if shareName == "" {
		shareName = p.ShareName
	}

	// build mount options
	var mountOptions []string = nil

	if _, ok := options.StorageClass.Parameters["noDefaultMountOptions"]; !ok {
		mountOptions = append(mountOptions, p.MountOptions...)
	}

	if len(options.StorageClass.MountOptions) > 0 {
		mountOptions = append(mountOptions, options.StorageClass.MountOptions...)
	}

	// ensure folder does exist
	folderPath := fmt.Sprintf("/%s/%s_%s_%s", shareName, options.PVC.Namespace, options.PVC.Name, options.PVName)

	klog.Infof("Provisioning persistent volume '%v' on NFS host '%v' in folder '%v' with options %+v", options.PVName, p.StorageNFSHostname, folderPath, mountOptions)

	storage, err := filestation.Connect(p.StorageURL, p.StorageUser, p.StoragePassword, nil)
	defer storage.Logout()

	_, err = storage.EnsureFolder(folderPath)
	if err != nil {
		return nil, controller.ProvisioningNoChange, fmt.Errorf("Failed to ensure storage folder '%v': %v", folderPath, err)
	}

	// build volume information
	pv := &v1.PersistentVolume{
		ObjectMeta: metav1.ObjectMeta{
			Name: options.PVName,
			Annotations: map[string]string{
				"storageName":  p.StorageNFSHostname,
				"storageShare": shareName,
				"storagePath":  folderPath,
			},
		},
		Spec: v1.PersistentVolumeSpec{
			PersistentVolumeReclaimPolicy: *options.StorageClass.ReclaimPolicy,
			AccessModes:                   options.PVC.Spec.AccessModes,
			Capacity: v1.ResourceList{
				v1.ResourceStorage: options.PVC.Spec.Resources.Requests[v1.ResourceStorage],
			},
			MountOptions: mountOptions,
			PersistentVolumeSource: v1.PersistentVolumeSource{
				NFS: &v1.NFSVolumeSource{
					Server:   p.StorageNFSHostname,
					Path:     folderPath,
					ReadOnly: false,
				},
			},
		},
	}

	return pv, controller.ProvisioningFinished, nil
}

// Delete removes the storage asset that was created by Provision represented
// by the given PV.
func (p *qnapStorageProvisioner) Delete(ctx context.Context, volume *v1.PersistentVolume) error {
	// ensure the storage is the same
	if storageName := volume.ObjectMeta.Annotations["storageName"]; !strings.EqualFold(storageName, p.StorageNFSHostname) {
		return &controller.IgnoredError{Reason: "storage name mismatch"}
	}

	// get volume path
	folderPath := volume.ObjectMeta.Annotations["storagePath"]
	if folderPath == "" {
		return &controller.IgnoredError{Reason: "missing storage path annotation (storagePath)"}
	}

	// delete folder from storage
	klog.Infof("Deleting persistent volume '%v' on %v in %v", volume.Name, p.StorageNFSHostname, folderPath)

	storage, err := filestation.Connect(p.StorageURL, p.StorageUser, p.StoragePassword, nil)
	defer storage.Logout()

	_, err = storage.DeleteFile(folderPath)
	if err != nil {
		return fmt.Errorf("failed to delete storage folder '%v': %v", folderPath, err)
	}

	return nil
}

func main() {
	syscall.Umask(0)

	flag.Parse()
	flag.Set("logtostderr", "true")

	// Create an InClusterConfig and use it to create a client for the controller
	// to use to communicate with Kubernetes
	config, err := rest.InClusterConfig()
	if err != nil {
		klog.Fatalf("Failed to create config: %v", err)
	}
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		klog.Fatalf("Failed to create client: %v", err)
	}

	// The controller needs to know what the server version is because out-of-tree
	// provisioners aren't officially supported until 1.5
	serverVersion, err := clientset.Discovery().ServerVersion()
	if err != nil {
		klog.Fatalf("Error getting server version: %v", err)
	}

	// Create the provisioner: it implements the Provisioner interface expected by
	// the controller
	hostPathProvisioner := NewQnapStorageProvisioner()

	// Start the provision controller which will dynamically provision hostPath
	// PVs
	pc := controller.NewProvisionController(clientset, provisionerName, hostPathProvisioner, serverVersion.GitVersion)

	// Never stops.
	pc.Run(context.Background())
}

// ShouldProvision returns whether provisioning for the claim should
// be attempted.
func (p *qnapStorageProvisioner) ShouldProvision(ctx context.Context, vpc *v1.PersistentVolumeClaim) bool {
	// always provision the drive, there is no dependency
	return true
}

// SupportsBlock returns whether provisioner supports block volume.
func (p *qnapStorageProvisioner) SupportsBlock(ctx context.Context) bool {
	// block volumes are not supported via NFS
	return false
}

// ShouldDelete returns whether deleting the PV should be attempted.
func (p *qnapStorageProvisioner) ShouldDelete(ctx context.Context, volume *v1.PersistentVolume) bool {
	// there is no protection for directories, yet
	return true
}
