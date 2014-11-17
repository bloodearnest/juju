// Copyright 2013 Canonical Ltd.
// Licensed under the AGPLv3, see LICENCE file for details.

package main

import (
	"fmt"

	"github.com/juju/cmd"

	"github.com/juju/juju/cmd/envcmd"
	"github.com/juju/juju/environs/config"
	"strings"
)

// ProtectionCommand is a super for environment protection commands that block/unblock operations.
type ProtectionCommand struct {
	envcmd.EnvCommandBase
	operation string
	desc      string
}

var (
	blockArgs = "destroy-environment"
)

// setBlockEnvironmentVariable sets desired environment variable to given value
func (p *ProtectionCommand) setBlockEnvironmentVariable(block bool) error {
	client, err := p.NewAPIClient()
	if err != nil {
		return err
	}
	defer client.Close()
	attrs := map[string]interface{}{config.BlockKeyPrefix + p.operation: block}
	return client.EnvironmentSet(attrs)
}

// assignValidOperation verifies that supplied operation is supported.
func (p *ProtectionCommand) assignValidOperation(cmd string, args []string) error {
	if len(args) != 1 {
		return fmt.Errorf("must specify operation (%v) to %v", blockArgs, cmd)
	}
	var err error
	p.operation, err = p.obtainValidArgument(args[0])
	return err
}

// obtainValidArgument returns polished argument:
// it checks that the argument is a supported operation and
// forces it into lower case for consistency
func (p *ProtectionCommand) obtainValidArgument(arg string) (string, error) {
	for _, valid := range strings.Split(blockArgs, "|") {
		if strings.EqualFold(valid, arg) {
			return strings.ToLower(arg), nil
		}
	}
	return "", fmt.Errorf("%q is not a valid argument: use one of %v", arg, blockArgs)
}

// BlockCommand blocks specified operation.
type BlockCommand struct {
	ProtectionCommand
}

var blockDoc = `

Juju allows to safeguard deployed environments from unintentional damage by preventing
execution of operations that could alter environment.

This is done by blocking certain operations from successful execution. Blocked operations
must be manually unblocked to proceed.

Operations that can be blocked are

destroy environment


Examples:
   juju block destroy-environment      (blocks destroy environment)

See Also:
   juju help unblock
`

func (c *BlockCommand) Info() *cmd.Info {
	return &cmd.Info{
		Name:    "block",
		Args:    blockArgs,
		Purpose: "block operation",
		Doc:     blockDoc,
	}
}

func (c *BlockCommand) Init(args []string) error {
	return c.assignValidOperation("block", args)
}

func (c *BlockCommand) Run(_ *cmd.Context) error {
	return c.setBlockEnvironmentVariable(true)
}
