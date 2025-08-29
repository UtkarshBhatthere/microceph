package main

import (
	"context"
	"fmt"

	"github.com/canonical/microcluster/v2/microcluster"
	"github.com/spf13/cobra"

	"github.com/canonical/microceph/microceph/api/types"
	"github.com/canonical/microceph/microceph/client"
)

type cmdDisableIngress struct {
	common        *CmdControl
	flagServiceID string
	flagTargetNode string
}

func (c *cmdDisableIngress) Command() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "ingress --service-id <service-id> [--target-node <server>]",
		Short: "Disable the ingress service on a target node",
		RunE:  c.Run,
	}
	cmd.PersistentFlags().StringVar(&c.flagServiceID, "service-id", "", "Ingress Service ID")
	cmd.PersistentFlags().StringVar(&c.flagTargetNode, "target-node", "", "Node to disable ingress on (default: this server)")
	return cmd
}

// Run handles the disable ingress command.
func (c *cmdDisableIngress) Run(cmd *cobra.Command, args []string) error {
	if !types.IngressServiceIDRegex.MatchString(c.flagServiceID) {
		return fmt.Errorf("please provide a valid service ID using the `--service-id` flag")
	}

	m, err := microcluster.App(microcluster.Args{StateDir: c.common.FlagStateDir})
	if err != nil {
		return err
	}

	cli, err := m.LocalClient()
	if err != nil {
		return err
	}

	svc := &types.IngressService{ClusterID: c.flagServiceID}
	err = client.DeleteIngressService(context.Background(), cli, c.flagTargetNode, svc)
	if err != nil {
		return err
	}

	return nil
}
