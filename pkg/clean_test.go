package kln

import (
	"context"
	"testing"
	"time"

	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
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
	ri := ResourceIdentifier{GVR: resource1.GVR}
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
