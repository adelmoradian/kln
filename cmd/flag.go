package cmd

import (
	kln "github.com/adelmoradian/kln/pkg"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

var undoSwitch bool

var flagCmd = &cobra.Command{
	Use:   "flag",
	Short: "Flags objects for deletion",
	Long: `Flags objects for deletion by adding a "kln/com/delete: true"
label. By providing the undo flag, it "undo" the flagging by
changing the label from true to to false`,
	Example: `# Flag for deletion by patching label "kln.com/delete=true"
kln flag

# Undo the deletion flag by patching label "kln.com/delete=false"
kln flag -u
`,
	Run: func(cmd *cobra.Command, args []string) {
		dynamicClient := kln.GetDynamicClient(kubeconfig)
		config := kln.ReadFile(file)
		err := yaml.Unmarshal(config, &riList)
		if err != nil {
			panic(err)
		}
		for _, ri := range riList.Items {
			err := kln.FlagForDeletion(dynamicClient, ri, undoSwitch)
			if err != nil {
				kln.ErrorLog.Println(err)
			}
		}
	},
}

func init() {
	rootCmd.AddCommand(flagCmd)
	flagCmd.Flags().BoolVarP(&undoSwitch, "undo", "u", false, "When provided, will label kln.com/delete: false")
}
