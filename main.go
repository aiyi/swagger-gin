package main

import (
	"os"

	"github.com/aiyi/swagger-gin/generator"
	"github.com/jessevdk/go-flags"
)

var opts struct {
	Spec           flags.Filename `long:"spec" short:"f" description:"the spec file to use" default:"./swagger.json"`
	APIPackage     string         `long:"api-package" short:"a" description:"the package to save the operations" default:"operations"`
	ModelPackage   string         `long:"model-package" short:"m" description:"the package to save the models" default:"models"`
	ServerPackage  string         `long:"server-package" short:"s" description:"the package to save the server specific code" default:"restapi"`
	Target         flags.Filename `long:"target" short:"t" default:"./" description:"the base directory for generating the files"`
	Name           string         `long:"name" short:"A" description:"the name of the application, defaults to a mangled value of info.title"`
	Operations     []string       `long:"operation" short:"O" description:"specify an operation to include, repeat for multiple"`
	Tags           []string       `long:"tags" description:"the tags to include, if not specified defaults to all"`
	Principal      string         `long:"principal" short:"P" description:"the model to use for the security principal"`
	Models         []string       `long:"model" short:"M" description:"specify a model to include, repeat for multiple"`
	SkipModels     bool           `long:"skip-models" description:"no models will be generated when this flag is specified"`
	SkipOperations bool           `long:"skip-operations" description:"no operations will be generated when this flag is specified"`
}

func main() {
	_, err := flags.ParseArgs(&opts, os.Args[1:])
	if err != nil {
		panic(err)
		os.Exit(1)
	}

	genOpts := generator.GenOpts{
		Spec:          string(opts.Spec),
		Target:        string(opts.Target),
		APIPackage:    opts.APIPackage,
		ModelPackage:  opts.ModelPackage,
		ServerPackage: opts.ServerPackage,
		Principal:     opts.Principal,
	}

	if !opts.SkipModels && (len(opts.Models) > 0 || len(opts.Operations) == 0) {
		if err := generator.GenerateDefinition(opts.Models, true, true, genOpts); err != nil {
			panic(err)
			os.Exit(1)
		}
	}

	if !opts.SkipOperations && (len(opts.Operations) > 0 || len(opts.Models) == 0) {
		if err := generator.GenerateServerOperation(opts.Operations, opts.Tags, true, true, genOpts); err != nil {
			panic(err)
			os.Exit(1)
		}
	}
}
