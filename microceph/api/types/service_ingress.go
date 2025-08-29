package types

import (
	"fmt"
	"net"
	"regexp"
	"strings"
)

// IngressServiceIDRegex is the regular expression that a valid Ingress ServiceID must match.
var IngressServiceIDRegex = regexp.MustCompile(`^[a-zA-Z0-9.\-_]+$`)

// IngressServicePlacement represents the configuration for an ingress service.
type IngressServicePlacement struct {
	ServiceID    string `json:"service_id"`
	VIPAddress   string `json:"vip_address"`
	VIPInterface string `json:"vip_interface"`
	Target       string `json:"target"`
}

// Validate checks if the IngressServicePlacement has valid fields.
func (isp *IngressServicePlacement) Validate() error {
	if !IngressServiceIDRegex.MatchString(isp.ServiceID) {
		return fmt.Errorf("expected service_id to be valid (regex: '%s')", IngressServiceIDRegex.String())
	}
	if net.ParseIP(isp.VIPAddress) == nil {
		return fmt.Errorf("vip_address '%s' could not be parsed", isp.VIPAddress)
	}
	if isp.VIPInterface == "" {
		return fmt.Errorf("vip_interface must be provided")
	}
	if isp.Target == "" {
		return fmt.Errorf("target must be provided")
	}
	parts := strings.Split(isp.Target, ".")
	if len(parts) != 2 {
		return fmt.Errorf("target must be in the format <service>.<id>")
	}
	return nil
}
