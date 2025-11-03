package usage

import (
	"flag"
	"fmt"
	"os"
	"text/tabwriter"

	"github.com/kelseyhightower/envconfig"
)

func ShowHelp(envconfigPrefix string, optionsSpec interface{}) {
	usagePadding := 4
	tabs := tabwriter.NewWriter(os.Stdout, 1, 0, usagePadding, ' ', 0)
	_ = envconfig.Usagef(envconfigPrefix, optionsSpec, tabs, envconfig.DefaultTableFormat)
	_ = tabs.Flush()
	// use below approach instead of flag.Usage() for customized output: https://stackoverflow.com/a/23726033/4292075
	fmt.Println("\nIn addition, the following CLI arguments are supported")
	flag.PrintDefaults()
	fmt.Println()
}
