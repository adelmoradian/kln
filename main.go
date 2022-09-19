package main

import "k8s.io/apimachinery/pkg/runtime/schema"

func GetResources(client *dynamicfake.FakeDynamicClient, FakeGVR map[string]schema.GroupVersionResource) {
	panic("unimplemented")
}
