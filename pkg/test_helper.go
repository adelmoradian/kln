package kln

import (
	"context"
	"testing"

	"k8s.io/apimachinery/pkg/api/equality"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	dynamicfake "k8s.io/client-go/dynamic/fake"
)

var (
	customGVR = schema.GroupVersionResource{
		Group:    "agroup",
		Version:  "aversion",
		Resource: "akinds",
	}

	fakeGVR = schema.GroupVersionResource{
		Group:    "fake",
		Version:  "fakeVersion",
		Resource: "fakes",
	}

	status1 = map[string]interface{}{
		"foo":    "bar",
		"tomato": "potato",
	}
	status2 = map[string]interface{}{
		"foo": "bar",
		"status": map[string]interface{}{
			"baz":    map[string]interface{}{"deep": "nest"},
			"tomato": "potato",
		},
	}

	status3 = map[string]interface{}{
		"foo": "notBar",
		"status": map[string]interface{}{
			"baz": "fail",
		},
	}

	resourceFake = ResourceIdentifier{
		GVR: schema.GroupVersionResource{
			Group:    "fake",
			Version:  "fakeVersion",
			Resource: "fakes",
		},
		ApiVersion: "fake/fakeVersion",
		Kind:       "Fake",
	}
)

type listTestCases struct {
	name string
	ri   ResourceIdentifier
	want []map[string]interface{}
	skip bool
}

func CreateResource(gvr schema.GroupVersionResource, meta map[string]interface{}, status map[string]interface{}) ResourceIdentifier {
	ri := ResourceIdentifier{
		GVR:        gvr,
		ApiVersion: "agroup/aversion",
		Kind:       "AKind",
		Metadata:   meta,
		Status:     status,
	}
	return ri
}

func SetupFakeDynamicClient(t *testing.T, riList ...ResourceIdentifier) *dynamicfake.FakeDynamicClient {
	t.Helper()
	scheme := runtime.NewScheme()
	for _, ri := range riList {
		scheme.AddKnownTypeWithName(schema.GroupVersionKind{Group: ri.GVR.Group, Version: ri.GVR.Version, Kind: ri.Kind + "List"}, &unstructured.Unstructured{})
	}
	client := dynamicfake.NewSimpleDynamicClient(scheme)
	return client
}

func ApplyResource(t *testing.T, client *dynamicfake.FakeDynamicClient, ri ResourceIdentifier, rm *unstructured.Unstructured) *unstructured.Unstructured {
	t.Helper()
	ns := ri.Metadata["namespace"].(string)
	response, err := client.Resource(ri.GVR).Namespace(ns).Create(context.TODO(), rm, v1.CreateOptions{})
	if err != nil {
		t.Error(err)
	}
	return response
}

func NewUnstructured(t *testing.T, ri ResourceIdentifier, creationTimestamp string) *unstructured.Unstructured {
	t.Helper()
	ns := ri.Metadata["namespace"].(string)
	name := ri.Metadata["name"].(string)
	return &unstructured.Unstructured{
		Object: map[string]interface{}{
			"apiVersion": ri.ApiVersion,
			"kind":       ri.Kind,
			"metadata": map[string]interface{}{
				"creationTimestamp": creationTimestamp,
				"namespace":         ns,
				"name":              name,
			},
			"status": ri.Status,
		},
	}
}

func equalityCheck(wantItem map[string]interface{}, got []unstructured.Unstructured) bool {
	for _, gotItem := range got {
		if equality.Semantic.DeepEqual(gotItem.Object, wantItem) {
			return true
		}
	}
	return false
}

func flagAssertion(t *testing.T, client *dynamicfake.FakeDynamicClient, gvr schema.GroupVersionResource, name, ns, key string, err error) map[string]string {
	t.Helper()

	if err != nil {
		t.Error(err)
	}
	item, err := client.Resource(gvr).Namespace(ns).Get(context.TODO(), name, v1.GetOptions{})
	if err != nil {
		t.Error(err)
	}
	labels := item.GetLabels()
	if labels == nil {
		t.Errorf("got no labels")
	}
	if _, ok := labels["kln.com/delete"]; !ok {
		t.Errorf("%s does not contain the kln.com/delete key", labels)
	}
	if value, _ := labels["kln.com/delete"]; value != key {
		t.Errorf("got %s, want %s", value, key)
	}
	return labels
}
