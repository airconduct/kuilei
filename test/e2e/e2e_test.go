package e2e

import (
	"flag"
	"os"
	"testing"

	"github.com/spf13/pflag"
)

func TestMain(m *testing.M) {
	flags := pflag.NewFlagSet("tess-cluster-e2e", pflag.ExitOnError)
	flag.CommandLine.VisitAll(func(f *flag.Flag) {
		if f.Name != "kubeconfig" {
			flags.AddGoFlag(f)
		}
	})
	pflag.CommandLine = flags
	// TODO: build app and bind flags
	pflag.Parse()
	// TODO: add test cases
	os.Exit(m.Run())
}
