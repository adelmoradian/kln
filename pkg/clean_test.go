package kln

import (
	"context"
	"testing"

	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	dynamicfake "k8s.io/client-go/dynamic/fake"
)

func TestReal(t *testing.T) {
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
	response3, err := client.Resource(aGVRK.GVR).Namespace("ns3").Create(context.TODO(), r3, v1.CreateOptions{})
	if err != nil {
		t.Error(err)
	}

	patchTrue := []byte(`{"metadata":{"labels":{"kln.com/delete":"true"}}}`)
	patchFalse := []byte(`{"metadata":{"labels":{"kln.com/delete":"false"}}}`)
	ri := ResourceIdentifier{GVR: aGVRK.GVR}
	ns1 := "ns"
	name1 := "name1"
	ns2 := "ns"
	name2 := "name2"
	client.Resource(ri.GVR).Namespace(ns1).Patch(context.TODO(), name1, types.MergePatchType, patchTrue, v1.PatchOptions{})
	response2, _ := client.Resource(ri.GVR).Namespace(ns2).Patch(context.TODO(), name2, types.MergePatchType, patchFalse, v1.PatchOptions{})
	t.Run("happy - deletes only the resource which is labeled", func(t *testing.T) {
		err := DeleteResources(client, ri.GVR)
		if err != nil {
			t.Errorf("got err %s", err)
		}
		got, _ := client.Resource(ri.GVR).List(context.TODO(), v1.ListOptions{})
		want := []map[string]interface{}{response2.Object, response3.Object}
		if len(got.Items) != len(want) {
			t.Errorf("Expected %d items but got %d", len(want), len(got.Items))
		}
		for _, wantItem := range want {
			if !equalityCheck(wantItem, got.Items) {
				t.Errorf("want --->\n%s\nbut did not find it in --->\n%v\n", wantItem, got)
			}

		}
	})

}
