package kln

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"

	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/tools/clientcmd"
)

var InfoLog = log.New(os.Stdout, "INFO: ", log.Ldate|log.Ltime|log.Lshortfile)
var WarningLog = log.New(os.Stdout, "WARNING: ", log.Ldate|log.Ltime|log.Lshortfile)
var ErrorLog = log.New(os.Stdout, "ERROR: ", log.Ldate|log.Ltime|log.Lshortfile)

const (
	RFC3339 = "2006-01-02T15:04:05Z07:00"
)

type ResourceIdentifier struct {
	GVR         schema.GroupVersionResource `yaml:"gvr"`
	MinAge      float64                     `yaml:"minAge"`
	ApiVersion  string                      `yaml:"apiVersion"`
	Kind        string                      `yaml:"kind"`
	Metadata    map[string]interface{}      `yaml:"metadata"`
	Spec        map[string]interface{}      `yaml:"spec"`
	Status      map[string]interface{}      `yaml:"status"`
	Name        string                      `yaml:"name"`
	Description string                      `yaml:"description"`
}

func MapIntersection(mapA, mapB map[string]interface{}) bool {
	for k, vA := range mapA {
		if vB, ok := mapB[k]; ok && Typeof(vA) == Typeof(vB) {
			if Typeof(vA) == "map[string]interface {}" {
				return MapIntersection(vA.(map[string]interface{}), vB.(map[string]interface{}))
			}
			if Typeof(vA) == "[]interface {}" {
				return ArrayIntersection(vA.([]interface{}), vB.([]interface{}))
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

func ArrayIntersection(sliceSmall, sliceBig []interface{}) bool {
	if len(sliceSmall) > len(sliceBig) {
		return false
	}
	for _, vSmall := range sliceSmall {
		if !Contains(sliceBig, vSmall) {
			return false
		}
	}
	return true
}

func Contains(sliceBig []interface{}, vSmall interface{}) bool {
	for _, vBig := range sliceBig {
		if Typeof(vBig) != Typeof(vSmall) {
			return false
		}
		if Typeof(vBig) == "map[string]interface {}" {
			return MapIntersection(vSmall.(map[string]interface{}), vBig.(map[string]interface{}))
		}
		if Typeof(vBig) == "[]interface {}" {
			return ArrayIntersection(vSmall.([]interface{}), vBig.([]interface{}))
		}
		if vSmall != vBig {
			return false
		}
	}
	return true
}

func Typeof(v interface{}) string {
	return fmt.Sprintf("%T", v)
}

func GetDynamicClient(kubeconfig string) dynamic.Interface {
	config, err := clientcmd.BuildConfigFromFlags("", kubeconfig)
	if err != nil {
		panic(err)
	}
	client, err := dynamic.NewForConfig(config)
	if err != nil {
		panic(err)
	}
	return client
}

func ReadFile(file string) []byte {
	filename, _ := filepath.Abs(file)
	config, err := ioutil.ReadFile(filename)
	if err != nil {
		panic(err)
	}
	return config
}
