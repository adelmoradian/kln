/*
Copyright © 2022 NAME HERE <EMAIL ADDRESS>

*/
package cmd

import (
	"time"

	kflag "github.com/adelmoradian/kln/internal/flag"
	kutility "github.com/adelmoradian/kln/internal/utility"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

type RiList struct {
	Items []kutility.ResourceIdentifier `yaml:"items"`
}

var riList RiList

// listCmd represents the list command
var listCmd = &cobra.Command{
	Use:   "list",
	Short: "A brief description of your command",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	Run: func(cmd *cobra.Command, args []string) {
		client := kutility.GetDynamicClient(kubeconfig)
		config := kutility.ReadFile(file)
		err := yaml.Unmarshal(config, &riList)
		if err != nil {
			panic(err)
		}
		for _, ri := range riList.Items {
			kutility.InfoLog.Printf("--- GVR: %s, Name: %s, Description: %s\n", ri.GVR, ri.Name, ri.Description)
			list := kflag.ListResources(client, ri)
			for _, item := range list {
				age := item.GetCreationTimestamp()
				ns := item.GetNamespace()
				name := item.GetName()
				kutility.InfoLog.Printf("Name: %s, Namespace: %s, Age: %s\n", name, ns, time.Since(age.Time))
			}
		}
	},
}

func init() {
	rootCmd.AddCommand(listCmd)
	listCmd.Flags().StringVarP(&file, "file", "f", "./kln.yaml", "Path to yaml file containing the resource identifiers")

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// listCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// listCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
