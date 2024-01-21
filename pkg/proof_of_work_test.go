package pkg

import (
	"testing"
	"time"
)

func TestHashcashDataToString(t *testing.T) {
	t.Run("unreasonable data", func(t *testing.T) {
		got := HashcashHeader{
			Version:    1,
			ZerosCount: 32,
			Date:       PointTo(time.Unix(17000000012, 0)),
			Resource:   "asd",
			Extension:  "321",
			Rand:       "123",
			Counter:    54,
		}.ToString()
		want := "1:32:080916081332:asd:321:123:54"
		if got != want {
			t.Errorf("HashcashData.ToString() test failed. Expected: %s, got %s", want, got)
		}
	})

}
