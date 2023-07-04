/*
Apache Score
Copyright 2020 The Apache Software Foundation

This product includes software developed at
The Apache Software Foundation (http://www.apache.org/).
*/
package helm

import (
	"testing"

	score "github.com/score-spec/score-go/types"
	"github.com/stretchr/testify/assert"
)

func TestScoreConvert(t *testing.T) {
	var tests = []struct {
		Name     string
		Spec     *score.WorkloadSpec
		Values   map[string]interface{}
		Expected map[string]interface{}
		Error    error
	}{
		// Success path
		//
		{
			Name: "Should convert SCORE to Helm values",
			Spec: &score.WorkloadSpec{
				Metadata: score.WorkloadMeta{
					Name: "test",
				},
				Service: score.ServiceSpec{
					Ports: score.ServicePortsSpecs{
						"www": score.ServicePortSpec{
							Port:       80,
							TargetPort: 8080,
						},
						"admin": score.ServicePortSpec{
							Port:     8080,
							Protocol: "UDP",
						},
					},
				},
				Containers: score.ContainersSpecs{
					"backend": score.ContainerSpec{
						Image: "busybox",
						Command: []string{
							"/bin/sh",
						},
						Args: []string{
							"-c",
							"while true; do printenv; echo ...sleeping 10 sec...; sleep 10; done",
						},
						Variables: map[string]string{
							"CONNECTION_STRING": "test connection string",
						},
					},
				},
			},
			Values: nil,
			Expected: map[string]interface{}{
				"service": map[string]interface{}{
					"type": "ClusterIP",
					"ports": []interface{}{
						map[string]interface{}{
							"name":     "admin",
							"port":     8080,
							"protocol": "UDP",
						},
						map[string]interface{}{
							"name":       "www",
							"port":       80,
							"targetPort": 8080,
						},
					},
				},
				"containers": map[string]interface{}{
					"backend": map[string]interface{}{
						"image": map[string]interface{}{
							"name": "busybox",
						},
						"command": []string{"/bin/sh"},
						"args": []string{
							"-c",
							"while true; do printenv; echo ...sleeping 10 sec...; sleep 10; done",
						},
						"env": []interface{}{
							map[string]interface{}{
								"name":  "CONNECTION_STRING",
								"value": "test connection string",
							},
						},
					},
				},
			},
		},
		{
			Name: "Should convert all resources references",
			Spec: &score.WorkloadSpec{
				Metadata: score.WorkloadMeta{
					Name: "test",
				},
				Containers: score.ContainersSpecs{
					"backend": score.ContainerSpec{
						Image: "busybox",
						Variables: map[string]string{
							"DEBUG":             "${resources.env.DEBUG}",
							"LOGS_LEVEL":        "$${LOGS_LEVEL}",
							"DOMAIN_NAME":       "${resources.dns.domain_name}",
							"CONNECTION_STRING": "postgresql://${resources.app-db.host}:${resources.app-db.port}/${resources.app-db.name}",
						},
						Volumes: []score.VolumeMountSpec{
							{
								Source:   "${resources.data}",
								Path:     "sub/path",
								Target:   "/mnt/data",
								ReadOnly: true,
							},
						},
					},
				},
				Resources: map[string]score.ResourceSpec{
					"env": {
						Type: "environment",
					},
					"app-db": {
						Type: "postgres",
					},
					"dns": {
						Type: "dns",
					},
					"data": {
						Type: "volume",
					},
				},
			},
			Values: map[string]interface{}{
				"app-db": map[string]interface{}{
					"host":      ".",
					"port":      5432,
					"name":      "test-db",
					"user.name": "<secret>",
					"password":  "<secret>",
				},
				"dns": map[string]interface{}{
					"domain": "test.domain.name",
				},
			},
			Expected: map[string]interface{}{
				"containers": map[string]interface{}{
					"backend": map[string]interface{}{
						"image": map[string]interface{}{
							"name": "busybox",
						},
						"env": []interface{}{
							map[string]interface{}{
								"name":  "CONNECTION_STRING",
								"value": "postgresql://.:5432/test-db",
							},
							map[string]interface{}{
								"name":  "DEBUG",
								"value": "", // fallback to default value
							},
							map[string]interface{}{
								"name":  "DOMAIN_NAME",
								"value": "", // referenced property does not exist
							},
							map[string]interface{}{
								"name":  "LOGS_LEVEL",
								"value": "${LOGS_LEVEL}", // do not expand escaped sequences, e.g. "$${..}"
							},
						},
						"volumeMounts": []interface{}{
							map[string]interface{}{
								"name":      "data", // expands to the resource name
								"subPath":   "sub/path",
								"mountPath": "/mnt/data",
								"readOnly":  true,
							},
						},
					},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.Name, func(t *testing.T) {
			var res = make(map[string]interface{})
			err := ConvertSpec(res, tt.Spec, tt.Values)

			if tt.Error != nil {
				// On Error
				//
				assert.ErrorContains(t, err, tt.Error.Error())
			} else {
				// On Success
				//
				assert.NoError(t, err)
				assert.Equal(t, tt.Expected, res)
			}
		})
	}
}
