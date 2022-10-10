package kln

import (
	"testing"
	"time"
)

func TestListResources(t *testing.T) {
	resource1 := CreateResource(customGVR, map[string]interface{}{"namespace": "ns", "name": "name1"}, status1)
	resource2 := CreateResource(customGVR, map[string]interface{}{"namespace": "ns", "name": "name2"}, status2)
	resource3 := CreateResource(customGVR, map[string]interface{}{"namespace": "ns3", "name": "name3"}, status3)
	client := SetupFakeDynamicClient(t, resource1, resourceFake)
	manifest1 := NewUnstructured(t, resource1, time.Now().Add(-10*time.Minute).Format(RFC3339))
	manifest2 := NewUnstructured(t, resource2, time.Now().Add(-40*time.Minute).Format(RFC3339))
	manifest3 := NewUnstructured(t, resource3, time.Now().Add(-70*time.Minute).Format(RFC3339))
	response1 := ApplyResource(t, client, resource1, manifest1)
	response2 := ApplyResource(t, client, resource2, manifest2)
	response3 := ApplyResource(t, client, resource3, manifest3)

	listTests := []listTestCases{
		{
			name: "happy - finds resources given only gvr",
			ri:   ResourceIdentifier{GVR: resource1.GVR},
			want: []map[string]interface{}{response1.Object, response2.Object, response3.Object},
			skip: false,
		},

		{
			name: "sad - finds resources given only gvr",
			ri:   ResourceIdentifier{GVR: fakeGVR},
			want: []map[string]interface{}{},
			skip: false,
		},

		{
			name: "happy - finds resources given minAge",
			ri:   ResourceIdentifier{GVR: resource1.GVR, MinAge: 0.5},
			want: []map[string]interface{}{response2.Object, response3.Object},
			skip: false,
		},

		{
			name: "sad - finds resources given minAge",
			ri:   ResourceIdentifier{GVR: resource1.GVR, MinAge: 1.5},
			want: []map[string]interface{}{},
			skip: false,
		},

		{
			name: "happy - finds resources given metadata",
			ri:   ResourceIdentifier{GVR: resource1.GVR, Metadata: map[string]interface{}{"namespace": "ns"}},
			want: []map[string]interface{}{response1.Object, response2.Object},
			skip: false,
		},

		{
			name: "sad - finds resources given metadata",
			ri:   ResourceIdentifier{GVR: resource1.GVR, Metadata: map[string]interface{}{"namespace": "ns", "name": "fake"}},
			want: []map[string]interface{}{},
			skip: false,
		},

		{
			name: "happy - finds resources given minAge and metadata",
			ri:   ResourceIdentifier{GVR: resource1.GVR, Metadata: map[string]interface{}{"namespace": "ns"}, MinAge: 0.5},
			want: []map[string]interface{}{response2.Object},
			skip: false,
		},

		{
			name: "sad - finds resources given minAge and metadata",
			ri:   ResourceIdentifier{GVR: resource1.GVR, Metadata: map[string]interface{}{"name": "fake"}, MinAge: 0.5},
			want: []map[string]interface{}{},
			skip: false,
		},

		{
			name: "happy - finds resources given status",
			ri:   ResourceIdentifier{GVR: resource1.GVR, Status: map[string]interface{}{"foo": "bar"}},
			want: []map[string]interface{}{response1.Object, response2.Object},
			skip: false,
		},

		{
			name: "happy - finds resources given nested status",
			ri:   ResourceIdentifier{GVR: resource1.GVR, Status: map[string]interface{}{"status": map[string]interface{}{"baz": map[string]interface{}{"deep": "nest"}}}},
			want: []map[string]interface{}{response2.Object},
			skip: false,
		},

		{
			name: "happy - finds correct resources given multiple resources with desired key-value pair",
			ri:   ResourceIdentifier{GVR: resource1.GVR, Status: map[string]interface{}{"tomato": "potato"}},
			want: []map[string]interface{}{response1.Object},
			skip: false,
		},

		{
			name: "happy - finds correct resources given multiple resources with desired key-value pair - nested",
			ri:   ResourceIdentifier{GVR: resource1.GVR, Status: map[string]interface{}{"status": map[string]interface{}{"tomato": "potato"}}},
			want: []map[string]interface{}{response2.Object},
			skip: false,
		},

		{
			name: "sad - finds resources given status",
			ri:   ResourceIdentifier{GVR: resource1.GVR, Status: map[string]interface{}{"status": map[string]interface{}{"baz": "fail"}}},
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
