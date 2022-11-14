/*
Apache Score
Copyright 2020 The Apache Software Foundation

This product includes software developed at
The Apache Software Foundation (http://www.apache.org/).
*/
package helm

import (
	"fmt"
	"os"
	"regexp"
	"sort"
	"strings"

	score "github.com/score-spec/score-go/types"
)

// resourceRefRegex extracts the resource ID from the resource reference: '${resources.RESOURCE_ID}'
var resourceRefRegex = regexp.MustCompile(`\${resources\.(.+)}`)

// resourcesMap is an internal utility type to group some helper methods.
type resourcesMap struct {
	Spec   map[string]score.ResourceSpec
	Values map[string]interface{}
}

// mapResourceVar maps resources properties references.
// Returns an empty string if the reference can't be resolved.
func (r resourcesMap) mapVar(ref string) string {
	if ref == "$" {
		return ref
	}

	var segments = strings.SplitN(ref, ".", 3)
	if segments[0] != "resources" || len(segments) != 3 {
		return ""
	}

	var resName = segments[1]
	var propName = segments[2]
	if res, ok := r.Spec[resName]; ok {
		if prop, ok := res.Properties[propName]; ok {

			// Look-up the value for the property
			if src, ok := r.Values[resName]; ok {
				if srcMap, ok := src.(map[string]interface{}); ok {
					if val, ok := srcMap[propName]; ok {
						return fmt.Sprintf("%v", val)
					}
				}
			}

			// Use the default value provided (if any)
			return fmt.Sprintf("%v", prop.Default)
		}
	}

	return ""
}

// getProbeDetails extracts an httpGet probe details from the source spec.
// Returns nil if the source spec is empty.
func getProbeDetails(probe *score.ContainerProbeSpec) map[string]interface{} {
	if probe.HTTPGet.Path == "" {
		return nil
	}

	var res = map[string]interface{}{
		"type": "http",
		"path": probe.HTTPGet.Path,
		"port": probe.HTTPGet.Port,
	}

	if len(probe.HTTPGet.HTTPHeaders) > 0 {
		var hdrs = map[string]string{}
		for _, hdr := range probe.HTTPGet.HTTPHeaders {
			hdrs[hdr.Name] = hdr.Value
		}
		res["httpHeaders"] = hdrs
	}

	return res
}

// ConvertSpec converts SCORE specification into Helm values map.
func ConvertSpec(dest map[string]interface{}, spec *score.WorkloadSpec, values map[string]interface{}) error {
	if values == nil {
		values = make(map[string]interface{})
	}
	var resourcesSpec = resourcesMap{
		Spec:   spec.Resources,
		Values: values,
	}

	if len(spec.Service.Ports) > 0 {
		var ports = make([]interface{}, 0, len(spec.Service.Ports))
		for name, port := range spec.Service.Ports {
			var pVals = map[string]interface{}{
				"name": name,
				"port": port.Port,
			}
			if port.Protocol != "" {
				pVals["protocol"] = port.Protocol
			}
			if port.TargetPort > 0 {
				pVals["targetPort"] = port.TargetPort
			}
			ports = append(ports, pVals)
		}

		// NOTE: Sorting is necessary for DeepEqual call within our Unit Tests to work reliably
		sort.Slice(ports, func(i, j int) bool {
			return ports[i].(map[string]interface{})["name"].(string) < ports[j].(map[string]interface{})["name"].(string)
		})
		// END (NOTE)
		dest["service"] = map[string]interface{}{
			"type":  "ClusterIP",
			"ports": ports,
		}
	}

	var containers = map[string]interface{}{}
	for name, cSpec := range spec.Containers {
		var cVals = map[string]interface{}{
			"image": map[string]interface{}{
				"name": cSpec.Image,
			},
		}

		if len(cSpec.Command) > 0 {
			cVals["command"] = cSpec.Command
		}
		if len(cSpec.Args) > 0 {
			cVals["args"] = cSpec.Args
		}
		if len(cSpec.Variables) > 0 {
			var env = make([]interface{}, 0, len(cSpec.Variables))
			for key, val := range cSpec.Variables {
				val = os.Expand(val, resourcesSpec.mapVar)
				env = append(env, map[string]interface{}{"name": key, "value": val})
			}

			// NOTE: Sorting is necessary for DeepEqual call within our Unit Tests to work reliably
			sort.Slice(env, func(i, j int) bool {
				return env[i].(map[string]interface{})["name"].(string) < env[j].(map[string]interface{})["name"].(string)
			})
			// END (NOTE)
			cVals["env"] = env
		}

		if len(cSpec.Volumes) > 0 {
			var volumes = make([]interface{}, 0, len(cSpec.Volumes))
			for _, vol := range cSpec.Volumes {
				var source = resourceRefRegex.ReplaceAllString(vol.Source, "$1")
				var vVals = map[string]interface{}{
					"name":      source,
					"subPath":   vol.Path,
					"mountPath": vol.Target,
					"readOnly":  vol.ReadOnly,
				}
				volumes = append(volumes, vVals)
			}
			cVals["volumeMounts"] = volumes
		}

		if probe := getProbeDetails(&cSpec.LivenessProbe); len(probe) > 0 {
			cVals["livenessProbe"] = probe
		}
		if probe := getProbeDetails(&cSpec.ReadinessProbe); len(probe) > 0 {
			cVals["readinessProbe"] = probe
		}

		if len(cSpec.Resources.Requests) > 0 || len(cSpec.Resources.Limits) > 0 {
			cVals["resources"] = map[string]interface{}{
				"requests": cSpec.Resources.Requests,
				"limits":   cSpec.Resources.Limits,
			}
		}

		containers[name] = cVals
	}
	dest["containers"] = containers

	return nil
}
