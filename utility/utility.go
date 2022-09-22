package utility

import (
	"context"
	"math/rand"
	"testing"

	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	dynamicfake "k8s.io/client-go/dynamic/fake"
)

var (
	RealGVRs = map[string]schema.GroupVersionResource{
		"Job": {
			Group:    "batch",
			Version:  "v1",
			Resource: "jobs",
		},
		"Pod": {
			Group:    "",
			Version:  "v1",
			Resource: "pods",
		},
		"Deployment": {
			Group:    "apps",
			Version:  "v1",
			Resource: "deployments",
		},
		"MyResource": {
			Group:    "mygroup",
			Version:  "v1",
			Resource: "myresources",
		},
	}
	FakeGVRs = map[string]schema.GroupVersionResource{
		"ConfigMap": {
			Group:    "",
			Version:  "v1",
			Resource: "configmaps",
		},
	}
)

func RandomString(t *testing.T, n int) string {
	t.Helper()
	var letters = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789")
	s := make([]rune, n)
	for i := range s {
		s[i] = letters[rand.Intn(len(letters))]
	}
	return string(s)
}

func SetupFakeDynamicClient(t *testing.T, GVRs ...map[string]schema.GroupVersionResource) *dynamicfake.FakeDynamicClient {
	t.Helper()
	scheme := runtime.NewScheme()
	for _, gvrMap := range GVRs {
		for kind, gvr := range gvrMap {
			scheme.AddKnownTypeWithName(schema.GroupVersionKind{Group: gvr.Group, Version: gvr.Version, Kind: kind}, &unstructured.Unstructured{})
			scheme.AddKnownTypeWithName(schema.GroupVersionKind{Group: gvr.Group, Version: gvr.Version, Kind: kind + "List"}, &unstructured.Unstructured{})
		}
	}
	client := dynamicfake.NewSimpleDynamicClient(scheme)
	return client
}

func ApplyResources(t *testing.T, client *dynamicfake.FakeDynamicClient, GVRs ...map[string]schema.GroupVersionResource) map[string]unstructured.Unstructured {
	t.Helper()
	responseMap := make(map[string]unstructured.Unstructured)
	for _, gvrMap := range GVRs {
		for kind, gvr := range gvrMap {
			resources, nsMap := ConstructResources(t, RealGVRs)
			response, err := client.Resource(gvr).Namespace(nsMap[kind]).Create(context.TODO(), resources[kind], v1.CreateOptions{})
			if err != nil {
				t.Error(err)
			}
			responseMap[kind] = *response
		}
	}
	return responseMap
}

func ConstructResources(t *testing.T, GVRs ...map[string]schema.GroupVersionResource) (map[string]*unstructured.Unstructured, map[string]string) {
	t.Helper()
	var slash string
	nsMap := make(map[string]string)
	resources := make(map[string]*unstructured.Unstructured)
	for _, gvrMap := range GVRs {

		for kind, gvr := range gvrMap {
			ns := RandomString(t, 10)
			nsMap[kind] = "namespace-" + ns
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
						"name":      gvr.Resource + "-" + RandomString(t, 10),
						"namespace": "namespace-" + ns,
					},
				},
			}
		}

	}
	return resources, nsMap
}
