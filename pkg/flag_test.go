package kln

import (
	"context"
	"testing"
	"time"

	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"k8s.io/apimachinery/pkg/types"
)

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
