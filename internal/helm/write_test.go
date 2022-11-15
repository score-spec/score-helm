/*
Apache Score
Copyright 2020 The Apache Software Foundation

This product includes software developed at
The Apache Software Foundation (http://www.apache.org/).
*/
package helm

import (
	"bufio"
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestYamlEncode(t *testing.T) {
	var tests = []struct {
		Name   string
		Source map[string]interface{}
		Output []byte
		Error  error
	}{
		{
			Name: "Should encode the values",
			Source: map[string]interface{}{
				"service": map[string]interface{}{
					"type": "ClusterIP",
					"ports": []interface{}{
						map[string]interface{}{
							"name":       "www",
							"protocol":   "TCP",
							"port":       80,
							"targetPort": 8080,
						},
					},
				},
				"containers": map[string]interface{}{
					"my-container": map[string]interface{}{
						"image": map[string]interface{}{
							"name": "busybox:latest",
						},
						"command": []string{"/bin/echo"},
						"args": []string{
							"-c",
							"Hello $(FRIEND)",
						},
						"env": []interface{}{
							map[string]interface{}{
								"name":  "FRIEND",
								"value": "World!",
							},
						},
						"volumeMounts": []interface{}{
							map[string]interface{}{
								"name":      "${resources.data}",
								"subPath":   "sub/path",
								"mountPath": "/mnt/data",
								"readOnly":  true,
							},
						},
						"livenessProbe": map[string]interface{}{
							"httpGet": map[string]interface{}{
								"path": "/health",
								"port": "http",
							},
						},
						"readinessProbe": map[string]interface{}{
							"httpGet": map[string]interface{}{
								"path": "/ready",
								"port": "http",
								"httpHeaders": []interface{}{
									map[string]interface{}{
										"name":  "Custom-Header",
										"value": "Awesome",
									},
								},
							},
						},
						"resources": map[string]interface{}{
							"limits": map[string]interface{}{
								"cpu":    "100m",
								"memory": "128Mi",
							},
							"requests": map[string]interface{}{
								"cpu":    "100m",
								"memory": "128Mi",
							},
						},
					},
				},
			},
			Output: []byte(`containers:
  my-container:
    args:
      - -c
      - Hello $(FRIEND)
    command:
      - /bin/echo
    env:
      - name: FRIEND
        value: World!
    image:
      name: busybox:latest
    livenessProbe:
      httpGet:
        path: /health
        port: http
    readinessProbe:
      httpGet:
        httpHeaders:
          - name: Custom-Header
            value: Awesome
        path: /ready
        port: http
    resources:
      limits:
        cpu: 100m
        memory: 128Mi
      requests:
        cpu: 100m
        memory: 128Mi
    volumeMounts:
      - mountPath: /mnt/data
        name: ${resources.data}
        readOnly: true
        subPath: sub/path
service:
  ports:
    - name: www
      port: 80
      protocol: TCP
      targetPort: 8080
  type: ClusterIP
`),
		},
	}

	for _, tt := range tests {
		t.Run(tt.Name, func(t *testing.T) {
			buf := bytes.Buffer{}
			w := bufio.NewWriter(&buf)

			err := WriteYAML(w, tt.Source)
			w.Flush()

			if tt.Error != nil {
				// On Error
				//
				assert.ErrorContains(t, err, tt.Error.Error())
			} else {
				// On Success
				//
				assert.NoError(t, err)
				assert.Equal(t, tt.Output, buf.Bytes())
			}
		})
	}
}
