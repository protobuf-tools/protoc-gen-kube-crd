// Copyright 2021 The protobuf-tools Authors
// SPDX-License-Identifier: Apache-2.0

// Command protoc-gen-kube generates the Kubernetes controller APIs from Protocol Buffer schemas.
package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"path/filepath"

	"github.com/go-logr/logr"
	"github.com/go-logr/zapr"
	"github.com/spf13/pflag"
	"go.uber.org/zap"
	"k8s.io/gengo/args"
	klog "k8s.io/klog/v2"

	"github.com/protobuf-tools/protoc-gen-kube/pkg/generator/kubetype"
	"github.com/protobuf-tools/protoc-gen-kube/pkg/scanner"
	"github.com/protobuf-tools/protoc-gen-kube/pkg/version"
)

func main() {
	if err := gen(); err != nil {
		fmt.Fprintf(os.Stderr, "protoc-gen-kube: %#v\n", err)
	}
}

func gen() error {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()

	zl, err := zap.NewDevelopment()
	if err != nil {
		return fmt.Errorf("create Development zap logger: %w", err)
	}
	logf := zapr.NewLogger(zl)
	ctx = logr.NewContext(ctx, logf)

	// inject third-party packages klog to zapr
	klog.SetLogger(logf)

	gen := args.Default().WithoutDefaultFlagParsing() // WithoutDefaultFlagParsing is suck

	// setup flags
	showVersion := flag.Bool("version", false, "print the version and exit")
	pflag.CommandLine.AddGoFlagSet(flag.CommandLine)
	gen.AddFlags(pflag.CommandLine)
	pflag.Parse()

	if *showVersion {
		fmt.Fprintf(os.Stderr, "protoc-gen-kube@%s\n", version.Version())
		os.Exit(0)
	}

	// setup template
	gen.GeneratedByCommentTemplate = `// Code generated by protoc-gen-kube. DO NOT EDIT.`
	gen.GoHeaderFilePath = filepath.Join(args.DefaultSourceTree(), "hack/boilerplate/boilerplate.go.txt")

	// execute
	errc := make(chan error, 1)
	go func() {
		defer close(errc)

		if err := gen.Execute(
			kubetype.NameSystems("", nil),
			kubetype.DefaultNameSystem(),
			(&scanner.Scanner{}).WithContext(ctx).Scan,
		); err != nil {
			errc <- fmt.Errorf("execute Generator: %w", err)
			return
		}

		logf.Info("completed successfully")
	}()

	select {
	case err := <-errc:
		if err != nil {
			return err
		}
	case <-ctx.Done():
		return ctx.Err()
	}

	return nil
}
