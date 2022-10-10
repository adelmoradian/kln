package kln

import (
	"context"
	"testing"
	"time"

	"k8s.io/apimachinery/pkg/api/equality"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"

	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
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

func TestFlagForDeletion(t *testing.T) {
	resource2 := CreateResource(customGVR, map[string]interface{}{"namespace": "ns", "name": "name2"}, status2)
	client := SetupFakeDynamicClient(t, resource2)
	manifest2 := NewUnstructured(t, resource2, time.Now().Add(-40*time.Minute).Format(RFC3339))
	ApplyResource(t, client, resource2, manifest2)
	ri := ResourceIdentifier{GVR: resource2.GVR, MinAge: 0.5}
	ns := resource2.Metadata["namespace"].(string)
	name := resource2.Metadata["name"].(string)

	t.Run("happy - resource has no existing label", func(t *testing.T) {
		err := FlagForDeletion(client, ri, false)
		flagAssertion(t, client, ri.GVR, name, ns, "true", err)
	})

	t.Run("happy - kln.com/delete label is initially false", func(t *testing.T) {
		patch := []byte(`{"metadata":{"labels":{"kln.com/delete":"false"}}}`)
		client.Resource(ri.GVR).Namespace(ns).Patch(context.TODO(), name, types.MergePatchType, patch, v1.PatchOptions{})
		err := FlagForDeletion(client, ri, false)
		flagAssertion(t, client, ri.GVR, name, ns, "true", err)
	})

	t.Run("happy - resource has some label", func(t *testing.T) {
		patch := []byte(`{"metadata":{"labels":{"foo":"bar"}}}`)
		client.Resource(ri.GVR).Namespace(ns).Patch(context.TODO(), name, types.MergePatchType, patch, v1.PatchOptions{})
		err := FlagForDeletion(client, ri, false)
		labels := flagAssertion(t, client, ri.GVR, name, ns, "true", err)
		if _, ok := labels["foo"]; !ok {
			t.Error("original label was lost")
		}
	})

	t.Run("undo - resource has no existing label", func(t *testing.T) {
		err := FlagForDeletion(client, ri, true)
		flagAssertion(t, client, ri.GVR, name, ns, "false", err)
	})

	t.Run("undo - kln.com/delete label is initially true", func(t *testing.T) {
		patch := []byte(`{"metadata":{"labels":{"kln.com/delete":"true"}}}`)
		client.Resource(ri.GVR).Namespace(ns).Patch(context.TODO(), name, types.MergePatchType, patch, v1.PatchOptions{})
		err := FlagForDeletion(client, ri, true)
		flagAssertion(t, client, ri.GVR, name, ns, "false", err)
	})

	t.Run("undo - resource has some labels", func(t *testing.T) {
		patch := []byte(`{"metadata":{"labels":{"foo":"bar"}}}`)
		client.Resource(ri.GVR).Namespace(ns).Patch(context.TODO(), name, types.MergePatchType, patch, v1.PatchOptions{})
		err := FlagForDeletion(client, ri, true)
		labels := flagAssertion(t, client, ri.GVR, name, ns, "false", err)
		if _, ok := labels["foo"]; !ok {
			t.Error("original label was lost")
		}
	})
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
