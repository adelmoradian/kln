package main

import (
	"context"
	"fmt"
	"math/rand"
	"testing"

	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	dynamicfake "k8s.io/client-go/dynamic/fake"
)

var (
	GVRs = map[string]schema.GroupVersionResource{
		"Job": {
			Group:    "batch",
			Version:  "v1",
			Resource: "job",
		},
		"Pod": {
			Group:    "",
			Version:  "v1",
			Resource: "pod",
		},
		"Deployment": {
			Group:    "apps",
			Version:  "v1",
			Resource: "deployment",
		},
		"MyResource": {
			Group:    "mygroup",
			Version:  "v1",
			Resource: "myresource",
		},
	}
	FakeGVR = map[string]schema.GroupVersionResource{
		"ConfigMap": {
			Group:    "",
			Version:  "v1",
			Resource: "configmap",
		},
	}
)

func TestList(t *testing.T) {

	client := dynamicfake.NewSimpleDynamicClient(runtime.NewScheme())
	resources := constructResources(t)
	responseList := applyResources(t, resources, client)

	t.Run("Resources exist", func(t *testing.T) {
		t.Skip("TODO")
		got, err := GetResources(client, GVRs)
		want := responseList
		if err != nil {
			t.Error(err)
		}
		if got != want {
			t.Errorf("got %v, want %v", got, want)
		}
	})

	t.Run("Resources do not exist", func(t *testing.T) {
		t.Skip("TODO")
		got, err := GetResources(client, FakeGVR)
		if err == nil {
			t.Error("Expected an error but did not get")
		}
	})

	t.Run("Mix of existing and non existing resources", func(t *testing.T) {
		t.Skip("TODO")
	})

	// 	t.Run("Can list k8s resources", func(t *testing.T) {
	// 		_, err := client.Resource(jobGVR).Create(context.TODO(), job, v1.CreateOptions{})
	// 		if err != nil {
	// 			fmt.Println(err)
	// 			fmt.Println("potato------")
	// 		}
	// 		a, err := client.Resource(jobGVR).Get(context.TODO(), "some-name", v1.GetOptions{})
	// 		fmt.Print(err)
	// 		fmt.Println("sfasdfdsa------")
	// 		fmt.Print(a)

	// 		if err != nil {
	// 			t.Error(err)
	// 		}
	// 	})
}
func randomString(t *testing.T, n int) string {
	t.Helper()
	var letters = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789")
	s := make([]rune, n)
	for i := range s {
		s[i] = letters[rand.Intn(len(letters))]
	}
	return string(s)
}

func constructResources(t *testing.T) map[string]*unstructured.Unstructured {
	t.Helper()
	var slash string
	resources := make(map[string]*unstructured.Unstructured)
	for kind, gvr := range GVRs {
		if gvr.Group == "" {
			slash = ""
		} else {
			slash = "/"
		}

		resources[kind] = &unstructured.Unstructured{
			Object: map[string]interface{}{
				"kind":       kind,
				"apiVersion": gvr.Group + slash + gvr.Version,
				"metadata": map[string]interface{}{
					"name":      gvr.Resource + "-" + randomString(t, 10),
					"namespace": "",
				},
			},
		}
	}
	return resources
}

func applyResources(t *testing.T, resources map[string]*unstructured.Unstructured, client *dynamicfake.FakeDynamicClient) []*unstructured.Unstructured {
	t.Helper()
	responseList := make([]*unstructured.Unstructured, 0)
	for kind, gvr := range GVRs {
		response, err := client.Resource(gvr).Namespace("").Create(context.TODO(), resources[kind], v1.CreateOptions{})
		if err != nil {
			t.Error(err)
		}
		fmt.Printf("--- APIVersion: %s - Kind: %s - Name: %s\n", response.GetAPIVersion(), response.GetKind(), response.GetName())
		responseList = append(responseList, response)
	}
	return responseList
}
