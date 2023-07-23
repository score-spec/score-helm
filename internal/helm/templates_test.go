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

func TestMapVar(t *testing.T) {
	var meta = score.WorkloadMeta{
		Name: "test-name",
	}

	var resources = score.ResourcesSpecs{
		"env": score.ResourceSpec{
			Type: "environment",
		},
		"db": score.ResourceSpec{
			Type: "postgres",
		},
		"dns": score.ResourceSpec{
			Type: "dns",
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

	assert.Equal(t, "", context.mapVar(""))
	assert.Equal(t, "$", context.mapVar("$"))

	assert.Equal(t, "test-name", context.mapVar("metadata.name"))
	assert.Equal(t, "", context.mapVar("metadata.name.nil"))
	assert.Equal(t, "", context.mapVar("metadata.nil"))

	assert.Equal(t, "", context.mapVar("resources.env.DEBUG"))

	assert.Equal(t, "db", context.mapVar("resources.db"))
	assert.Equal(t, "localhost", context.mapVar("resources.db.host"))
	assert.Equal(t, "", context.mapVar("resources.db.port"))
	assert.Equal(t, "test-db", context.mapVar("resources.db.name"))
	assert.Equal(t, "", context.mapVar("resources.db.name.nil"))
	assert.Equal(t, "", context.mapVar("resources.db.nil"))
	assert.Equal(t, "", context.mapVar("resources.nil"))
	assert.Equal(t, "", context.mapVar("nil.db.name"))
}

func TestSubstitute(t *testing.T) {
	var meta = score.WorkloadMeta{
		Name: "test-name",
	}

	var resources = score.ResourcesSpecs{
		"env": score.ResourceSpec{
			Type: "environment",
		},
		"db": score.ResourceSpec{
			Type: "postgres",
		},
		"dns": score.ResourceSpec{
			Type: "dns",
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

	assert.Equal(t, "", context.Substitute(""))
	assert.Equal(t, "abc", context.Substitute("abc"))
	assert.Equal(t, "$abc", context.Substitute("$abc"))
	assert.Equal(t, "abc $ abc", context.Substitute("abc $$ abc"))
	assert.Equal(t, "${abc}", context.Substitute("$${abc}"))

	assert.Equal(t, "The name is 'test-name'", context.Substitute("The name is '${metadata.name}'"))
	assert.Equal(t, "The name is ''", context.Substitute("The name is '${metadata.nil}'"))

	assert.Equal(t, "resources.badref.DEBUG", context.Substitute("resources.badref.DEBUG"))

	assert.Equal(t, "db", context.Substitute("${resources.db}"))
	assert.Equal(t,
		"postgresql://:@localhost:/test-db",
		context.Substitute("postgresql://${resources.db.user}:${resources.db.password}@${resources.db.host}:${resources.db.port}/${resources.db.name}"))
}
