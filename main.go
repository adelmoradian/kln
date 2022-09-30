package main

import (
	"flag"
	"fmt"
	"path/filepath"
	"time"

	kln "github.com/adelmoradian/kln/internal"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
)

func main() {
	a := kln.ResourceIdentifier{
		Gvr: schema.GroupVersionResource{
			Group:    "tekton.dev",
			Version:  "v1beta1",
			Resource: "pipelineruns",
		},
		MinAge:     160,
		ApiVersion: "tekton.dev/v1beta1",
		Kind:       "PipelineRun",
		// Metadata: map[string]interface{}{
		// "namespace": "ci-shared",
		// },
		Metadata: map[string]interface{}{},
		Spec:     map[string]interface{}{},
		// Status: map[string]interface{}{},
		Status: map[string]interface{}{
			"conditions": []interface{}{
				map[string]interface{}{
					"status": "False",
					"type":   "Succeeded",
					"reason": "CouldntGetPipeline",
				},
			},
		},
	}
	var kubeconfig *string
	if home := homedir.HomeDir(); home != "" {
		kubeconfig = flag.String("kubeconfig", filepath.Join(home, ".kube", "config"), "(optional) abs path to kubeconfig file")
	} else {
		kubeconfig = flag.String("kubeconfig", "", "absolute path to the kubeconfig file")
	}
	flag.Parse()

	config, err := clientcmd.BuildConfigFromFlags("", *kubeconfig)
	if err != nil {
		panic(err)
	}
	client, err := dynamic.NewForConfig(config)
	if err != nil {
		panic(err)
	}

	list := kln.ListResources(client, a)
	for _, item := range list {
		age := item.GetCreationTimestamp()
		ns := item.GetNamespace()
		name := item.GetName()
		apiversion := item.GetAPIVersion()
		kind := item.GetKind()
		stuff := item.Object["status"].(map[string]interface{})["conditions"].([]interface{})[0].(map[string]interface{})
		status := stuff["status"].(string)
		typeof := stuff["type"].(string)
		reason := stuff["reason"]
		fmt.Printf("--- Name: %s, NS: %s, Age: %s, apiVersion: %s, kind: %s --- %s is %s because %s\n", name, ns, time.Since(age.Time), apiversion, kind, typeof, status, reason)
	}

}
