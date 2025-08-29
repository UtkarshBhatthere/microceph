package ceph

import (
	"context"
	"crypto/rand"
	"database/sql"
	"encoding/json"
	"fmt"
	"math/big"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/canonical/lxd/shared/api"
	"github.com/canonical/lxd/shared/revert"

	"github.com/canonical/microceph/microceph/constants"
	"github.com/canonical/microceph/microceph/database"
	"github.com/canonical/microceph/microceph/interfaces"
	"github.com/canonical/microceph/microceph/logger"
)

const (
	keepalivedConfigTemplate = `
{{range .}}
vrrp_instance VI_{{.Placement.ServiceID}} {
    state MASTER
    interface {{.Placement.VIPInterface}}
    virtual_router_id {{.VRRPRouterID}}
    priority 101
    advert_int 1
    authentication {
        auth_type PASS
        auth_pass {{.VRRPPassword}}
    }
    virtual_ipaddress {
        {{.Placement.VIPAddress}}
    }
}
{{end}}
`
	haproxyConfigTemplate = `
global
    log /dev/log local0
    log /dev/log local1 notice
    chroot /var/lib/haproxy
    stats socket /run/haproxy/admin.sock mode 660 level admin expose-fd listeners
    stats timeout 30s
    user haproxy
    group haproxy
    daemon

defaults
    log     global
    mode    tcp
    option  tcplog
    option  dontlognull
    timeout connect 5000
    timeout client  50000
    timeout server  50000

frontend main
    bind *:2049
    mode tcp
    {{range .}}
    acl is_{{.Placement.ServiceID}} dst {{.Placement.VIPAddress}}
    use_backend backend_{{.Placement.ServiceID}} if is_{{.Placement.ServiceID}}
    {{end}}

{{range .}}
backend backend_{{.Placement.ServiceID}}
    mode tcp
    balance roundrobin
    {{range .BackendServers}}
    server {{.Name}} {{.Address}} check
    {{end}}
{{end}}
`
)

type backendServer struct {
	Name    string
	Address string
}

type ingressTemplateData struct {
	Placement      *database.IngressServiceGroupConfig
	BackendServers []backendServer
}

// EnableIngress enables or updates the ingress service on the host.
func EnableIngress(ctx context.Context, s interfaces.StateInterface, isp *IngressServicePlacement) error {
	logger.Debugf("Enabling ingress service with ServiceID '%s'", isp.ServiceID)

	// Check if service group already exists and generate/fetch VRRP params
	err := s.ClusterState().Database().Transaction(ctx, func(ctx context.Context, tx *sql.Tx) error {
		group, err := database.GetServiceGroup(ctx, tx, "ingress", isp.ServiceID)
		if err != nil && !api.StatusErrorCheck(err, http.StatusNotFound) {
			return fmt.Errorf("failed to get service group: %w", err)
		}

		if group != nil {
			// Group exists, this is an update or re-enablement.
			// We don't need to do anything here, the config will be regenerated later.
		} else {
			// Group doesn't exist, generate new VRRP params
			password, err := generateRandomString(16)
			if err != nil {
				return fmt.Errorf("failed to generate VRRP password: %w", err)
			}
			isp.vrrpPassword = password

			routerID, err := rand.Int(rand.Reader, big.NewInt(254))
			if err != nil {
				return fmt.Errorf("failed to generate VRRP router ID: %w", err)
			}
			isp.vrrpRouterID = int(routerID.Int64()) + 1 // 1-255
		}
		return nil
	})
	if err != nil {
		return err
	}

	return regenerateIngressConfigs(ctx, s)
}

// DisableIngress disables an ingress service instance on the host.
func DisableIngress(ctx context.Context, s interfaces.StateInterface, serviceID string) error {
	logger.Debugf("Disabling ingress service with ServiceID '%s'", serviceID)

	// Remove database records.
	err := database.GroupedServicesQuery.RemoveForHost(ctx, s, "ingress", serviceID)
	if err != nil {
		return err
	}

	return regenerateIngressConfigs(ctx, s)
}

func regenerateIngressConfigs(ctx context.Context, s interfaces.StateInterface) error {
	// Get all ingress service groups
	ingressGroups, err := database.GroupedServicesQuery.GetGroupedServices(ctx, s, database.GroupedServiceFilter{Service: "ingress"})
	if err != nil {
		return fmt.Errorf("failed to get ingress service groups: %w", err)
	}

	if len(ingressGroups) == 0 {
		// No more ingress services, stop the service and remove configs
		logger.Debugf("No more ingress services, stopping service.")
		err := snapStop("ingress", true)
		if err != nil {
			return err
		}
		pathConsts := constants.GetPathConst()
		ingressConfDir := filepath.Join(pathConsts.ConfPath, "ingress")
		return os.RemoveAll(ingressConfDir)
	}

	var allTemplateData []ingressTemplateData

	for _, group := range ingressGroups {
		var config database.IngressServiceGroupConfig
		err := json.Unmarshal([]byte(group.Config), &config)
		if err != nil {
			return fmt.Errorf("failed to unmarshal ingress group config for %s: %w", group.GroupID, err)
		}

		parts := strings.Split(config.Target, ".")
		targetService := parts[0]
		targetGroupID := parts[1]

		allServices, err := database.GroupedServicesQuery.GetGroupedServices(ctx, s, database.GroupedServiceFilter{Service: &targetService, GroupID: &targetGroupID})
		if err != nil {
			return fmt.Errorf("failed to get target service endpoints for %s: %w", config.Target, err)
		}

		var backendServers []backendServer
		for _, ts := range allServices {
			var nfsInfo database.NFSServiceInfo
			err := json.Unmarshal([]byte(ts.Info), &nfsInfo)
			if err != nil {
				return fmt.Errorf("failed to unmarshal service info for member %s: %w", ts.Member, err)
			}
			member, err := s.ClusterState().GetMember(ts.Member)
			if err != nil {
				return fmt.Errorf("failed to get member %s: %w", ts.Member, err)
			}
			backendServers = append(backendServers, backendServer{Name: ts.Member, Address: fmt.Sprintf("%s:%d", member.Address, nfsInfo.BindPort)})
		}

		allTemplateData = append(allTemplateData, ingressTemplateData{
			Placement:      &config,
			BackendServers: backendServers,
		})
	}

	pathConsts := constants.GetPathConst()
	ingressConfDir := filepath.Join(pathConsts.ConfPath, "ingress")
	err = os.MkdirAll(ingressConfDir, 0755)
	if err != nil && !os.IsExist(err) {
		return err
	}

	// Create keepalived.conf
	keepalivedConfPath := filepath.Join(ingressConfDir, "keepalived.conf")
	err = writeTemplate(keepalivedConfPath, keepalivedConfigTemplate, allTemplateData)
	if err != nil {
		return fmt.Errorf("failed to write keepalived config: %w", err)
	}

	// Create haproxy.cfg
	haproxyConfPath := filepath.Join(ingressConfDir, "haproxy.cfg")
	err = writeTemplate(haproxyConfPath, haproxyConfigTemplate, allTemplateData)
	if err != nil {
		return fmt.Errorf("failed to write haproxy config: %w", err)
	}

	// Start or reload the ingress service.
	err = snapCheckActive("ingress")
	if err == nil {
		return snapRestart("ingress", true)
	}
	return snapStart("ingress", true)
}

func writeTemplate(path, tmpl string, data any) error {
	t, err := template.New(filepath.Base(path)).Parse(tmpl)
	if err != nil {
		return err
	}

	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()

	return t.Execute(f, data)
}

func generateRandomString(n int) (string, error) {
	const letters = "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz"
	ret := make([]byte, n)
	for i := 0; i < n; i++ {
		num, err := rand.Int(rand.Reader, big.NewInt(int64(len(letters))))
		if err != nil {
			return "", err
		}
		ret[i] = letters[num.Int64()]
	}
	return string(ret), nil
}
