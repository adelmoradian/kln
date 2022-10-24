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
	dynamicfake "k8s.io/client-go/dynamic/fake"
)

type listTestCases struct {
	name string
	ri   ResourceIdentifier
	want []map[string]interface{}
	skip bool
}

type GVRK struct {
	GVR  schema.GroupVersionResource
	Kind string
}

var (
	aGVRK = GVRK{
		GVR: schema.GroupVersionResource{
			Group:    "agroup",
			Version:  "aversion",
			Resource: "akinds",
		},
		Kind: "AKind",
	}

	fakeGVRK = GVRK{
		GVR: schema.GroupVersionResource{
			Group:    "fake",
			Version:  "fakeVersion",
			Resource: "fakes",
		},
		Kind: "Fake",
	}

	r1 = &unstructured.Unstructured{
		Object: map[string]interface{}{
			"apiVersion": aGVRK.GVR.Version,
			"kind":       aGVRK.Kind,
			"metadata": map[string]interface{}{
				"creationTimestamp": time.Now().Add(-10 * time.Minute).Format(RFC3339),
				"namespace":         "ns",
				"name":              "name1",
			},
			"status": map[string]interface{}{
				"foo":    "bar",
				"tomato": "potato",
			},
		},
	}

	r2 = &unstructured.Unstructured{
		Object: map[string]interface{}{
			"apiVersion": aGVRK.GVR.Version,
			"kind":       aGVRK.Kind,
			"metadata": map[string]interface{}{
				"creationTimestamp": time.Now().Add(-40 * time.Minute).Format(RFC3339),
				"namespace":         "ns",
				"name":              "name2",
			},
			"status": map[string]interface{}{
				"foo": "bar",
				"status": map[string]interface{}{
					"baz":    map[string]interface{}{"deep": "nest"},
					"tomato": "potato",
				},
			},
		},
	}

	r3 = &unstructured.Unstructured{
		Object: map[string]interface{}{
			"apiVersion": aGVRK.GVR.Version,
			"kind":       aGVRK.Kind,
			"metadata": map[string]interface{}{
				"creationTimestamp": time.Now().Add(-70 * time.Minute).Format(RFC3339),
				"namespace":         "ns3",
				"name":              "name3",
			},
			"status": map[string]interface{}{
				"foo": "notBar",
				"status": map[string]interface{}{
					"baz": "fail",
				},
			},
		},
	}
)

func TestListResources(t *testing.T) {
	scheme := runtime.NewScheme()
	scheme.AddKnownTypeWithName(schema.GroupVersionKind{Group: aGVRK.GVR.Group, Version: aGVRK.GVR.Version, Kind: aGVRK.Kind + "List"}, &unstructured.Unstructured{})
	scheme.AddKnownTypeWithName(schema.GroupVersionKind{Group: fakeGVRK.GVR.Group, Version: fakeGVRK.GVR.Version, Kind: fakeGVRK.Kind + "List"}, &unstructured.Unstructured{})
	client := dynamicfake.NewSimpleDynamicClient(scheme)
	response1, err := client.Resource(aGVRK.GVR).Namespace("ns").Create(context.TODO(), r1, v1.CreateOptions{})
	if err != nil {
		t.Error(err)
	}
	response2, err := client.Resource(aGVRK.GVR).Namespace("ns").Create(context.TODO(), r2, v1.CreateOptions{})
	if err != nil {
		t.Error(err)
	}
	response3, err := client.Resource(aGVRK.GVR).Namespace("ns3").Create(context.TODO(), r3, v1.CreateOptions{})
	if err != nil {
		t.Error(err)
	}

	listTests := []listTestCases{
		{
			name: "happy - finds resources given only gvr",
			ri:   ResourceIdentifier{GVR: aGVRK.GVR},
			want: []map[string]interface{}{response1.Object, response2.Object, response3.Object},
			skip: false,
		},

		{
			name: "sad - finds resources given only gvr",
			ri:   ResourceIdentifier{GVR: fakeGVRK.GVR},
			want: []map[string]interface{}{},
			skip: false,
		},

		{
			name: "happy - finds resources given minAge",
			ri:   ResourceIdentifier{GVR: aGVRK.GVR, MinAge: 0.5},
			want: []map[string]interface{}{response2.Object, response3.Object},
			skip: false,
		},

		{
			name: "sad - finds resources given minAge",
			ri:   ResourceIdentifier{GVR: aGVRK.GVR, MinAge: 1.5},
			want: []map[string]interface{}{},
			skip: false,
		},

		{
			name: "happy - finds resources given metadata",
			ri:   ResourceIdentifier{GVR: aGVRK.GVR, Metadata: map[string]interface{}{"namespace": "ns"}},
			want: []map[string]interface{}{response1.Object, response2.Object},
			skip: false,
		},

		{
			name: "sad - finds resources given metadata",
			ri:   ResourceIdentifier{GVR: aGVRK.GVR, Metadata: map[string]interface{}{"namespace": "ns", "name": "fake"}},
			want: []map[string]interface{}{},
			skip: false,
		},

		{
			name: "happy - finds resources given minAge and metadata",
			ri:   ResourceIdentifier{GVR: aGVRK.GVR, Metadata: map[string]interface{}{"namespace": "ns"}, MinAge: 0.5},
			want: []map[string]interface{}{response2.Object},
			skip: false,
		},

		{
			name: "sad - finds resources given minAge and metadata",
			ri:   ResourceIdentifier{GVR: aGVRK.GVR, Metadata: map[string]interface{}{"name": "fake"}, MinAge: 0.5},
			want: []map[string]interface{}{},
			skip: false,
		},

		{
			name: "happy - finds resources given status",
			ri:   ResourceIdentifier{GVR: aGVRK.GVR, Status: map[string]interface{}{"foo": "bar"}},
			want: []map[string]interface{}{response1.Object, response2.Object},
			skip: false,
		},

		{
			name: "happy - finds resources given nested status",
			ri:   ResourceIdentifier{GVR: aGVRK.GVR, Status: map[string]interface{}{"status": map[string]interface{}{"baz": map[string]interface{}{"deep": "nest"}}}},
			want: []map[string]interface{}{response2.Object},
			skip: false,
		},

		{
			name: "happy - finds correct resources given multiple resources with desired key-value pair",
			ri:   ResourceIdentifier{GVR: aGVRK.GVR, Status: map[string]interface{}{"tomato": "potato"}},
			want: []map[string]interface{}{response1.Object},
			skip: false,
		},

		{
			name: "happy - finds correct resources given multiple resources with desired key-value pair - nested",
			ri:   ResourceIdentifier{GVR: aGVRK.GVR, Status: map[string]interface{}{"status": map[string]interface{}{"tomato": "potato"}}},
			want: []map[string]interface{}{response2.Object},
			skip: false,
		},

		{
			name: "sad - finds resources given status",
			ri:   ResourceIdentifier{GVR: aGVRK.GVR, Status: map[string]interface{}{"status": map[string]interface{}{"baz": "fail"}}},
			want: []map[string]interface{}{response3.Object},
			skip: false,
		},
	}

	for _, tc := range listTests {
		t.Run(tc.name, func(t *testing.T) {
			if tc.skip {
				t.Skip()
			}
			got := ListResources(client, tc.ri)
			if len(tc.want) != len(got) {
				t.Errorf("Expected %d items but got %d", len(tc.want), len(got))
			}
			for _, wantItem := range tc.want {
				if !equalityCheck(wantItem, got) {
					t.Errorf("want --->\n%s\nbut did not find it in --->\n%v\n", wantItem, got)
				}
			}
		})
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
