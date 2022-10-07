package cmd

import (
	kflag "github.com/adelmoradian/kln/internal/flag"
	kutility "github.com/adelmoradian/kln/internal/utility"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

type RiList struct {
	Items []kutility.ResourceIdentifier `yaml:"items"`
}

var riList RiList

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List unwanted objects",
	Long: `Lists unwanted objects according to the criteria given
in the resource identifier yaml file.`,
	Example: `# List unwated objects
kln list

# Provide path to resource identifier
kln list -f ../rltv/path/to/identifier.yaml`,
	Run: func(cmd *cobra.Command, args []string) {
		client := kutility.GetDynamicClient(kubeconfig)
		config := kutility.ReadFile(file)
		err := yaml.Unmarshal(config, &riList)
		if err != nil {
			panic(err)
		}
		for _, ri := range riList.Items {
			kflag.ListResources(client, ri)
		}
	},
}

func init() {
	rootCmd.AddCommand(listCmd)
}
