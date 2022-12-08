package main

import (
	"flag"
	"os"

	"github.com/Azure/kubelogin/pkg/cmd"
	"github.com/spf13/pflag"
	"k8s.io/klog"
)

func main() {
	klog.InitFlags(nil)
	pflag.CommandLine.AddGoFlag(flag.CommandLine.Lookup("v"))
	pflag.CommandLine.AddGoFlag(flag.CommandLine.Lookup("logtostderr"))
	_ = pflag.CommandLine.Set("logtostderr", "true")
	root := cmd.NewRootCmd(v.String())
	if err := root.Execute(); err != nil {
		os.Exit(1)
	}
}
