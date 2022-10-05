package cmd

import (
	kflag "github.com/adelmoradian/kln/internal/flag"
	kutility "github.com/adelmoradian/kln/internal/utility"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

var undoSwitch bool

var flagCmd = &cobra.Command{
	Use:   "flag",
	Short: "Flags resources for deletion",
	Long: `Flags the resources for deletion by adding a "kln/com/delete: true"
label. By providing the undo flag, you can "undo" the flagging by
changing the label to false`,
	Run: func(cmd *cobra.Command, args []string) {
		dynamicClient := kutility.GetDynamicClient(kubeconfig)
		config := kutility.ReadFile(file)
		err := yaml.Unmarshal(config, &riList)
		if err != nil {
			panic(err)
		}
		for _, ri := range riList.Items {
			err := kflag.FlagForDeletion(dynamicClient, ri, undoSwitch)
			if err != nil {
				kutility.ErrorLog.Println(err)
			}
		}
	},
}

func init() {
	rootCmd.AddCommand(flagCmd)
	flagCmd.Flags().BoolVarP(&undoSwitch, "undo", "u", false, "When set to true, will label kln/com/delete: false (default: false)")
}
