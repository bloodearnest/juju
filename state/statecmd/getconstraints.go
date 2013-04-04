// Code shared by the CLI and API for the SetConstraints function.

package statecmd

import (
	"launchpad.net/juju-core/constraints"
	"launchpad.net/juju-core/state"
	"launchpad.net/juju-core/state/api/params"
)

// GetServiceContstraints returns the constraints for a given service
func GetServiceConstraints(st *state.State, args params.GetServiceConstraints) (params.GetServiceConstraintsResults, error) {
	svc, err := st.Service(args.ServiceName)
	if err != nil {
		return params.GetServiceConstraintsResults{constraints.Value{}}, err
	}
	constraints, err := svc.Constraints()
	return params.GetServiceConstraintsResults{constraints}, err
}
