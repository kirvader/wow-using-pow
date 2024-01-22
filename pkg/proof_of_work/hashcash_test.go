package proof_of_work

import (
	"testing"
	"time"
)

func TestHashcashDataToString(t *testing.T) {
	date := time.Unix(17000000012, 0)

	t.Run("unreasonable data", func(t *testing.T) {
		got := Hashcash{
			Version:    1,
			ZerosCount: 32,
			Date:       &date,
			Resource:   "asd",
			Extension:  "321",
			Rand:       "123",
			Counter:    54,
		}.Encode()
		want := "1:32:080916081332:asd:321:123:54"
		if got != want {
			t.Errorf("HashcashData.ToString() test failed. Expected: %s, got %s", want, got)
		}
	})

}
