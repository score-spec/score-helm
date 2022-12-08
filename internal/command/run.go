/*
Apache Score
Copyright 2020 The Apache Software Foundation

This product includes software developed at
The Apache Software Foundation (http://www.apache.org/).
*/
package command

import (
	"fmt"
	"io"
	"log"
	"os"

	"github.com/imdario/mergo"
	"github.com/mitchellh/mapstructure"
	loader "github.com/score-spec/score-go/loader"
	score "github.com/score-spec/score-go/types"
	helm "github.com/score-spec/score-helm/internal/helm"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

const (
	scoreFileDefault     = "./score.yaml"
	overridesFileDefault = "./overrides.score.yaml"
)

var (
	scoreFile     string
	overridesFile string
	outFile       string
	valuesFile    string

	verbose bool
)

func init() {
	runCmd.Flags().StringVarP(&scoreFile, "file", "f", scoreFileDefault, "Source SCORE file")
	runCmd.Flags().StringVar(&overridesFile, "overrides", overridesFileDefault, "Overrides file")
	runCmd.Flags().StringVarP(&valuesFile, "values", "", "", "Imported values file (in YAML format)")
	runCmd.Flags().StringVarP(&outFile, "output", "o", "", "Output file")

	runCmd.Flags().BoolVar(&verbose, "verbose", false, "Enable diagnostic messages (written to STDERR)")

	rootCmd.AddCommand(runCmd)
}

var runCmd = &cobra.Command{
	Use:   "run",
	Short: "Translate SCORE file into Helm values file",
	RunE:  run,
}

func run(cmd *cobra.Command, args []string) error {
	if !verbose {
		log.SetOutput(io.Discard)
	}

	// Open source file
	//
	var err error
	var src *os.File
	log.Printf("Reading '%s'...\n", scoreFile)
	if src, err = os.Open(scoreFile); err != nil {
		return err
	}
	defer src.Close()

	// Parse SCORE spec
	//
	log.Print("Parsing SCORE spec...\n")
	var srcMap map[string]interface{}
	if err := loader.ParseYAML(&srcMap, src); err != nil {
		return err
	}

	// Apply overrides (optional)
	//
	if overridesFile != "" {
		log.Printf("Checking '%s'...\n", overridesFile)
		if ovr, err := os.Open(overridesFile); err == nil {
			defer ovr.Close()

			log.Print("Applying SCORE overrides...\n")
			var ovrMap map[string]interface{}
			if err := loader.ParseYAML(&ovrMap, ovr); err != nil {
				return err
			}
			if err := mergo.MergeWithOverwrite(&srcMap, ovrMap); err != nil {
				return fmt.Errorf("applying overrides fom '%s': %w", overridesFile, err)
			}
		} else if !os.IsNotExist(err) || overridesFile != overridesFileDefault {
			return err
		}
	}

	// Load imported values (optional)
	//
	var values = make(map[string]interface{})
	if valuesFile != "" {
		log.Printf("Importing values from '%s'...\n", valuesFile)
		if srcFile, err := os.Open(valuesFile); err != nil {
			return err
		} else {
			defer srcFile.Close()

			if err := yaml.NewDecoder(srcFile).Decode(values); err != nil {
				return fmt.Errorf("parsing values file '%s': %w", valuesFile, err)
			}
		}
	}

	// Validate SCORE spec
	//
	var spec score.WorkloadSpec
	log.Print("Validating SCORE spec...\n")
	if err = mapstructure.Decode(srcMap, &spec); err != nil {
		return fmt.Errorf("validating workload spec: %w", err)
	}

	// Prepare Helm values
	//
	var vals = make(map[string]interface{})
	log.Print("Preparing Helm values...\n")
	if err = helm.ConvertSpec(vals, &spec, values); err != nil {
		return fmt.Errorf("preparing Helm values: %w", err)
	}

	// Open output file (optional)
	//
	var dest = io.Writer(os.Stdout)
	if outFile != "" {
		log.Printf("Creating '%s'...\n", outFile)
		destFile, err := os.Create(outFile)
		if err != nil {
			return err
		}
		defer destFile.Close()

		dest = io.MultiWriter(dest, destFile)
	}

	// Output Helm values
	//
	log.Print("Writing Helm values...\n")
	if err = helm.WriteYAML(dest, vals); err != nil {
		return err
	}

	return nil
}
