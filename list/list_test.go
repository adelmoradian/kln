package list

import (
	"context"
	"testing"
	"time"

	"k8s.io/apimachinery/pkg/api/equality"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	dynamicfake "k8s.io/client-go/dynamic/fake"
)

var (
	resourceCustom1 = resourceIdentifier{
		gvr: schema.GroupVersionResource{
			Group:    "agroup",
			Version:  "aversion",
			Resource: "akinds",
		},
		apiVersion: "agroup/aversion",
		kind:       "AKind",
		metadata:   map[string]interface{}{"namespace": "ns", "name": "name1"},
		status: map[string]interface{}{
			"foo": "bar",
		},
	}

	resourceCustom2 = resourceIdentifier{
		gvr: schema.GroupVersionResource{
			Group:    "agroup",
			Version:  "aversion",
			Resource: "akinds",
		},
		apiVersion: "agroup/aversion",
		kind:       "AKind",
		metadata:   map[string]interface{}{"namespace": "ns", "name": "name2"},
		status: map[string]interface{}{
			"status": map[string]interface{}{
				"baz": "foo",
			},
		},
	}

	resourceCustom3 = resourceIdentifier{
		gvr: schema.GroupVersionResource{
			Group:    "agroup",
			Version:  "aversion",
			Resource: "akinds",
		},
		apiVersion: "agroup/aversion",
		kind:       "AKind",
		metadata:   map[string]interface{}{"namespace": "ns2", "name": "name3"},
		status: map[string]interface{}{
			"status": map[string]interface{}{
				"baz": "fail",
			},
		},
	}
)

type listTestCases struct {
	name               string
	ri                 resourceIdentifier
	wantResponse       []map[string]interface{}
	wantResponseLength int
}

func TestListResources(t *testing.T) {
	client := setupFakeDynamicClient(t, resourceCustom1)
	manifestCustom := newUnstructured(t, resourceCustom1, time.Now().Add(-10*time.Minute).Format(RFC3339))
	manifestCustom2 := newUnstructured(t, resourceCustom2, time.Now().Add(-40*time.Minute).Format(RFC3339))
	manifestCustom3 := newUnstructured(t, resourceCustom3, time.Now().Add(-70*time.Minute).Format(RFC3339))
	responseCustom1 := ApplyResource(t, client, resourceCustom1, manifestCustom)
	responseCustom2 := ApplyResource(t, client, resourceCustom2, manifestCustom2)
	responseCustom3 := ApplyResource(t, client, resourceCustom3, manifestCustom3)

	listTests := []listTestCases{
		{
			name:               "finds resource given only a valid gvr",
			ri:                 resourceIdentifier{gvr: resourceCustom1.gvr},
			wantResponseLength: 3,
			wantResponse:       []map[string]interface{}{responseCustom1.Object, responseCustom2.Object, responseCustom3.Object},
		},
		{
			name:               "returns items that are older than 0.5 hours",
			ri:                 resourceIdentifier{gvr: resourceCustom1.gvr, age: 0.5},
			wantResponseLength: 2,
			wantResponse:       []map[string]interface{}{responseCustom2.Object, responseCustom3.Object},
		},
		{
			name:               "returns items based on metadata",
			ri:                 resourceIdentifier{gvr: resourceCustom1.gvr, metadata: map[string]interface{}{"namespace": "ns"}},
			wantResponseLength: 2,
			wantResponse:       []map[string]interface{}{responseCustom1.Object, responseCustom2.Object},
		},
	}

	for _, tc := range listTests {
		t.Run(tc.name, func(t *testing.T) {
			got, _ := ListResources(client, tc.ri)
			if tc.wantResponseLength != len(got) {
				t.Errorf("Expected %d items but got %d", tc.wantResponseLength, len(got))
			}
			for _, wantItem := range tc.wantResponse {
				if !equalityCheck(wantItem, got) {
					t.Errorf("did not find %s in %v\n", wantItem, got)
				}
			}
		})
	}
}

func setupFakeDynamicClient(t *testing.T, riList ...resourceIdentifier) *dynamicfake.FakeDynamicClient {
	t.Helper()
	scheme := runtime.NewScheme()
	for _, ri := range riList {
		scheme.AddKnownTypeWithName(schema.GroupVersionKind{Group: ri.gvr.Group, Version: ri.gvr.Version, Kind: ri.kind + "List"}, &unstructured.Unstructured{})
	}
	client := dynamicfake.NewSimpleDynamicClient(scheme)
	return client
}

func ApplyResource(t *testing.T, client *dynamicfake.FakeDynamicClient, ri resourceIdentifier, rm *unstructured.Unstructured) *unstructured.Unstructured {
	t.Helper()
	ns := ri.metadata["namespace"].(string)
	response, err := client.Resource(ri.gvr).Namespace(ns).Create(context.TODO(), rm, v1.CreateOptions{})
	if err != nil {
		t.Error(err)
	}
	return response
}

func newUnstructured(t *testing.T, ri resourceIdentifier, creationTimestamp string) *unstructured.Unstructured {
	t.Helper()
	ns := ri.metadata["namespace"].(string)
	name := ri.metadata["name"].(string)
	return &unstructured.Unstructured{
		Object: map[string]interface{}{
			"apiVersion": ri.apiVersion,
			"kind":       ri.kind,
			"metadata": map[string]interface{}{
				"creationTimestamp": creationTimestamp,
				"namespace":         ns,
				"name":              name,
			},
			"status": ri.status,
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
