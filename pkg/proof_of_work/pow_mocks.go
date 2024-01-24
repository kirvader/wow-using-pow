package proof_of_work

import (
	"time"
)

type POWChallengeBuilderMock struct {
	GenerateRandomChallengeFunc func(currentTime *time.Time, resource string) (POWChallenge, string)
	GenerateChallengeByIdFunc   func(currentTime *time.Time, resource, randomId string) POWChallenge

	GetChallengeDurationFunc func() *time.Duration
}

func (bm *POWChallengeBuilderMock) GenerateRandomChallenge(currentTime *time.Time, resource string) (POWChallenge, string) {
	return bm.GenerateRandomChallengeFunc(currentTime, resource)
}

func (bm *POWChallengeBuilderMock) GenerateChallengeById(currentTime *time.Time, resource, randomId string) POWChallenge {
	return bm.GenerateChallengeByIdFunc(currentTime, resource, randomId)

}

func (bm *POWChallengeBuilderMock) GetChallengeDuration() *time.Duration {
	return bm.GetChallengeDurationFunc()
}

type POWChallengeMock struct {
	SolveFunc  func(int32) error
	VerifyFunc func() (bool, error)

	Resource string
	RandVal  string
	Date     *time.Time
}

func (c *POWChallengeMock) GetDate() *time.Time {
	return c.Date
}

func (c *POWChallengeMock) GetRand() string {
	return c.RandVal
}

func (c *POWChallengeMock) GetResourse() string {
	return c.Resource
}
func (c *POWChallengeMock) Solve(maxIterationsAmount int32) error {
	return c.SolveFunc(maxIterationsAmount)
}

func (c *POWChallengeMock) Verify() (bool, error) {
	return c.VerifyFunc()
}

// Json (un)marshalling is already covered with tests, so these are just the stubs
func (c *POWChallengeMock) ToJSON() (string, error) {
	return "", nil
}

func (c *POWChallengeMock) FromJSON(jsonString string) error {
	return nil
}
