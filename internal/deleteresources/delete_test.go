package deleteresources

import (
	"context"
	"testing"
	"time"

	kutility "github.com/adelmoradian/kln/internal/utility"
	"k8s.io/apimachinery/pkg/api/equality"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	dynamicfake "k8s.io/client-go/dynamic/fake"
)

const (
	RFC3339 = "2006-01-02T15:04:05Z07:00"
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

	resourceFake = kutility.ResourceIdentifier{
		GVR: schema.GroupVersionResource{
			Group:    "fake",
			Version:  "fakeVersion",
			Resource: "fakes",
		},
		ApiVersion: "fake/fakeVersion",
		Kind:       "Fake",
	}
)

func TestReal(t *testing.T) {
	resource1 := CreateResource(customGVR, map[string]interface{}{"namespace": "ns", "name": "name1"}, status1)
	resource2 := CreateResource(customGVR, map[string]interface{}{"namespace": "ns", "name": "name2"}, status2)
	resource3 := CreateResource(customGVR, map[string]interface{}{"namespace": "ns3", "name": "name3"}, status3)
	client := SetupFakeDynamicClient(t, resource1, resourceFake)
	manifest1 := NewUnstructured(t, resource1, time.Now().Add(-10*time.Minute).Format(RFC3339))
	manifest2 := NewUnstructured(t, resource2, time.Now().Add(-40*time.Minute).Format(RFC3339))
	manifest3 := NewUnstructured(t, resource3, time.Now().Add(-70*time.Minute).Format(RFC3339))
	ApplyResource(t, client, resource1, manifest1)
	ApplyResource(t, client, resource2, manifest2)
	response3 := ApplyResource(t, client, resource3, manifest3)
	patchTrue := []byte(`{"metadata":{"labels":{"kln.com/delete":"true"}}}`)
	patchFalse := []byte(`{"metadata":{"labels":{"kln.com/delete":"false"}}}`)
	ri := kutility.ResourceIdentifier{GVR: resource1.GVR}
	ns1 := resource1.Metadata["namespace"].(string)
	name1 := resource1.Metadata["name"].(string)
	ns2 := resource2.Metadata["namespace"].(string)
	name2 := resource2.Metadata["name"].(string)
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

func CreateResource(gvr schema.GroupVersionResource, meta map[string]interface{}, status map[string]interface{}) kutility.ResourceIdentifier {
	ri := kutility.ResourceIdentifier{
		GVR:        gvr,
		ApiVersion: "agroup/aversion",
		Kind:       "AKind",
		Metadata:   meta,
		Status:     status,
	}
	return ri
}

func SetupFakeDynamicClient(t *testing.T, riList ...kutility.ResourceIdentifier) *dynamicfake.FakeDynamicClient {
	t.Helper()
	scheme := runtime.NewScheme()
	for _, ri := range riList {
		scheme.AddKnownTypeWithName(schema.GroupVersionKind{Group: ri.GVR.Group, Version: ri.GVR.Version, Kind: ri.Kind + "List"}, &unstructured.Unstructured{})
	}
	client := dynamicfake.NewSimpleDynamicClient(scheme)
	return client
}

func ApplyResource(t *testing.T, client *dynamicfake.FakeDynamicClient, ri kutility.ResourceIdentifier, rm *unstructured.Unstructured) *unstructured.Unstructured {
	t.Helper()
	ns := ri.Metadata["namespace"].(string)
	response, err := client.Resource(ri.GVR).Namespace(ns).Create(context.TODO(), rm, v1.CreateOptions{})
	if err != nil {
		t.Error(err)
	}
	return response
}

func NewUnstructured(t *testing.T, ri kutility.ResourceIdentifier, creationTimestamp string) *unstructured.Unstructured {
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
