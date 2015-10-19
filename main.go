package main

import (
	"flag"
	"os"

	"github.com/aiyi/swagger-gin/generator"
)

func main() {
	spec := flag.String("spec", "./swagger.json", "the spec file to use")
	target := flag.String("target", "./", "the directory for generating the files")

	flag.Parse()

	genOpts := generator.GenOpts{
		Spec:         *spec,
		Target:       *target,
		APIPackage:   "operations",
		ModelPackage: "models",
	}

	if err := generator.GenerateDefinition(true, true, genOpts); err != nil {
		panic(err)
		os.Exit(1)
	}

	if err := generator.GenerateServerOperation(true, true, genOpts); err != nil {
		panic(err)
		os.Exit(1)
	}
}
