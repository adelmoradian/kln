package kln

import (
	"context"

	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/dynamic"
)

func FlagForDeletion(client dynamic.Interface, ri ResourceIdentifier, cleanSwitch bool) error {
	resources, err := ListResources(client, ri)
	if err != nil {
		return err
	}

	if len(resources) == 0 {
		InfoLog.Printf("did not find any resources that match the following resource identifier\n%v", ri)
		return nil
	}

	for _, resource := range resources {
		var patch []byte
		ns := resource.GetNamespace()
		name := resource.GetName()

		if cleanSwitch {
			patch = []byte(`{"metadata":{"labels":{"kln.com/delete":"true"}}}`)
		} else {
			patch = []byte(`{"metadata":{"labels":{"kln.com/delete":"false"}}}`)
		}
		_, err := client.Resource(ri.GVR).Namespace(ns).Patch(context.TODO(), name, types.MergePatchType, patch, v1.PatchOptions{})
		if err != nil {
			return err
		}
	}
	return nil
}
