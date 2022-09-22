package list

import (
	"reflect"
	"testing"

	utility "github.com/adelmoradian/kln/utility"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

func TestListResources(t *testing.T) {
	client := utility.SetupFakeDynamicClient(t, utility.RealGVRs, utility.FakeGVRs)
	responseMap := utility.ApplyResources(t, client, utility.RealGVRs)

	t.Run("lists existing resources associated with a gvr", func(t *testing.T) {
		want := responseMap
		gotMap := make(map[string]unstructured.Unstructured)
		for _, gvr := range utility.RealGVRs {
			got, err := ListResources(client, gvr)
			if err != nil {
				t.Errorf("got error %s\n", err.Error())
			}
			for _, item := range got.Items {
				if !reflect.DeepEqual(item, want[item.Object["kind"].(string)]) {
					t.Errorf("got %v\n\n want %v\n", item, want[item.Object["kind"].(string)])
				}
				gotMap[got.GetKind()] = item
			}
		}

		for _, item := range gotMap {
			a := item.Object["kind"].(string)
			if !reflect.DeepEqual(item, want[a]) {
				t.Errorf("got %v\n\n\n want %v", item, want[a])
			}
		}
	})

	t.Run("throws error if no resources found for a given gvr", func(t *testing.T) {
		for _, gvr := range utility.FakeGVRs {
			got, err := ListResources(client, gvr)
			if err == nil {
				t.Error("Should get error but did not")
			}
			if got != nil {
				t.Errorf("Should get nil but instead got %s", got.Items)
			}
		}
	})
}
