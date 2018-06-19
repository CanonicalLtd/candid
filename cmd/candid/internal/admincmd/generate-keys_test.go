// Copyright 2016 Canonical Ltd.
// Licensed under the AGPLv3, see LICENCE file for details.

package admincmd_test

import (
	"encoding/json"

	gc "gopkg.in/check.v1"
)

type generateKeysSuite struct {
	commandSuite
}

var _ = gc.Suite(&generateKeysSuite{})

func (s *generateKeysSuite) TestAddGroup(c *gc.C) {
	stdout := CheckSuccess(c, s.Run, "generate-keys")
	var data interface{}
	err := json.Unmarshal([]byte(stdout), &data)
	c.Assert(err, gc.IsNil)
	c.Assert(data.(map[string]interface{})["private"], gc.Not(gc.Equals), "")
	c.Assert(data.(map[string]interface{})["public"], gc.Not(gc.Equals), "")
}
