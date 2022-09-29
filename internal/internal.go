package internal

import (
	"context"
	"fmt"
	"time"

	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
)

const (
	RFC3339 = "2006-01-02T15:04:05Z07:00"
)

type resourceIdentifier struct {
	gvr        schema.GroupVersionResource
	age        float64
	apiVersion string
	kind       string
	metadata   map[string]interface{}
	spec       map[string]interface{}
	status     map[string]interface{}
}

func (ri *resourceIdentifier) ListResources(client dynamic.Interface) []unstructured.Unstructured {
	var responseList []unstructured.Unstructured
	responseFromServer, err := client.Resource(ri.gvr).List(context.TODO(), v1.ListOptions{})
	if err != nil {
		return responseList
	}
	responseList, err = filterByAge(responseFromServer, ri.age)
	responseList = filterByMetadata(responseList, ri.metadata)
	responseList = filterByStatus(responseList, ri.status)
	return responseList
}

// func FlagResources(client dynamic.Interface) {
// }

func filterByAge(responseFromServer *unstructured.UnstructuredList, minAgeFilter float64) ([]unstructured.Unstructured, error) {
	var responseList []unstructured.Unstructured
	if minAgeFilter == 0 {
		for _, item := range responseFromServer.Items {
			responseList = append(responseList, item)
		}
		return responseList, nil
	}
	for _, item := range responseFromServer.Items {
		creationTimestamp, err := time.Parse(RFC3339, item.Object["metadata"].(map[string]interface{})["creationTimestamp"].(string))
		if err != nil {
			return nil, err
		}
		age := time.Since(creationTimestamp)
		if age.Hours() > minAgeFilter {
			responseList = append(responseList, item)
		}
	}
	return responseList, nil
}

func filterByMetadata(responseFromServer []unstructured.Unstructured, metadataFilter map[string]interface{}) []unstructured.Unstructured {
	var responseList []unstructured.Unstructured
	if metadataFilter == nil {
		return responseFromServer
	}
	for _, item := range responseFromServer {
		objectMeta := item.Object["metadata"].(map[string]interface{})
		if mapIntersectionCheck(metadataFilter, objectMeta) {
			responseList = append(responseList, item)
		}
	}
	return responseList
}

func filterByStatus(responseFromServer []unstructured.Unstructured, statusFilter map[string]interface{}) []unstructured.Unstructured {
	var responseList []unstructured.Unstructured
	if statusFilter == nil {
		return responseFromServer
	}
	for _, item := range responseFromServer {
		objectMeta := item.Object["status"].(map[string]interface{})
		if mapIntersectionCheck(statusFilter, objectMeta) {
			responseList = append(responseList, item)
		}
	}
	return responseList
}

func mapIntersectionCheck(mapA, mapB map[string]interface{}) bool {
	for k, vA := range mapA {
		if vB, ok := mapB[k]; ok && typeof(vA) == typeof(vB) {
			if typeof(vA) == "map[string]interface {}" {
				return mapIntersectionCheck(vA.(map[string]interface{}), vB.(map[string]interface{}))
			}
			if vA != vB {
				return false
			}
		} else {
			return false
		}
	}
	return true
}

func typeof(v interface{}) string {
	return fmt.Sprintf("%T", v)
}
