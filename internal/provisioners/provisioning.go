// Copyright 2026 The Score Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package provisioners

import (
	"fmt"
	"maps"

	"github.com/score-spec/score-go/framework"

	"github.com/score-spec/score-helm/internal/state"
)

func ProvisionResources(currentState *state.State) (*state.State, error) {
	out := currentState

	// provision in sorted order
	orderedResources, err := currentState.GetSortedResourceUids()
	if err != nil {
		return nil, fmt.Errorf("failed to determine sort order for provisioning: %w", err)
	}

	out.Resources = maps.Clone(out.Resources)
	for _, resUid := range orderedResources {
		resState := out.Resources[resUid]

		var params map[string]interface{}
		if len(resState.Params) > 0 {
			resOutputs, err := out.GetResourceOutputForWorkload(resState.SourceWorkload)
			if err != nil {
				return nil, fmt.Errorf("%s: failed to find resource params for resource: %w", resUid, err)
			}
			sf := framework.BuildSubstitutionFunction(out.Workloads[resState.SourceWorkload].Spec.Metadata, resOutputs)
			rawParams, err := framework.Substitute(resState.Params, sf)
			if err != nil {
				return nil, fmt.Errorf("%s: failed to substitute params for resource: %w", resUid, err)
			}
			params = rawParams.(map[string]interface{})
		}
		resState.Params = params

		// ==========================================================================================
		// TODO: HERE IS WHERE YOU WOULD USE THE RESOURCE TYPE, CLASS, ID, AND PARAMS TO PROVISION IT
		// ==========================================================================================

		resState.Outputs = map[string]interface{}{}
		out.Resources[resUid] = resState
	}

	return out, nil
}
