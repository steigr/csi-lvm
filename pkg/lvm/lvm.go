/*
Copyright 2017 The Kubernetes Authors.

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

package lvm

import (
	"fmt"

	"github.com/container-storage-interface/spec/lib/go/csi"
	"github.com/golang/glog"

	timestamp "github.com/golang/protobuf/ptypes/timestamp"
	"github.com/kubernetes-csi/drivers/pkg/csi-common"
)

const (
	kib    int64 = 1024
	mib    int64 = kib * 1024
	gib    int64 = mib * 1024
	gib100 int64 = gib * 100
	tib    int64 = gib * 1024
	tib100 int64 = tib * 100
)

type lvm struct {
	driver *csicommon.CSIDriver

	ids *identityServer
	ns  *nodeServer
	cs  *controllerServer

	cap   []*csi.VolumeCapability_AccessMode
	cscap []*csi.ControllerServiceCapability
}

type lvmVolume struct {
	VolName string `json:"volName"`
	VolID   string `json:"volID"`
	VolSize int64  `json:"volSize"`
	VolPath string `json:"volPath"`
}

type lvmSnapshot struct {
	Name         string              `json:"name"`
	Id           string              `json:"id"`
	VolID        string              `json:"volID"`
	Path         string              `json:"path"`
	CreationTime timestamp.Timestamp `json:"creationTime"`
	SizeBytes    int64               `json:"sizeBytes"`
	ReadyToUse   bool                `json:"readyToUse"`
}

var lvmVolumes map[string]lvmVolume
var lvmVolumeSnapshots map[string]lvmSnapshot

var (
	lvmDriver *lvm
	vendorVersion  = "dev"
)

func init() {
	lvmVolumes = map[string]lvmVolume{}
	lvmVolumeSnapshots = map[string]lvmSnapshot{}
}

func GetCsiLvmDriver() *lvm {
	return &lvm{}
}

func NewIdentityServer(d *csicommon.CSIDriver) *identityServer {
	return &identityServer{
		DefaultIdentityServer: csicommon.NewDefaultIdentityServer(d),
	}
}

func NewControllerServer(d *csicommon.CSIDriver) *controllerServer {
	return &controllerServer{
		DefaultControllerServer: csicommon.NewDefaultControllerServer(d),
	}
}

func NewNodeServer(d *csicommon.CSIDriver) *nodeServer {
	return &nodeServer{
		DefaultNodeServer: csicommon.NewDefaultNodeServer(d),
	}
}

func (l *lvm) Run(driverName, nodeID, endpoint string) {
	glog.Infof("Driver: %v ", driverName)
	glog.Infof("Version: %s", vendorVersion)

	// Initialize default library driver
	l.driver = csicommon.NewCSIDriver(driverName, vendorVersion, nodeID)
	if l.driver == nil {
		glog.Fatalln("Failed to initialize CSI Driver.")
	}
	l.driver.AddControllerServiceCapabilities(
		[]csi.ControllerServiceCapability_RPC_Type{
			csi.ControllerServiceCapability_RPC_CREATE_DELETE_VOLUME,
			csi.ControllerServiceCapability_RPC_CREATE_DELETE_SNAPSHOT,
			csi.ControllerServiceCapability_RPC_LIST_SNAPSHOTS,
		})
	l.driver.AddVolumeCapabilityAccessModes([]csi.VolumeCapability_AccessMode_Mode{csi.VolumeCapability_AccessMode_SINGLE_NODE_WRITER})

	// Create GRPC servers
	l.ids = NewIdentityServer(l.driver)
	l.ns = NewNodeServer(l.driver)
	l.cs = NewControllerServer(l.driver)

	s := csicommon.NewNonBlockingGRPCServer()
	s.Start(endpoint, l.ids, l.cs, l.ns)
	s.Wait()
}

func getVolumeByID(volumeID string) (lvmVolume, error) {
	if lvmVol, ok := lvmVolumes[volumeID]; ok {
		return lvmVol, nil
	}
	return lvmVolume{}, fmt.Errorf("volume id %s does not exit in the volumes list", volumeID)
}

func getVolumeByName(volName string) (lvmVolume, error) {
	for _, lvmVol := range lvmVolumes {
		if lvmVol.VolName == volName {
			return lvmVol, nil
		}
	}
	return lvmVolume{}, fmt.Errorf("volume name %s does not exit in the volumes list", volName)
}

func getSnapshotByName(name string) (lvmSnapshot, error) {
	for _, snapshot := range lvmVolumeSnapshots {
		if snapshot.Name == name {
			return snapshot, nil
		}
	}
	return lvmSnapshot{}, fmt.Errorf("snapshot name %s does not exit in the snapshots list", name)
}
