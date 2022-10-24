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
	labelFalse := []byte(`{"metadata":{"labels":{"kln.com/delete":"false"}}}`)
	labelTrue := []byte(`{"metadata":{"labels":{"kln.com/delete":"true"}}}`)
	client.Resource(ri.GVR).Namespace("ns").Patch(context.TODO(), "name1", types.MergePatchType, labelTrue, v1.PatchOptions{})
	client.Resource(ri.GVR).Namespace("ns3").Patch(context.TODO(), "name3", types.MergePatchType, labelFalse, v1.PatchOptions{})

	t.Run("happy - flagging resources", func(t *testing.T) {
		flagAssertion(t, client, ri.GVR, r1, true, "true")
		flagAssertion(t, client, ri.GVR, r2, false, "true")
		flagAssertion(t, client, ri.GVR, r3, true, "false")
		err := FlagForDeletion(client, ri, false)
		if err != nil {
			t.Error(err)
		}
		flagAssertion(t, client, ri.GVR, r1, true, "true")
		flagAssertion(t, client, ri.GVR, r2, true, "true")
		flagAssertion(t, client, ri.GVR, r3, true, "true")
	})

	t.Run("happy - unflag resources", func(t *testing.T) {
		err := FlagForDeletion(client, ri, true)
		if err != nil {
			t.Error(err)
		}
		flagAssertion(t, client, ri.GVR, r1, true, "true")
		flagAssertion(t, client, ri.GVR, r2, true, "false")
		flagAssertion(t, client, ri.GVR, r3, true, "false")
	})
}

func flagAssertion(t *testing.T, client *dynamicfake.FakeDynamicClient, gvr schema.GroupVersionResource, resource *unstructured.Unstructured, hasLabel bool, flagIs string) {
	t.Helper()

	ns := resource.Object["metadata"].(map[string]interface{})["namespace"].(string)
	name := resource.Object["metadata"].(map[string]interface{})["name"].(string)

	item, err := client.Resource(gvr).Namespace(ns).Get(context.TODO(), name, v1.GetOptions{})
	if err != nil {
		t.Error(err)
	}
	labels := item.GetLabels()
	if _, ok := labels["kln.com/delete"]; ok != hasLabel {
		t.Errorf("expectation to have the kln.com/delete label not satisfied. got %v, want %v", ok, hasLabel)
	}
	if value, ok := labels["kln.com/delete"]; ok && value != flagIs {
		t.Errorf("got label value %s, want %s", value, flagIs)
	}
}
