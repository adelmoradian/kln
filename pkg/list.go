package kln

import (
	"context"
	"time"

	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/client-go/dynamic"
)

func ListResources(client dynamic.Interface, ri ResourceIdentifier) []unstructured.Unstructured {
	var responseList []unstructured.Unstructured
	// InfoLog.Printf("--- GVR: %s, Name: %s, Description: %s\n", ri.GVR, ri.Name, ri.Description)
	responseFromServer, err := client.Resource(ri.GVR).List(context.TODO(), v1.ListOptions{})
	if err != nil {
		return responseList
	}
	responseList, err = filterByAge(responseFromServer, ri.MinAge)
	responseList = filterByMetadata(responseList, ri.Metadata)
	responseList = filterByStatus(responseList, ri.Status)
	return responseList
}

func filterByAge(responseFromServer *unstructured.UnstructuredList, minAgeFilter float64) ([]unstructured.Unstructured, error) {
	var responseList []unstructured.Unstructured
	if minAgeFilter == 0 {
		for _, item := range responseFromServer.Items {
			responseList = append(responseList, item)
		}
		return responseList, nil
	}
	for _, item := range responseFromServer.Items {
		age := time.Since(item.GetCreationTimestamp().Time)
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
		if mapIntersection(metadataFilter, objectMeta) {
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
		objectStatus := item.Object["status"].(map[string]interface{})
		if mapIntersection(statusFilter, objectStatus) {
			responseList = append(responseList, item)
		}
	}
	return responseList
}
