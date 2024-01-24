package proof_of_work

import "time"

// now all POW algos based on rand string can easily extend this project

type POWChallengeBuilder interface {
	GenerateRandomChallenge(currentTime *time.Time, resource string) (POWChallenge, string)
	GenerateChallengeById(currentTime *time.Time, resource, randomId string) POWChallenge

	GetChallengeDuration() *time.Duration
}

type POWChallenge interface {
	Solve(int32) error
	Verify() (bool, error)

	// to check client's id
	GetResourse() string
	// to check client's token if applicable
	GetRand() string
	// to check if challenge expired
	GetDate() *time.Time

	ToJSON() (string, error)
	FromJSON(string) error
}
