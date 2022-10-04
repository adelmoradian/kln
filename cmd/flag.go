package cmd

import (
	kutility "github.com/adelmoradian/kln/internal/utility"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

var undoSwitch bool

var flagCmd = &cobra.Command{
	Use:   "flag",
	Short: "Flags resources for deletion",
	Long: `Flags the resources for deletion by adding a "kln/com/delete: true"
annotation. By providing the undo flag, you can "undo" the flagging by
changing the annotation to false. Examples:
# default behavior
kln flag
# provide relative path to resources
kln flag -f ../path/to/kln.yaml
# undo
kln flag -u`,
	Run: func(cmd *cobra.Command, args []string) {
		dynamicClient := kutility.GetDynamicClient(kubeconfig)
		config := kutility.ReadFile(file)
		err := yaml.Unmarshal(config, &riList)
		if err != nil {
			panic(err)
		}
		for _, ri := range riList.Items {
			err := ri.FlagForDeletion(dynamicClient, undoSwitch)
			if err != nil {
				kutility.WarningLog.Println(err)
			}
		}
	},
}

func init() {
	rootCmd.AddCommand(flagCmd)
	flagCmd.Flags().StringVarP(&file, "file", "f", "./kln.yaml", "Relative path to yaml file containing the resource identifiers")
	flagCmd.Flags().BoolVarP(&undoSwitch, "undo", "u", false, "When set to true, will annotate kln/com/delete: false (default: false)")
}
