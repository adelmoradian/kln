package cmd

import (
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
	"k8s.io/client-go/util/homedir"
)

var kubeconfig string
var file string

var rootCmd = &cobra.Command{
	Use:   "kln",
	Short: "Keep your cluster clean!",
	Long: `kln finds, flags and deletes unwanted objects in your kubernetes
cluster using the user provided resource identifier yaml file`,
	Example: `# List unwated objects
kln list

# Provide path to resource identifier
kln list -f ../rltv/path/to/identifier.yaml

# Flag for deletion by patching label "kln.com/delete=true"
kln flag

# Undo the deletion flag by patching label "kln.com/delete=false"
kln flag -d=false

# Delete resources that have "kln.com/delete=true" label
kln delete`,
}

func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	rootCmd.PersistentFlags().StringVarP(&kubeconfig, "kube-config", "k", filepath.Join(homedir.HomeDir(), ".kube", "config"), "abs path to the kubeconfig file")
	rootCmd.PersistentFlags().StringVarP(&file, "file", "f", "./kln.yaml", "relative path to resource identifier yaml file")
}
