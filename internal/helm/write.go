package helm

import (
	"io"

	"gopkg.in/yaml.v3"
)

// WriteYAML exports Helm values into YAML.
func WriteYAML(w io.Writer, values map[string]interface{}) error {
	enc := yaml.NewEncoder(w)
	enc.SetIndent(2)
	return enc.Encode(values)
}
