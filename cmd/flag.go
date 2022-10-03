package cmd

import (
	"flag"
	"fmt"
)

const flagSetName = "flag"

type FlagCommand struct {
	fs         *flag.FlagSet
	configFile string
}

func NewFlagCommand() *FlagCommand {
	fc := &FlagCommand{
		fs: flag.NewFlagSet(flagSetName, flag.ContinueOnError),
	}
	fc.fs.StringVar(&fc.configFile, "f", "./kln.yaml", "Config file containing the resource definitions")
	return fc
}

func (fc *FlagCommand) Name() string {
	return fc.fs.Name()
}

func (fc *FlagCommand) Init(args []string) error {
	return fc.fs.Parse(args)
}

func (fc *FlagCommand) Run() {
	fmt.Println(fc.configFile)
}
