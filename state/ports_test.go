// Copyright 2014 Canonical Ltd.
// Licensed under the AGPLv3, see LICENCE file for details.

package state_test

import (
	"strings"

	"github.com/juju/errors"
	jc "github.com/juju/testing/checkers"
	gc "launchpad.net/gocheck"

	"github.com/juju/juju/network"
	"github.com/juju/juju/state"
	statetesting "github.com/juju/juju/state/testing"
	"github.com/juju/juju/testing/factory"
)

type PortsDocSuite struct {
	ConnSuite
	charm   *state.Charm
	service *state.Service
	unit1   *state.Unit
	unit2   *state.Unit
	machine *state.Machine
	ports   *state.Ports
}

var _ = gc.Suite(&PortsDocSuite{})

func (s *PortsDocSuite) SetUpTest(c *gc.C) {
	s.ConnSuite.SetUpTest(c)

	f := factory.NewFactory(s.State)
	s.charm = f.MakeCharm(c, &factory.CharmParams{Name: "wordpress"})
	s.service = f.MakeService(c, &factory.ServiceParams{Name: "wordpress", Charm: s.charm})
	s.machine = f.MakeMachine(c, &factory.MachineParams{Series: "quantal"})
	s.unit1 = f.MakeUnit(c, &factory.UnitParams{Service: s.service, Machine: s.machine})
	s.unit2 = f.MakeUnit(c, &factory.UnitParams{Service: s.service, Machine: s.machine})

	var err error
	s.ports, err = state.GetOrCreatePorts(s.State, s.machine.Id(), network.DefaultPublic)
	c.Assert(err, gc.IsNil)
	c.Assert(s.ports, gc.NotNil)
}

func (s *PortsDocSuite) TestCreatePorts(c *gc.C) {
	ports, err := state.GetOrCreatePorts(s.State, s.machine.Id(), network.DefaultPublic)
	c.Assert(err, gc.IsNil)
	c.Assert(ports, gc.NotNil)
	err = ports.OpenPorts(state.PortRange{
		FromPort: 100,
		ToPort:   200,
		UnitName: s.unit1.Name(),
		Protocol: "TCP",
	})
	c.Assert(err, gc.IsNil)

	ports, err = state.GetPorts(s.State, s.machine.Id(), network.DefaultPublic)
	c.Assert(err, gc.IsNil)
	c.Assert(ports, gc.NotNil)

	c.Assert(ports.PortsForUnit(s.unit1.Name()), gc.HasLen, 1)
}

func (s *PortsDocSuite) TestOpenAndClosePorts(c *gc.C) {

	testCases := []struct {
		about    string
		existing []state.PortRange
		open     *state.PortRange
		close    *state.PortRange
		expected string
	}{{
		about:    "open and close same port range",
		existing: nil,
		open: &state.PortRange{
			FromPort: 100,
			ToPort:   200,
			UnitName: s.unit1.Name(),
			Protocol: "TCP",
		},
		close: &state.PortRange{
			FromPort: 100,
			ToPort:   200,
			UnitName: s.unit1.Name(),
			Protocol: "TCP",
		},
		expected: "",
	}, {
		about: "try to close part of a port range",
		existing: []state.PortRange{{
			FromPort: 100,
			ToPort:   200,
			UnitName: s.unit1.Name(),
			Protocol: "TCP",
		}},
		open: nil,
		close: &state.PortRange{
			FromPort: 100,
			ToPort:   150,
			UnitName: s.unit1.Name(),
			Protocol: "TCP",
		},
		expected: "mismatched port ranges: port ranges 100-200/tcp \\(wordpress/0\\) and 100-150/tcp \\(wordpress/0\\) conflict",
	}, {
		about: "close an unopened port range with existing clash from other unit",
		existing: []state.PortRange{{
			FromPort: 100,
			ToPort:   150,
			UnitName: s.unit1.Name(),
			Protocol: "TCP",
		}},
		open: nil,
		close: &state.PortRange{
			FromPort: 100,
			ToPort:   150,
			UnitName: s.unit1.Name(),
			Protocol: "TCP",
		},
		expected: "",
	}, {
		about:    "close an unopened port range",
		existing: nil,
		open:     nil,
		close: &state.PortRange{
			FromPort: 100,
			ToPort:   150,
			UnitName: s.unit1.Name(),
			Protocol: "TCP",
		},
		expected: "",
	}, {
		about: "try to close an overlapping port range",
		existing: []state.PortRange{{
			FromPort: 100,
			ToPort:   200,
			UnitName: s.unit1.Name(),
			Protocol: "TCP",
		}},
		open: nil,
		close: &state.PortRange{
			FromPort: 100,
			ToPort:   300,
			UnitName: s.unit1.Name(),
			Protocol: "TCP",
		},
		expected: "mismatched port ranges: port ranges 100-200/tcp \\(wordpress/0\\) and 100-300/tcp \\(wordpress/0\\) conflict",
	}, {
		about: "try to open an overlapping port range with different unit",
		existing: []state.PortRange{{
			FromPort: 100,
			ToPort:   200,
			UnitName: s.unit1.Name(),
			Protocol: "TCP",
		}},
		open: &state.PortRange{
			FromPort: 100,
			ToPort:   300,
			UnitName: s.unit2.Name(),
			Protocol: "TCP",
		},
		expected: "cannot open ports 100-300/tcp on machine 0: port ranges 100-200/tcp \\(wordpress/0\\) and 100-300/tcp \\(wordpress/1\\) conflict",
	}, {
		about: "try to open an identical port range with different unit",
		existing: []state.PortRange{{
			FromPort: 100,
			ToPort:   200,
			UnitName: s.unit1.Name(),
			Protocol: "TCP",
		}},
		open: &state.PortRange{
			FromPort: 100,
			ToPort:   200,
			UnitName: s.unit2.Name(),
			Protocol: "TCP",
		},
		expected: "cannot open ports 100-200/tcp on machine 0: port ranges 100-200/tcp \\(wordpress/0\\) and 100-200/tcp \\(wordpress/1\\) conflict",
	}, {
		about: "try to open a port range with different protocol with different unit",
		existing: []state.PortRange{{
			FromPort: 100,
			ToPort:   200,
			UnitName: s.unit1.Name(),
			Protocol: "TCP",
		}},
		open: &state.PortRange{
			FromPort: 100,
			ToPort:   200,
			UnitName: s.unit2.Name(),
			Protocol: "UDP",
		},
		expected: "",
	}, {
		about: "try to open a non-overlapping port range with different unit",
		existing: []state.PortRange{{
			FromPort: 100,
			ToPort:   200,
			UnitName: s.unit1.Name(),
			Protocol: "TCP",
		}},
		open: &state.PortRange{
			FromPort: 300,
			ToPort:   400,
			UnitName: s.unit2.Name(),
			Protocol: "TCP",
		},
		expected: "",
	}}

	for i, t := range testCases {
		c.Logf("test %d: %s", i, t.about)

		ports, err := state.GetOrCreatePorts(s.State, s.machine.Id(), network.DefaultPublic)
		c.Assert(err, gc.IsNil)
		c.Assert(ports, gc.NotNil)

		// open ports that should exist for the test case
		for _, portRange := range t.existing {
			err := ports.OpenPorts(portRange)
			c.Check(err, gc.IsNil)
		}
		if t.existing != nil {
			err = ports.Refresh()
			c.Check(err, gc.IsNil)
		}
		if t.open != nil {
			err = ports.OpenPorts(*t.open)
			if t.expected == "" {
				c.Check(err, gc.IsNil)
			} else {
				c.Check(err, gc.ErrorMatches, t.expected)
			}
			err = ports.Refresh()
			c.Check(err, gc.IsNil)

		}

		if t.close != nil {
			err := ports.ClosePorts(*t.close)
			if t.expected == "" {
				c.Check(err, gc.IsNil)
			} else {
				c.Check(err, gc.ErrorMatches, t.expected)
			}
		}
		err = ports.Remove()
		c.Check(err, gc.IsNil)
	}
}

func (s *PortsDocSuite) TestAllPortRanges(c *gc.C) {
	portRange := state.PortRange{
		FromPort: 100,
		ToPort:   200,
		UnitName: s.unit1.Name(),
		Protocol: "TCP",
	}
	err := s.ports.OpenPorts(portRange)
	c.Assert(err, gc.IsNil)

	ranges := s.ports.AllPortRanges()
	c.Assert(ranges, gc.HasLen, 1)

	c.Assert(ranges[network.PortRange{100, 200, "TCP"}], gc.Equals, s.unit1.Name())
}

func (s *PortsDocSuite) TestOpenInvalidRange(c *gc.C) {
	portRange := state.PortRange{
		FromPort: 400,
		ToPort:   200,
		UnitName: s.unit1.Name(),
		Protocol: "TCP",
	}
	err := s.ports.OpenPorts(portRange)
	c.Assert(err, gc.ErrorMatches, "invalid port range .*")
}

func (s *PortsDocSuite) TestCloseInvalidRange(c *gc.C) {
	portRange := state.PortRange{
		FromPort: 100,
		ToPort:   200,
		UnitName: s.unit1.Name(),
		Protocol: "TCP",
	}
	err := s.ports.OpenPorts(portRange)
	c.Assert(err, gc.IsNil)

	err = s.ports.Refresh()
	c.Assert(err, gc.IsNil)
	err = s.ports.ClosePorts(state.PortRange{
		FromPort: 150,
		ToPort:   200,
		UnitName: s.unit1.Name(),
		Protocol: "TCP",
	})
	c.Assert(err, gc.ErrorMatches, "mismatched port ranges: port ranges 100-200/tcp \\(wordpress/0\\) and 150-200/tcp \\(wordpress/0\\) conflict")
}

func (s *PortsDocSuite) TestRemovePortsDoc(c *gc.C) {
	portRange := state.PortRange{
		FromPort: 100,
		ToPort:   200,
		UnitName: s.unit1.Name(),
		Protocol: "TCP",
	}
	err := s.ports.OpenPorts(portRange)
	c.Assert(err, gc.IsNil)

	ports, err := state.GetPorts(s.State, s.machine.Id(), network.DefaultPublic)
	c.Assert(err, gc.IsNil)
	c.Assert(ports, gc.NotNil)

	allPorts, err := s.machine.AllPorts()
	c.Assert(err, gc.IsNil)

	for _, prt := range allPorts {
		err := prt.Remove()
		c.Assert(err, gc.IsNil)
	}

	ports, err = state.GetPorts(s.State, s.machine.Id(), network.DefaultPublic)
	c.Assert(ports, gc.IsNil)
	c.Assert(err, jc.Satisfies, errors.IsNotFound)
	c.Assert(err, gc.ErrorMatches, `ports for machine 0, network "juju-public" not found`)
}

func (s *PortsDocSuite) TestWatchPorts(c *gc.C) {
	w := s.State.WatchOpenedPorts()
	c.Assert(w, gc.NotNil)

	defer statetesting.AssertStop(c, w)
	wc := statetesting.NewStringsWatcherC(c, s.State, w)
	wc.AssertChange()
	wc.AssertNoChange()

	portRange := state.PortRange{
		FromPort: 100,
		ToPort:   200,
		UnitName: s.unit1.Name(),
		Protocol: "TCP",
	}
	globalKey := state.PortsGlobalKey(s.machine.Id(), network.DefaultPublic)
	err := s.ports.OpenPorts(portRange)
	c.Assert(err, gc.IsNil)
	wc.AssertChange(globalKey)

	err = s.ports.Refresh()
	c.Assert(err, gc.IsNil)
	err = s.ports.ClosePorts(portRange)
	c.Assert(err, gc.IsNil)
	wc.AssertChange(globalKey)
}

type PortRangeSuite struct{}

var _ = gc.Suite(&PortRangeSuite{})

// Create a port range or panic if it is invalid.
func MustPortRange(unitName string, fromPort, toPort int, protocol string) state.PortRange {
	portRange, err := state.NewPortRange(unitName, fromPort, toPort, protocol)
	if err != nil {
		panic(err)
	}
	return portRange
}

func (p *PortRangeSuite) TestPortRangeConflicts(c *gc.C) {
	var testCases = []struct {
		about    string
		first    state.PortRange
		second   state.PortRange
		expected interface{}
	}{{
		"identical ports",
		MustPortRange("wordpress/0", 80, 80, "TCP"),
		MustPortRange("wordpress/0", 80, 80, "TCP"),
		nil,
	}, {
		"identical port ranges",
		MustPortRange("wordpress/0", 80, 100, "TCP"),
		MustPortRange("wordpress/0", 80, 100, "TCP"),
		nil,
	}, {
		"different ports",
		MustPortRange("wordpress/0", 80, 80, "TCP"),
		MustPortRange("wordpress/0", 90, 90, "TCP"),
		nil,
	}, {
		"touching ranges",
		MustPortRange("wordpress/0", 100, 200, "TCP"),
		MustPortRange("wordpress/0", 201, 240, "TCP"),
		nil,
	}, {
		"touching ranges with overlap",
		MustPortRange("wordpress/0", 100, 200, "TCP"),
		MustPortRange("wordpress/0", 200, 240, "TCP"),
		"port ranges .* conflict",
	}, {
		"identical ports with different protocols",
		MustPortRange("wordpress/0", 80, 80, "UDP"),
		MustPortRange("wordpress/0", 80, 80, "TCP"),
		nil,
	}, {
		"overlapping ranges with different protocols",
		MustPortRange("wordpress/0", 80, 200, "UDP"),
		MustPortRange("wordpress/0", 80, 300, "TCP"),
		nil,
	}, {
		"outside range",
		MustPortRange("wordpress/0", 100, 200, "TCP"),
		MustPortRange("wordpress/0", 80, 80, "TCP"),
		nil,
	}, {
		"overlap end",
		MustPortRange("wordpress/0", 100, 200, "TCP"),
		MustPortRange("wordpress/0", 80, 120, "TCP"),
		"port ranges .* conflict",
	}, {
		"complete overlap",
		MustPortRange("wordpress/0", 100, 200, "TCP"),
		MustPortRange("wordpress/0", 120, 140, "TCP"),
		"port ranges .* conflict",
	}, {
		"overlap with same end",
		MustPortRange("wordpress/0", 100, 200, "TCP"),
		MustPortRange("wordpress/0", 120, 200, "TCP"),
		"port ranges .* conflict",
	}, {
		"overlap with same start",
		MustPortRange("wordpress/0", 100, 200, "TCP"),
		MustPortRange("wordpress/0", 100, 120, "TCP"),
		"port ranges .* conflict",
	}, {
		"invalid port range",
		state.PortRange{"wordpress/0", 100, 80, "TCP"},
		MustPortRange("wordpress/0", 80, 80, "TCP"),
		"invalid port range 100-80",
	}, {
		"different units, same port",
		MustPortRange("mysql/0", 80, 80, "TCP"),
		MustPortRange("wordpress/0", 80, 80, "TCP"),
		"port ranges .* conflict",
	}, {
		"different units, different port ranges",
		MustPortRange("mysql/0", 80, 100, "TCP"),
		MustPortRange("wordpress/0", 180, 280, "TCP"),
		nil,
	}, {
		"different units, overlapping port ranges",
		MustPortRange("mysql/0", 80, 100, "TCP"),
		MustPortRange("wordpress/0", 90, 280, "TCP"),
		"port ranges .* conflict",
	}}

	for i, t := range testCases {
		c.Logf("test %d: %s", i, t.about)
		if t.expected == nil {
			c.Check(t.first.CheckConflicts(t.second), gc.IsNil)
			c.Check(t.second.CheckConflicts(t.first), gc.IsNil)
		} else if _, isString := t.expected.(string); isString {
			c.Check(t.first.CheckConflicts(t.second), gc.ErrorMatches, t.expected.(string))
			c.Check(t.second.CheckConflicts(t.first), gc.ErrorMatches, t.expected.(string))
		}
		// change test case protocols and test again
		c.Logf("test %d: %s (after protocol swap)", i, t.about)
		t.first.Protocol = swapProtocol(t.first.Protocol)
		t.second.Protocol = swapProtocol(t.second.Protocol)
		c.Logf("%+v %+v %v", t.first, t.second, t.expected)
		if t.expected == nil {
			c.Check(t.first.CheckConflicts(t.second), gc.IsNil)
			c.Check(t.second.CheckConflicts(t.first), gc.IsNil)
		} else if _, isString := t.expected.(string); isString {
			c.Check(t.first.CheckConflicts(t.second), gc.ErrorMatches, t.expected.(string))
			c.Check(t.second.CheckConflicts(t.first), gc.ErrorMatches, t.expected.(string))
		}

	}
}

func swapProtocol(protocol string) string {
	if strings.ToLower(protocol) == "tcp" {
		return "udp"
	}
	if strings.ToLower(protocol) == "udp" {
		return "tcp"
	}
	return protocol
}

func (p *PortRangeSuite) TestPortRangeString(c *gc.C) {
	c.Assert(state.PortRange{"wordpress/0", 80, 80, "TCP"}.String(),
		gc.Equals,
		"80-80/tcp")
	c.Assert(state.PortRange{"wordpress/0", 80, 100, "TCP"}.String(),
		gc.Equals,
		"80-100/tcp")
}

func (p *PortRangeSuite) TestPortRangeValidity(c *gc.C) {
	testCases := []struct {
		about    string
		ports    state.PortRange
		expected string
	}{{
		"single valid port",
		state.PortRange{"wordpress/0", 80, 80, "tcp"},
		"",
	}, {
		"valid port range",
		state.PortRange{"wordpress/0", 80, 90, "tcp"},
		"",
	}, {
		"valid udp port range",
		state.PortRange{"wordpress/0", 80, 90, "UDP"},
		"",
	}, {
		"invalid port range boundaries",
		state.PortRange{"wordpress/0", 90, 80, "tcp"},
		"invalid port range.*",
	}, {
		"invalid protocol",
		state.PortRange{"wordpress/0", 80, 80, "some protocol"},
		"invalid protocol.*",
	}, {
		"invalid unit",
		state.PortRange{"invalid unit", 80, 80, "tcp"},
		"invalid unit.*",
	}}

	for i, t := range testCases {
		c.Logf("test %d: %s", i, t.about)
		if t.expected == "" {
			c.Check(t.ports.Validate(), gc.IsNil)
		} else {
			c.Check(t.ports.Validate(), gc.ErrorMatches, t.expected)
		}
	}
}