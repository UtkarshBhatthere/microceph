package ceph

import (
	"fmt"

	"github.com/lxc/lxd/shared"
)

func snapStart(service string, enable bool) error {
	args := []string{
		"start",
		fmt.Sprintf("microceph.%s", service),
	}

	if enable {
		args = append(args, "--enable")
	}

	_, err := shared.RunCommand("snapctl", args...)
	if err != nil {
		return err
	}

	return nil
}

func snapReload(service string) error {
	args := []string{
		"restart",
		"--reload",
		fmt.Sprintf("microceph.%s", service),
	}

	_, err := shared.RunCommand("snapctl", args...)
	if err != nil {
		return err
	}

	return nil
}
