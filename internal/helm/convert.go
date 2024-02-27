/*
Apache Score
Copyright 2020 The Apache Software Foundation

This product includes software developed at
The Apache Software Foundation (http://www.apache.org/).
*/
package helm

import (
	"fmt"
	"sort"

	score "github.com/score-spec/score-go/types"
)

// getProbeDetails extracts an httpGet probe details from the source spec.
// Returns nil if the source spec is empty.
func getProbeDetails(probe *score.ContainerProbe) map[string]interface{} {
	if probe.HttpGet.Path == "" {
		return nil
	}

	var res = map[string]interface{}{
		"type": "http",
		"path": probe.HttpGet.Path,
		"port": probe.HttpGet.Port,
	}

	if len(probe.HttpGet.HttpHeaders) > 0 {
		var hdrs = map[string]string{}
		for _, hdr := range probe.HttpGet.HttpHeaders {
			if hdr.Name != nil && hdr.Value != nil {
				hdrs[*hdr.Name] = *hdr.Value
			}
		}
		res["httpHeaders"] = hdrs
	}

	return res
}

// ConvertSpec converts SCORE specification into Helm values map.
func ConvertSpec(dest map[string]interface{}, spec *score.Workload, values map[string]interface{}) error {
	if values == nil {
		values = make(map[string]interface{})
	}
	context, err := buildContext(spec.Metadata, spec.Resources, values)
	if err != nil {
		return fmt.Errorf("preparing context: %w", err)
	}

	if spec.Service != nil && len(spec.Service.Ports) > 0 {
		var ports = make([]interface{}, 0, len(spec.Service.Ports))
		for name, port := range spec.Service.Ports {
			var pVals = map[string]interface{}{
				"name": name,
				"port": port.Port,
			}
			if port.Protocol != nil {
				pVals["protocol"] = string(*port.Protocol)
			}
			if port.TargetPort != nil {
				pVals["targetPort"] = *port.TargetPort
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
				val = context.Substitute(val)
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
				var source = context.Substitute(vol.Source)
				var vVals = map[string]interface{}{
					"name":      source,
					"mountPath": vol.Target,
				}
				if vol.Path != nil {
					vVals["subPath"] = *vol.Path
				}
				if vol.ReadOnly != nil {
					vVals["readOnly"] = *vol.ReadOnly
				}
				volumes = append(volumes, vVals)
			}
			cVals["volumeMounts"] = volumes
		}

		if cSpec.LivenessProbe != nil {
			if probe := getProbeDetails(cSpec.LivenessProbe); len(probe) > 0 {
				cVals["livenessProbe"] = probe
			}
		}
		if cSpec.ReadinessProbe != nil {
			if probe := getProbeDetails(cSpec.ReadinessProbe); len(probe) > 0 {
				cVals["readinessProbe"] = probe
			}
		}

		if cSpec.Resources != nil {
			containerResources := make(map[string]interface{})
			if out := getContainerResources(cSpec.Resources.Limits); len(out) > 0 {
				containerResources["limits"] = out
			}
			if out := getContainerResources(cSpec.Resources.Requests); len(out) > 0 {
				containerResources["requests"] = out
			}
			if len(containerResources) > 0 {
				cVals["resources"] = containerResources
			}
		}
		containers[name] = cVals
	}
	dest["containers"] = containers

	return nil
}

func getContainerResources(requests *score.ResourcesLimits) map[string]interface{} {
	out := make(map[string]interface{})
	if requests.Cpu != nil {
		out["cpu"] = *requests.Cpu
	}
	if requests.Memory != nil {
		out["memory"] = *requests.Memory
	}
	return out
}
