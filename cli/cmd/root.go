package cmd

import (
	"fmt"

	"github.com/score-spec/score-helm/internal/version"
	"github.com/spf13/cobra"
)

var (
	rootCmd = &cobra.Command{
		Use:   "score-helm",
		Short: "SCORE to Helm translator",
		Long: `SCORE is a specification for defining environment agnostic configuration for cloud based workloads.
This tool produces a Helm chart from the SCORE specification.
Complete documentation is available at https://score.sh.`,
		Version: fmt.Sprintf("%s (build: %s; sha: %s)", version.Version, version.BuildTime, version.GitSHA),
	}
)

func init() {
	rootCmd.SetVersionTemplate(`{{with .Name}}{{printf "%s " .}}{{end}}{{printf "%s" .Version}}
`)
}

func Execute() error {
	return rootCmd.Execute()
}
