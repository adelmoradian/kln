package list

import (
	"context"
	"errors"
	"fmt"

	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
)

func ListResources(client dynamic.Interface, gvr schema.GroupVersionResource) (*unstructured.UnstructuredList, error) {
	response, _ := client.Resource(gvr).List(context.TODO(), v1.ListOptions{})
	if len(response.Items) == 0 {
		return nil, errors.New(fmt.Sprintf("Did not find any objects for the following GVR\n  Group: %s\n  Version: %s\n  resource: %s\n", gvr.Resource, gvr.Resource, gvr.Group))
	}
	return response, nil
}
