package cmd

import (
	kln "github.com/adelmoradian/kln/pkg"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

type RiList struct {
	Items []kln.ResourceIdentifier `yaml:"items"`
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
		client := kln.GetDynamicClient(kubeconfig)
		config := kln.ReadFile(file)
		err := yaml.Unmarshal(config, &riList)
		if err != nil {
			panic(err)
		}
		for _, ri := range riList.Items {
			kln.ListResources(client, ri)
		}
	},
}

func init() {
	rootCmd.AddCommand(listCmd)
}
