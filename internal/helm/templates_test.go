/*
Apache Score
Copyright 2022 The Apache Software Foundation

This product includes software developed at
The Apache Software Foundation (http://www.apache.org/).
*/
package helm

import (
	"testing"

	score "github.com/score-spec/score-go/types"
	assert "github.com/stretchr/testify/assert"
)

func TestBuildContext(t *testing.T) {
	var meta = score.WorkloadMeta{
		Name: "test-name",
	}

	var resources = score.ResourcesSpecs{
		"env": score.ResourceSpec{
			Type: "environment",
			Properties: map[string]score.ResourcePropertySpec{
				"DEBUG": {Required: false, Default: true},
			},
		},
		"db": score.ResourceSpec{
			Type: "postgres",
			Properties: map[string]score.ResourcePropertySpec{
				"host": {Required: true, Default: "."},
				"port": {Required: true, Default: "5342"},
				"name": {Required: true},
			},
		},
		"dns": score.ResourceSpec{
			Type: "dns",
			Properties: map[string]score.ResourcePropertySpec{
				"domain": {},
			},
		},
	}

	var values = map[string]interface{}{
		"db": map[string]interface{}{
			"host": "localhost",
			"name": "test-db",
		},
		"dns": map[string]interface{}{
			"domain": "test.domain.name",
		},
	}

	context, err := buildContext(meta, resources, values)
	assert.NoError(t, err)

	assert.Equal(t, templatesContext{
		"metadata.name": "test-name",

		"resources.env":       "env",
		"resources.env.DEBUG": "true",

		"resources.db":      "db",
		"resources.db.host": "localhost",
		"resources.db.port": "5342",
		"resources.db.name": "test-db",

		"resources.dns":        "dns",
		"resources.dns.domain": "test.domain.name",
	}, context)
}

func TestMapVar(t *testing.T) {
	var context = templatesContext{
		"metadata.name": "test-name",

		"resources.env":       "env",
		"resources.env.DEBUG": "true",

		"resources.db":      "db",
		"resources.db.host": "localhost",
		"resources.db.port": "5342",
		"resources.db.name": "test-db",

		"resources.dns":        "shared.dns",
		"resources.dns.domain": "test.domain.name",
	}

	assert.Equal(t, "", context.mapVar(""))
	assert.Equal(t, "$", context.mapVar("$"))

	assert.Equal(t, "test-name", context.mapVar("metadata.name"))
	assert.Equal(t, "", context.mapVar("metadata.name.nil"))
	assert.Equal(t, "", context.mapVar("metadata.nil"))

	assert.Equal(t, "true", context.mapVar("resources.env.DEBUG"))

	assert.Equal(t, "db", context.mapVar("resources.db"))
	assert.Equal(t, "localhost", context.mapVar("resources.db.host"))
	assert.Equal(t, "5342", context.mapVar("resources.db.port"))
	assert.Equal(t, "test-db", context.mapVar("resources.db.name"))
	assert.Equal(t, "", context.mapVar("resources.db.name.nil"))
	assert.Equal(t, "", context.mapVar("resources.db.nil"))
	assert.Equal(t, "", context.mapVar("resources.nil"))
	assert.Equal(t, "", context.mapVar("nil.db.name"))
}

func TestSubstitute(t *testing.T) {
	var context = templatesContext{
		"metadata.name": "test-name",

		"resources.env":       "env",
		"resources.env.DEBUG": "true",

		"resources.db":      "db",
		"resources.db.host": "localhost",
		"resources.db.port": "5342",
		"resources.db.name": "test-db",

		"resources.dns":        "dns",
		"resources.dns.domain": "test.domain.name",
	}

	assert.Equal(t, "", context.Substitute(""))
	assert.Equal(t, "abc", context.Substitute("abc"))
	assert.Equal(t, "abc $ abc", context.Substitute("abc $$ abc"))
	assert.Equal(t, "${abc}", context.Substitute("$${abc}"))

	assert.Equal(t, "The name is 'test-name'", context.Substitute("The name is '${metadata.name}'"))
	assert.Equal(t, "The name is ''", context.Substitute("The name is '${metadata.nil}'"))

	assert.Equal(t, "resources.env.DEBUG", context.Substitute("resources.env.DEBUG"))

	assert.Equal(t, "db", context.Substitute("${resources.db}"))
	assert.Equal(t,
		"postgresql://:@localhost:5342/test-db",
		context.Substitute("postgresql://${resources.db.user}:${resources.db.password}@${resources.db.host}:${resources.db.port}/${resources.db.name}"))
}
