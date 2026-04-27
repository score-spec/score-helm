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
	"errors"
	"fmt"
	"log/slog"
	"os"

	"github.com/score-spec/score-go/framework"
	"github.com/spf13/cobra"

	"github.com/score-spec/score-helm/internal/state"
)

const (
	initCmdFileFlag         = "file"
	initCmdFileNoSampleFlag = "no-sample"

	DefaultScoreFileContent = `# Score provides a developer-centric and platform-agnostic
# Workload specification to improve developer productivity and experience.
# Score eliminates configuration management between local and remote environments.
#
# Get started with Score: https://docs.score.dev/docs/get-started/.
---
apiVersion: score.dev/v1b1
metadata:
  name: hello-world
  annotations:
    tags: "nodejs,http,website,javascript,postgres"
containers:
  hello-world:
    image: scorespec/sample-score-app:latest
    variables:
      PORT: "3000"
      MESSAGE: "Hello, World!"
resources:
  db:
    type: postgres
  dns:
    type: dns
  route:
    type: route
service:
  ports:
    www:
      port: 8080
      targetPort: 3000
`
)

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialise the local state directory and sample score file",
	Args:  cobra.NoArgs,
	CompletionOptions: cobra.CompletionOptions{
		HiddenDefaultCmd: true,
	},
	SilenceErrors: true,
	RunE: func(cmd *cobra.Command, args []string) error {
		cmd.SilenceUsage = true

		sd, ok, err := state.LoadStateDirectory(".")
		if err != nil {
			return fmt.Errorf("failed to load existing state directory: %w", err)
		} else if ok {
			slog.Info("Found existing state directory", "dir", sd.Path)
		} else {
			slog.Info("Writing new state directory", "dir", state.DefaultRelativeStateDirectory)
			sd = &state.StateDirectory{
				Path: state.DefaultRelativeStateDirectory,
				State: state.State{
					Workloads:   map[string]framework.ScoreWorkloadState[state.WorkloadExtras]{},
					Resources:   map[framework.ResourceUid]framework.ScoreResourceState[state.ResourceExtras]{},
					SharedState: map[string]interface{}{},
				},
			}
			slog.Info("Writing new state directory", "dir", sd.Path)
			if err := sd.Persist(); err != nil {
				return fmt.Errorf("failed to persist new state directory: %w", err)
			}
		}

		initCmdScoreFile, _ := cmd.Flags().GetString(initCmdFileFlag)
		if _, err := os.Stat(initCmdScoreFile); err != nil {
			if v, _ := cmd.Flags().GetBool(initCmdFileNoSampleFlag); v {
				slog.Info(fmt.Sprintf("Initial Score file '%s' does not exist - and sample generation is disabled", initCmdScoreFile))
			} else {
				if !errors.Is(err, os.ErrNotExist) {
					return fmt.Errorf("failed to check for existing Score file: %w", err)
				}
				if err := os.WriteFile(initCmdScoreFile, []byte(DefaultScoreFileContent), 0755); err != nil {
					return fmt.Errorf("failed to write Score file: %w", err)
				}
				slog.Info("Created initial Score file", "file", initCmdScoreFile)
			}
		} else {
			slog.Info("Skipping creation of initial Score file since it already exists", "file", initCmdScoreFile)
		}

		return nil
	},
}

func init() {
	initCmd.Flags().StringP(initCmdFileFlag, "f", "score.yaml", "The score file to initialize")
	initCmd.Flags().Bool(initCmdFileNoSampleFlag, false, "Disable generation of the sample score file")
	rootCmd.AddCommand(initCmd)
}
