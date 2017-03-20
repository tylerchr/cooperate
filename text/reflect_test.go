package text

import (
	"reflect"
	"testing"

	"github.com/tylerchr/cooperate"
)

func BenchmarkType_Reflect(b *testing.B) {

	var action cooperate.Action = RetainAction(0)
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_ = reflect.TypeOf(action)
	}

}

func BenchmarkType_Assert(b *testing.B) {
	var action cooperate.Action = RetainAction(0)
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_, _ = action.(RetainAction)
	}
}

func BenchmarkType_Switch(b *testing.B) {
	var action cooperate.Action = RetainAction(0)
	for i := 0; i < b.N; i++ {
		switch action.(type) {
		case RetainAction:
		default:
		}
	}
}
