// Copyright 2024 The Score Authors
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

package command

import (
	"context"
	"os"
	"strings"
	"testing"

	"github.com/score-spec/score-go/framework"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/score-spec/score-helm/internal/state"
)

func TestInitNominal(t *testing.T) {
	td := t.TempDir()

	wd, _ := os.Getwd()
	require.NoError(t, os.Chdir(td))
	defer func() {
		require.NoError(t, os.Chdir(wd))
	}()

	stdout, stderr, err := executeAndResetCommand(context.Background(), rootCmd, []string{"init"})
	assert.NoError(t, err)
	assert.Equal(t, "", stdout)
	assert.NotEqual(t, "", strings.TrimSpace(stderr))

	raw, err := os.ReadFile("score.yaml")
	require.NoError(t, err)
	assert.Equal(t, `apiVersion: score.dev/v1b1
containers:
  main:
    image: stefanprodan/podinfo
metadata:
  name: example
service:
  ports:
    web:
      port: 8080
`, string(raw))

	stdout, stderr, err = executeAndResetCommand(context.Background(), rootCmd, []string{"generate", "score.yaml"})
	assert.NoError(t, err)
	assert.Equal(t, ``, stdout)
	assert.NotEqual(t, "", strings.TrimSpace(stderr))

	sd, ok, err := state.LoadStateDirectory(".")
	assert.NoError(t, err)
	if assert.True(t, ok) {
		assert.Equal(t, state.DefaultRelativeStateDirectory, sd.Path)
		assert.Len(t, sd.State.Workloads, 1)
		assert.Equal(t, map[framework.ResourceUid]framework.ScoreResourceState[state.ResourceExtras]{}, sd.State.Resources)
		assert.Equal(t, map[string]interface{}{}, sd.State.SharedState)
	}
}

func TestInitNominal_run_twice(t *testing.T) {
	td := t.TempDir()

	wd, _ := os.Getwd()
	require.NoError(t, os.Chdir(td))
	defer func() {
		require.NoError(t, os.Chdir(wd))
	}()

	// first init
	stdout, stderr, err := executeAndResetCommand(context.Background(), rootCmd, []string{"init", "--file", "score2.yaml"})
	assert.NoError(t, err)
	assert.Equal(t, "", stdout)
	assert.NotEqual(t, "", strings.TrimSpace(stderr))

	// init again
	stdout, stderr, err = executeAndResetCommand(context.Background(), rootCmd, []string{"init"})
	assert.NoError(t, err)
	assert.Equal(t, "", stdout)
	assert.NotEqual(t, "", strings.TrimSpace(stderr))

	_, err = os.Stat("score.yaml")
	assert.NoError(t, err)
	_, err = os.Stat("score2.yaml")
	assert.NoError(t, err)

	sd, ok, err := state.LoadStateDirectory(".")
	assert.NoError(t, err)
	if assert.True(t, ok) {
		assert.Equal(t, state.DefaultRelativeStateDirectory, sd.Path)
		assert.Equal(t, map[string]framework.ScoreWorkloadState[state.WorkloadExtras]{}, sd.State.Workloads)
		assert.Equal(t, map[framework.ResourceUid]framework.ScoreResourceState[state.ResourceExtras]{}, sd.State.Resources)
		assert.Equal(t, map[string]interface{}{}, sd.State.SharedState)
	}
}
