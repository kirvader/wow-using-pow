package proof_of_work

import (
	"fmt"
	"math/rand"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/kirvader/wow-using-pow/internal/utils"
)

func TestHashcash_Encode(t *testing.T) {
	date := time.Unix(17000000012, 0)

	testCases := []struct {
		name  string
		input *Hashcash
		want  string
	}{
		{
			name:  "empty data",
			input: nil,
			want:  "",
		},
		{
			name: "unreasonable data",
			input: &Hashcash{
				Version:    123,
				ZerosCount: 152,
				Date:       &date,
				Resource:   "unavailable",
				Extension:  "some",
				Rand:       "abracadabra",
				Counter:    -239,
			},
			want: "123:152:080916:unavailable:some:abracadabra:-239",
		},
		{
			name: "common data",
			input: &Hashcash{
				Version:    1,
				ZerosCount: 3,
				Date:       &date,
				Resource:   "kirill.kondratiuk",
				Extension:  "",
				Rand:       "AnAJT=/34",
				Counter:    1512497,
			},
			want: "1:3:080916:kirill.kondratiuk::AnAJT=/34:1512497",
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			got := testCase.input.encode()
			if diff := cmp.Diff(got, testCase.want); diff != "" {
				t.Errorf("Hashcash.Encode failed. Diff: %s", diff)
			}
		})
	}
}

func TestHashcashProofOfWorkWorkflow(t *testing.T) {
	date := time.Unix(16000000012, 0)
	var maxIterations int32 = 20000000
	resource := "kirill.kondratiuk"
	rand.Seed(239)

	for index := 0; index < 100; index++ {
		t.Run(fmt.Sprintf("test #%d", index), func(t *testing.T) {
			current := &Hashcash{
				Version:    1,
				ZerosCount: 2,
				Date:       &date,
				Resource:   resource,
				Extension:  "",
				Rand:       utils.GenerateRandomString(),
				Counter:    0,
			}
			err := current.Solve(maxIterations)
			if err != nil {
				t.Fatalf("hashcash couldn't be solved: %v", err)
			}

			isValid, err := current.Verify()
			if err != nil {
				t.Fatalf("hashcash couldn't be verified: %v", err)
			}
			if !isValid {
				t.Fatalf("hashcash is invalid: %v", err)
			}
		})
	}
}
