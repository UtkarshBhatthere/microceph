package ceph

import (
	"context"
	"encoding/json"
	"fmt"
	"net"

	"github.com/canonical/microceph/microceph/api/types"
	"github.com/canonical/microceph/microceph/database"
	"github.com/canonical/microceph/microceph/interfaces"
)

type IngressServicePlacement struct {
	types.IngressServicePlacement

	// These fields will be populated by EnableIngress
	vrrpPassword string
	vrrpRouterID int
}

func (isp *IngressServicePlacement) PopulateParams(s interfaces.StateInterface, payload string) error {
	err := json.Unmarshal([]byte(payload), &isp.IngressServicePlacement)
	if err != nil {
		return err
	}
	return isp.Validate()
}

func (isp *IngressServicePlacement) HospitalityCheck(s interfaces.StateInterface) error {
	ifaces, err := net.Interfaces()
	if err != nil {
		return fmt.Errorf("failed to get network interfaces: %w", err)
	}

	for _, i := range ifaces {
		if i.Name == isp.VIPInterface {
			return genericHospitalityCheck("ingress")
		}
	}

	return fmt.Errorf("network interface '%s' not found", isp.VIPInterface)
}

// ServiceInit will call EnableIngress, which will generate or fetch VRRP params
// and store them in the isp struct for DbUpdate to use.
func (isp *IngressServicePlacement) ServiceInit(ctx context.Context, s interfaces.StateInterface) error {
	return EnableIngress(ctx, s, isp)
}

func (isp *IngressServicePlacement) PostPlacementCheck(s interfaces.StateInterface) error {
	return genericPostPlacementCheck("ingress")
}

func (isp *IngressServicePlacement) DbUpdate(ctx context.Context, s interfaces.StateInterface) error {
	groupConfig := database.IngressServiceGroupConfig{
		VIPAddress:   isp.VIPAddress,
		VIPInterface: isp.VIPInterface,
		Target:       isp.Target,
		VRRPPassword: isp.vrrpPassword,
		VRRPRouterID: isp.vrrpRouterID,
	}
	serviceInfo := database.IngressServiceInfo{}

	return database.GroupedServicesQuery.AddNew(ctx, s, "ingress", isp.ServiceID, groupConfig, serviceInfo)
}
