package kln

import (
	"context"
	"testing"

	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	dynamicfake "k8s.io/client-go/dynamic/fake"

	"k8s.io/apimachinery/pkg/types"
)

func TestFlagForDeletion(t *testing.T) {
	scheme := runtime.NewScheme()
	scheme.AddKnownTypeWithName(schema.GroupVersionKind{Group: aGVRK.GVR.Group, Version: aGVRK.GVR.Version, Kind: aGVRK.Kind + "List"}, &unstructured.Unstructured{})
	scheme.AddKnownTypeWithName(schema.GroupVersionKind{Group: fakeGVRK.GVR.Group, Version: fakeGVRK.GVR.Version, Kind: fakeGVRK.Kind + "List"}, &unstructured.Unstructured{})
	client := dynamicfake.NewSimpleDynamicClient(scheme)
	_, err := client.Resource(aGVRK.GVR).Namespace("ns").Create(context.TODO(), r1, v1.CreateOptions{})
	if err != nil {
		t.Error(err)
	}
	_, err = client.Resource(aGVRK.GVR).Namespace("ns").Create(context.TODO(), r2, v1.CreateOptions{})
	if err != nil {
		t.Error(err)
	}
	_, err = client.Resource(aGVRK.GVR).Namespace("ns3").Create(context.TODO(), r3, v1.CreateOptions{})
	if err != nil {
		t.Error(err)
	}

	ri := ResourceIdentifier{GVR: aGVRK.GVR, MinAge: 0.5}
	ns := "ns"
	name := "name2"

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
