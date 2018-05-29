// Copyright 2018 Canonical Ltd.
// Licensed under the AGPLv3, see LICENCE file for details.

package admincmd

import (
	"encoding/json"
	"fmt"

	"github.com/juju/cmd"
	"gopkg.in/errgo.v1"
	"gopkg.in/macaroon-bakery.v2/bakery"
)

type generateKeysCommand struct {
	*candidCommand
}

func newGenerateKeysCommand(cc *candidCommand) cmd.Command {
	c := &generateKeysCommand{}
	c.candidCommand = cc
	return c
}

var generateKeysDoc = `
The generate-keys command generates a public/private keypair that can be used
in candid configuration.
`

func (c *generateKeysCommand) Info() *cmd.Info {
	return &cmd.Info{
		Name:    "generate-keys",
		Purpose: "generate a public/private keypair",
		Doc:     generateKeysDoc,
	}
}

func (c *generateKeysCommand) Init(args []string) error {
	return errgo.Mask(c.candidCommand.Init(nil))
}

func (c *generateKeysCommand) Run(ctxt *cmd.Context) error {
	defer c.Close(ctxt)

	keyPair, err := bakery.GenerateKey()
	if err != nil {
		return errgo.Mask(err)
	}

	text, err := json.MarshalIndent(keyPair, "", "\t")
	if err != nil {
		return errgo.Mask(err)
	}
	fmt.Fprintf(ctxt.Stdout, "%s\n", text)
	return nil
}
