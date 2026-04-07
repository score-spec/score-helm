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

package convert

const defaultValuesTemplate = `{{ $workloadName := .WorkloadName }}{{ $service := .Spec.Service }}{{ $resources := .Spec.Resources }}containers:{{ range $containerName, $container := .Spec.Containers }}
  {{ $containerName }}:
    {{- if (gt (len $container.Args) 0) }}
    args:
      {{- range $i, $arg := $container.Args }}
      - {{ $arg }}
      {{- end }}
    {{- end }}
    {{- if (gt (len $container.Command) 0) }}
    command:
      {{- range $i, $cmd := $container.Command }}
      - {{ $cmd }}
      {{- end }}
    {{- end }}
	  {{- if (gt (len $container.Variables) 0) }}
    env:
    {{- range $variableName, $variableValue := $container.Variables }}
      - name: {{ $variableName }}
        value: "{{ $variableValue }}"
    {{- end }}
    {{- end }}
    image:
      name: {{ $container.Image }}
    {{- if (ne $container.LivenessProbe nil) }}
    livenessProbe:
      {{- if (ne $container.LivenessProbe.Exec nil) }}
      exec:
        command: {{ $container.LivenessProbe.Exec.Command }}
      {{- else if (ne $container.LivenessProbe.HttpGet nil) }}
      httpGet:
        port: {{ $container.LivenessProbe.HttpGet.Port }}
        {{- if (ne $container.LivenessProbe.HttpGet.Path "") }}
        path: '{{ $container.LivenessProbe.HttpGet.Path }}'
        {{- end }}
      {{- end }}
    {{- end }}
    {{- if (ne $container.ReadinessProbe nil) }}
    readinessProbe:
      {{- if (ne $container.ReadinessProbe.Exec nil) }}
      exec:
        command: {{ $container.ReadinessProbe.Exec.Command }}
      {{- else if (ne $container.ReadinessProbe.HttpGet nil) }}
      httpGet:
        port: {{ $container.ReadinessProbe.HttpGet.Port }}
        {{- if (ne $container.ReadinessProbe.HttpGet.Path "") }}
        path: '{{ $container.ReadinessProbe.HttpGet.Path }}'
        {{- end }}
      {{- end }}
    {{- end }}
    {{- if (ne $container.Resources nil) }}
    resources:
      {{- if (ne $container.Resources.Limits nil) }}
      limits:
        {{- if (ne $container.Resources.Limits.Cpu nil) }}
        cpu: {{ $container.Resources.Limits.Cpu }}
        {{- end }}
        {{- if (ne $container.Resources.Limits.Memory nil) }}
        memory: {{ $container.Resources.Limits.Memory }}
        {{- end }}
      {{- end }}
      {{- if (ne $container.Resources.Requests nil) }}
      requests:
        {{- if (ne $container.Resources.Requests.Cpu nil) }}
        cpu: {{ $container.Resources.Requests.Cpu }}
        {{- end }}
        {{- if (ne $container.Resources.Requests.Memory nil) }}
        memory: {{ $container.Resources.Requests.Memory }}
        {{- end }}
      {{- end }}
    {{- end }}
{{- end }}
{{- if and (ne $service nil) (gt (len $service.Ports) 0) }}
service:
  ports:
  {{- range $portName, $port := $service.Ports }}
    - name: {{ $portName }}
      port: {{ $port.Port }}
      {{- if ne $port.Protocol nil }}
      protocol: {{ $port.Protocol }}
      {{- end }}
      {{- if ne $port.TargetPort nil }}
      targetPort: {{ $port.TargetPort }}
      {{- end }}
  {{- end }}
{{- end }}
`