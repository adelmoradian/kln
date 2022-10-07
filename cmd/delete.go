package cmd

import (
	kdelete "github.com/adelmoradian/kln/internal/deleteresources"
	kutility "github.com/adelmoradian/kln/internal/utility"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

var propogationPolicy string

// deleteCmd represents the delete command
var deleteCmd = &cobra.Command{
	Use:   "delete",
	Short: "Deletes the flagged resources",
	Long: `Deletes the the resources that have a "kln.com/delete=true" label.
IMPORTANT -> In order for this to work properly, a resource identifier
must be provided with a valid gvr (Group, Version Resource). Delete
command does search all the api objects available in the cluster for
"kln.com/delete=true" label. Instead it only searchs the provided gvr(s).
Delete command gets the gvr(s) from the resource identifier yaml file.
However it only cares about the gvr(s) and not any other criteria that
may be available. For example assume that you run a cron job that flags
completed jobs older 100 hours. Then if you run the delete command with a
resources identifier that has "batch/v1 jobs" and "apps/v1 deployments"
gvrs with some other criteria, the delete command will only search deployments
and jobs for objects that are flagged for deletion and will delete them.
It will NOT flag and delete any new objects from the new criteria. Obviously
the "kln.com/delete" label can be manually changed as well. Currently kln
does not record anything about the objects which it flags or deletes.`,
	Run: func(cmd *cobra.Command, args []string) {
		dynamicClient := kutility.GetDynamicClient(kubeconfig)
		config := kutility.ReadFile(file)
		err := yaml.Unmarshal(config, &riList)
		if err != nil {
			panic(err)
		}
		for _, ri := range riList.Items {
			err := kdelete.DeleteResources(dynamicClient, ri.GVR)
			if err != nil {
				kutility.ErrorLog.Println(err)
			}
		}
	},
}

func init() {
	rootCmd.AddCommand(deleteCmd)
}
