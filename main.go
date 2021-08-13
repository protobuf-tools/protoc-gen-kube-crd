// Copyright 2021 The protobuf-tools Authors
// SPDX-License-Identifier: Apache-2.0

// Command protoc-gen-kube generates the Kubernetes controller APIs from Protocol Buffer schemas.
package main

import (
	"flag"
	"fmt"
	"os"

	"google.golang.org/protobuf/compiler/protogen"

	"github.com/protobuf-tools/protoc-gen-kube/pkg/generator"
)

// version is a protoc-gen-kube vesion.
var version = "v0.0.0"

func main() {
	showVersion := flag.Bool("version", false, "print the version and exit")
	flag.Parse()

	if *showVersion {
		fmt.Fprintf(os.Stderr, "protoc-gen-kube %s\n", version)
		os.Exit(0)
	}

	var flags flag.FlagSet
	g := protogen.Options{
		ParamFunc:         flags.Set,
		ImportRewriteFunc: func(imp protogen.GoImportPath) protogen.GoImportPath { return "" },
	}

	g.Run(func(gen *protogen.Plugin) error {
		gen.SupportedFeatures = generator.SupportedFeatures

		for _, f := range gen.Files {
			if f.Generate {
				generator.GenerateFile(gen, f)
			}
		}

		return nil
	})
}