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
	{{ $variableName }}:
	  value: '{{ $variableValue }}'
	{{- end }}
	{{- end }}
    image:
      name: {{ $container.Image }}
{{- end }}
`