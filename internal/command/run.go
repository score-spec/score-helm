/*
Apache Score
Copyright 2020 The Apache Software Foundation

This product includes software developed at
The Apache Software Foundation (http://www.apache.org/).
*/
package command

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"
	"strings"

	"github.com/imdario/mergo"
	"github.com/mitchellh/mapstructure"
	"github.com/spf13/cobra"
	"github.com/tidwall/sjson"
	"gopkg.in/yaml.v3"

	loader "github.com/score-spec/score-go/loader"
	schema "github.com/score-spec/score-go/schema"
	score "github.com/score-spec/score-go/types"

	helm "github.com/score-spec/score-helm/internal/helm"
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

	overrideParams []string

	skipValidation bool
	verbose        bool
)

func init() {
	runCmd.Flags().StringVarP(&scoreFile, "file", "f", scoreFileDefault, "Source SCORE file")
	runCmd.Flags().StringVar(&overridesFile, "overrides", overridesFileDefault, "Overrides file")
	runCmd.Flags().StringVarP(&valuesFile, "values", "", "", "Imported values file (in YAML format)")
	runCmd.Flags().StringVarP(&outFile, "output", "o", "", "Output file")

	runCmd.Flags().StringArrayVarP(&overrideParams, "property", "p", nil, "Overrides selected property value")

	runCmd.Flags().BoolVar(&skipValidation, "skip-validation", false, "DEPRECATED: Disables Score file schema validation")
	runCmd.Flags().BoolVar(&verbose, "verbose", false, "Enable diagnostic messages (written to STDERR)")

	rootCmd.AddCommand(runCmd)
}

var runCmd = &cobra.Command{
	Use:   "run",
	Short: "Translate SCORE file into Helm values file",
	RunE:  run,
	// we print errors ourselves at the top level
	SilenceErrors: true,
}

func run(cmd *cobra.Command, args []string) error {
	// don't print usage if we've parsed the args successfully
	cmd.SilenceUsage = true

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

	// Apply overrides from file (optional)
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

	// Apply overrides from command line (optional)
	//
	for _, pstr := range overrideParams {
		log.Print("Applying SCORE properties overrides...\n")

		jsonBytes, err := json.Marshal(srcMap)
		if err != nil {
			return fmt.Errorf("marshalling score spec: %w", err)
		}

		pmap := strings.SplitN(pstr, "=", 2)
		if len(pmap) <= 1 {
			var path = pmap[0]
			log.Printf("removing '%s'", path)
			if jsonBytes, err = sjson.DeleteBytes(jsonBytes, path); err != nil {
				return fmt.Errorf("removing '%s': %w", path, err)
			}
		} else {
			var path = pmap[0]
			var val interface{}
			if err := yaml.Unmarshal([]byte(pmap[1]), &val); err != nil {
				val = pmap[1]
			}

			log.Printf("overriding '%s' = '%s'", path, val)
			if jsonBytes, err = sjson.SetBytes(jsonBytes, path, val); err != nil {
				return fmt.Errorf("overriding '%s': %w", path, err)
			}
		}

		if err = json.Unmarshal(jsonBytes, &srcMap); err != nil {
			return fmt.Errorf("unmarshalling score spec: %w", err)
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

	// Apply upgrades to fix backports or backward incompatible things
	if changes, err := schema.ApplyCommonUpgradeTransforms(srcMap); err != nil {
		return fmt.Errorf("failed to upgrade spec: %w", err)
	} else if len(changes) > 0 {
		for _, change := range changes {
			log.Printf("Applying upgrade to specification: %s\n", change)
		}
	}
	
	// Validate SCORE spec
	//
	if !skipValidation {
		log.Print("Validating SCORE spec...\n")
		if err := schema.Validate(srcMap); err != nil {
			return fmt.Errorf("validating workload spec: %w", err)
		}
	}

	// Convert SCORE spec
	//
	var spec score.Workload
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
