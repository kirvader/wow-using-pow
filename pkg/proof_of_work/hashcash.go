package proof_of_work

import (
	"crypto/sha1"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/kirvader/wow-using-pow/internal/utils"
)

type HashcashBuilder struct {
	ZerosCount        int32
	ChallengeDuration *time.Duration
}

func (builder *HashcashBuilder) GenerateRandomChallenge(currentTime *time.Time, resource string) (POWChallenge, string) {
	randomId := utils.GenerateRandomString()
	return &Hashcash{
		Version:    1,
		ZerosCount: builder.ZerosCount,
		Date:       currentTime,
		Resource:   resource,
		Rand:       randomId,
		Counter:    0,
	}, randomId
}

func (builder *HashcashBuilder) GenerateChallengeById(currentTime *time.Time, resource, randomId string) POWChallenge {
	return &Hashcash{
		Version:    1,
		ZerosCount: builder.ZerosCount,
		Date:       currentTime,
		Resource:   resource,
		Rand:       randomId,
		Counter:    0,
	}
}

func (builder *HashcashBuilder) GetChallengeDuration() *time.Duration {
	return builder.ChallengeDuration
}

var _ POWChallenge = &Hashcash{}

type Hashcash struct {
	Version    int32
	ZerosCount int32
	Date       *time.Time
	Resource   string
	Extension  string // in general it is ignored and here it is added as a stub for forward compatibility
	Rand       string
	Counter    int32
}

func (h *Hashcash) GetDate() *time.Time {
	return h.Date
}

func (h *Hashcash) GetRand() string {
	return h.Rand
}

func (h *Hashcash) GetResourse() string {
	return h.Resource
}

func (h *Hashcash) encode() string {
	if h == nil {
		return ""
	}

	stringDate := fmt.Sprintf(
		"%02d%02d%02d%02d%02d%02d",
		h.Date.Year()%100,
		h.Date.Month(),
		h.Date.Day(),
		h.Date.Hour(),
		h.Date.Minute(),
		h.Date.Second(),
	)

	return fmt.Sprintf("%d:%d:%s:%s:%s:%s:%d", h.Version, h.ZerosCount, stringDate, h.Resource, h.Extension, h.Rand, h.Counter)
}

func countSHA1(data string) (string, error) {
	hasher := sha1.New()
	_, err := hasher.Write([]byte(data))
	if err != nil {
		return "", err
	}
	return string(hasher.Sum(nil)), nil
}

func isHashCorrect(hash string, zerosCount int32) bool {
	if zerosCount > int32(len(hash)) {
		return false
	}
	for _, ch := range hash[:zerosCount] {
		if ch != 0x30 {
			return false
		}
	}
	return true
}

func (header *Hashcash) Solve(maxIterationsAmount int32) error {
	for header.Counter <= maxIterationsAmount {
		strHeader := header.encode()
		hash, err := countSHA1(strHeader)
		if err != nil {
			return err
		}
		if isHashCorrect(hash, header.ZerosCount) {
			return nil
		}
		header.Counter++
	}
	return errors.New("max iterations exceeded")
}

func (header *Hashcash) Verify() (bool, error) {
	hash, err := countSHA1(header.encode())
	if err != nil {
		return false, fmt.Errorf("computing sha failed: %v", err)
	}
	return isHashCorrect(hash, header.ZerosCount), nil
}

func (header *Hashcash) ToJSON() (string, error) {
	headerBytes, err := json.Marshal(header)
	if err != nil {
		return "", fmt.Errorf("hashcash marshalling failed: %w", err)
	}
	return string(headerBytes), nil
}

func (header *Hashcash) FromJSON(jsonString string) error {
	return json.Unmarshal([]byte(jsonString), &header)
}
