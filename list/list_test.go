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
	resource1 = resourceIdentifier{
		gvr: schema.GroupVersionResource{
			Group:    "agroup",
			Version:  "aversion",
			Resource: "akinds",
		},
		apiVersion: "agroup/aversion",
		kind:       "AKind",
		metadata:   map[string]interface{}{"namespace": "ns", "name": "name1"},
		spec: map[string]interface{}{
			"foo": "bar",
			"baz": map[string]interface{}{
				"foo": "bar",
			},
		},
		status: map[string]interface{}{
			"foo":    "bar",
			"tomato": "potato",
		},
	}

	resource2 = resourceIdentifier{
		gvr: schema.GroupVersionResource{
			Group:    "agroup",
			Version:  "aversion",
			Resource: "akinds",
		},
		apiVersion: "agroup/aversion",
		kind:       "AKind",
		metadata:   map[string]interface{}{"namespace": "ns", "name": "name2"},
		spec: map[string]interface{}{
			"foo": "bar",
		},
		status: map[string]interface{}{
			"foo": "bar",
			"status": map[string]interface{}{
				"baz":    map[string]interface{}{"deep": "nest"},
				"tomato": "potato",
			},
		},
	}

	resource3 = resourceIdentifier{
		gvr: schema.GroupVersionResource{
			Group:    "agroup",
			Version:  "aversion",
			Resource: "akinds",
		},
		apiVersion: "agroup/aversion",
		kind:       "AKind",
		metadata:   map[string]interface{}{"namespace": "ns3", "name": "name3"},
		status: map[string]interface{}{
			"foo": "notBar",
			"status": map[string]interface{}{
				"baz": "fail",
			},
		},
	}

	resourceFake = resourceIdentifier{
		gvr: schema.GroupVersionResource{
			Group:    "fake",
			Version:  "fakeVersion",
			Resource: "fakes",
		},
		apiVersion: "fake/fakeVersion",
		kind:       "Fake",
	}
)

type listTestCases struct {
	name string
	ri   resourceIdentifier
	want []map[string]interface{}
	skip bool
}

func TestListResources(t *testing.T) {
	client := setupFakeDynamicClient(t, resource1, resourceFake)
	manifest1 := newUnstructured(t, resource1, time.Now().Add(-10*time.Minute).Format(RFC3339))
	manifest2 := newUnstructured(t, resource2, time.Now().Add(-40*time.Minute).Format(RFC3339))
	manifest3 := newUnstructured(t, resource3, time.Now().Add(-70*time.Minute).Format(RFC3339))
	response1 := ApplyResource(t, client, resource1, manifest1)
	response2 := ApplyResource(t, client, resource2, manifest2)
	response3 := ApplyResource(t, client, resource3, manifest3)

	listTests := []listTestCases{
		{
			name: "happy - finds resources given only gvr",
			ri:   resourceIdentifier{gvr: resource1.gvr},
			want: []map[string]interface{}{response1.Object, response2.Object, response3.Object},
			skip: false,
		},
		{
			name: "sad - finds resources given only gvr",
			ri:   resourceIdentifier{gvr: resourceFake.gvr},
			want: []map[string]interface{}{},
			skip: false,
		},
		{
			name: "happy - finds resources given minAge",
			ri:   resourceIdentifier{gvr: resource1.gvr, age: 0.5},
			want: []map[string]interface{}{response2.Object, response3.Object},
			skip: false,
		},
		{
			name: "sad - finds resources given minAge",
			ri:   resourceIdentifier{gvr: resource1.gvr, age: 1.5},
			want: []map[string]interface{}{},
			skip: false,
		},
		{
			name: "happy - finds resources given metadata",
			ri:   resourceIdentifier{gvr: resource1.gvr, metadata: map[string]interface{}{"namespace": "ns"}},
			want: []map[string]interface{}{response1.Object, response2.Object},
			skip: false,
		},
		{
			name: "sad - finds resources given metadata",
			ri:   resourceIdentifier{gvr: resource1.gvr, metadata: map[string]interface{}{"namespace": "ns", "name": "fake"}},
			want: []map[string]interface{}{},
			skip: false,
		},
		{
			name: "happy - finds resources given minAge and metadata",
			ri:   resourceIdentifier{gvr: resource1.gvr, metadata: map[string]interface{}{"namespace": "ns"}, age: 0.5},
			want: []map[string]interface{}{response2.Object},
			skip: false,
		},
		{
			name: "sad - finds resources given minAge and metadata",
			ri:   resourceIdentifier{gvr: resource1.gvr, metadata: map[string]interface{}{"name": "fake"}, age: 0.5},
			want: []map[string]interface{}{},
			skip: false,
		},
		{
			name: "happy - finds resources given status",
			ri:   resourceIdentifier{gvr: resource1.gvr, status: map[string]interface{}{"foo": "bar"}},
			want: []map[string]interface{}{response1.Object, response2.Object},
			skip: false,
		},
		{
			name: "happy - finds resources given nested status",
			ri:   resourceIdentifier{gvr: resource1.gvr, status: map[string]interface{}{"status": map[string]interface{}{"baz": map[string]interface{}{"deep": "nest"}}}},
			want: []map[string]interface{}{response2.Object},
			skip: false,
		},
		{
			name: "happy - finds correct resources given multiple resources with desired key-value pair",
			ri:   resourceIdentifier{gvr: resource1.gvr, status: map[string]interface{}{"tomato": "potato"}},
			want: []map[string]interface{}{response1.Object},
			skip: false,
		},
		{
			name: "happy - finds correct resources given multiple resources with desired key-value pair - nested",
			ri:   resourceIdentifier{gvr: resource1.gvr, status: map[string]interface{}{"status": map[string]interface{}{"tomato": "potato"}}},
			want: []map[string]interface{}{response2.Object},
			skip: false,
		},
		{
			name: "sad - finds resources given status",
			ri:   resourceIdentifier{gvr: resource1.gvr, status: map[string]interface{}{"status": map[string]interface{}{"baz": "fail"}}},
			want: []map[string]interface{}{response3.Object},
			skip: false,
		},
	}

	for _, tc := range listTests {
		t.Run(tc.name, func(t *testing.T) {
			if tc.skip {
				t.Skip()
			}
			got, _ := ListResources(client, tc.ri)
			if len(tc.want) != len(got) {
				t.Errorf("Expected %d items but got %d", len(tc.want), len(got))
			}
			for _, wantItem := range tc.want {
				if !equalityCheck(wantItem, got) {
					t.Errorf("did not find\n%s\nin\n%v\n", wantItem, got)
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
