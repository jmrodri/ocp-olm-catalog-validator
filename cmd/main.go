package main

import (
	"errors"
	"fmt"
	"os"

	log "github.com/sirupsen/logrus"
	flag "github.com/spf13/pflag"

	"github.com/camilamacedo86/ocp-olm-catalog-validator/pkg/result"
	"github.com/camilamacedo86/ocp-olm-catalog-validator/pkg/validation"
	apimanifests "github.com/operator-framework/api/pkg/manifests"
	apierrors "github.com/operator-framework/api/pkg/validation/errors"
)

func main() {

	var optionalValues map[string]string
	var outputFormat string

	optionalValueEmpty := map[string]string{}
	flag.StringToStringVarP(&optionalValues, "optional-values", "", optionalValueEmpty,
		"Inform a []string map of key=values which can be used by the validator. e.g. to check the operator bundle "+
			"against an Kubernetes version that it is intended to be distributed use `--optional-values=k8s-version=1.22`")
	flag.StringVarP(&outputFormat, "output", "o", result.Text,
		"Result format for results. One of: [text, json-alpha1]. Note: output format types containing "+
			"\"alphaX\" are subject to change and not covered by guarantees of stable APIs.")

	flag.Parse()

	validate(outputFormat)
	results := runValidator(optionalValues)
	printResults(results, outputFormat)
}

func printResults(results []apierrors.ManifestResult, outputFormat string) {
	// Create Result to be output.
	res := result.NewResult()
	res.AddManifestResults(results...)

	if err := res.PrintWithFormat(outputFormat); err != nil {
		log.Fatal(err)
	}
}

func runValidator(optionalValues map[string]string) []apierrors.ManifestResult {
	// Read the bundle
	bundle, err := apimanifests.GetBundleFromDir(os.Args[1])
	if err != nil {
		log.Fatal(err)
	}

	objs := bundle.ObjectsToValidate()
	for _, obj := range bundle.Objects {
		objs = append(objs, obj)
	}

	// Pass the --optional-values. e.g. --optional-values="k8s-version=1.22"
	// or --optional-values="image-path=bundle.Dockerfile"
	objs = append(objs, optionalValues)

	// pass the objects to the validator
	results := validation.OpenShiftValidator.Validate(objs...)
	return results
}

func validate(outputFormat string)  {
	if len(os.Args) < 2 {
		log.Fatal(errors.New("an image tag or directory is a required argument"))
	}
	if outputFormat != result.JSONAlpha1 && outputFormat != result.Text {
		log.Fatal(fmt.Errorf("invalid value for output flag: %v", outputFormat))
	}
}


// createLogger creates a new logrus Entry that is optionally verbose.
func createLogger(verbose bool) *log.Entry {
	logger := log.NewEntry(result.NewLoggerTo(os.Stderr))
	if verbose {
		logger.Logger.SetLevel(log.DebugLevel)
	}
	return logger
}