package main

import (
	"context"
	"encoding/json"

	"github.com/canonical/microcluster/v2/microcluster"
	"github.com/spf13/cobra"

	"github.com/canonical/microceph/microceph/api/types"
	"github.com/canonical/microceph/microceph/client"
)

type cmdEnableIngress struct {
	common           *CmdControl
	wait             bool
	flagServiceID    string
	flagVIPAddress   string
	flagVIPInterface string
	flagTarget       string
	flagTargetNode   string
}

func (c *cmdEnableIngress) Command() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "ingress --service-id <id> --vip-address <ip> --vip-interface <iface> --target <service.id>",
		Short: "Enable the ingress service on a target node",
		RunE:  c.Run,
	}
	cmd.PersistentFlags().StringVar(&c.flagServiceID, "service-id", "", "Ingress Service ID")
	cmd.PersistentFlags().StringVar(&c.flagVIPAddress, "vip-address", "", "Virtual IP address")
	cmd.PersistentFlags().StringVar(&c.flagVIPInterface, "vip-interface", "", "Network interface for the VIP")
	cmd.PersistentFlags().StringVar(&c.flagTarget, "target", "", "The service to provide ingress for (e.g. nfs.my-nfs-cluster)")
	cmd.PersistentFlags().StringVar(&c.flagTargetNode, "target-node", "", "Node to enable ingress on (default: this server)")
	cmd.Flags().BoolVar(&c.wait, "wait", true, "Wait for ingress service to be up")
	return cmd
}

// Run handles the enable ingress command.
func (c *cmdEnableIngress) Run(cmd *cobra.Command, args []string) error {
	obj := types.IngressServicePlacement{
		ServiceID:    c.flagServiceID,
		VIPAddress:   c.flagVIPAddress,
		VIPInterface: c.flagVIPInterface,
		Target:       c.flagTarget,
	}

	if err := obj.Validate(); err != nil {
		return err
	}

	jsp, err := json.Marshal(obj)
	if err != nil {
		return err
	}

	req := &types.EnableService{
		Name:    "ingress",
		Wait:    c.wait,
		Payload: string(jsp[:]),
	}

	m, err := microcluster.App(microcluster.Args{StateDir: c.common.FlagStateDir})
	if err != nil {
		return err
	}

	cli, err := m.LocalClient()
	if err != nil {
		return err
	}

	return client.SendServicePlacementReq(context.Background(), cli, req, c.flagTargetNode)
}
