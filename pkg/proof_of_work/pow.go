package proof_of_work

type POWPuzzle interface {
	Solve(int32) error
	Verify() (bool, error)

	ToJSON() (string, error)
	FromJSON(string) error
}
