package proof_of_work

import (
	"crypto/sha1"
	"encoding/json"
	"errors"
	"fmt"
	"time"
)

var _ POWPuzzle = &Hashcash{}

type Hashcash struct {
	Version    int32
	ZerosCount int32
	Date       *time.Time
	Resource   string
	Extension  string // in general it is ignored and here it is added as a stub for forward compatibility
	Rand       string
	Counter    int32
}

func (h Hashcash) Encode() string {
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
		strHeader := header.Encode()
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
	hash, err := countSHA1(header.Encode())
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
