package pkg

import (
	"crypto/sha1"
	"fmt"
	"time"
)

type HashcashHeader struct {
	Version    int32
	ZerosCount int32
	Date       *time.Time
	Resource   string
	Extension  string // in general it is ignored and here it is added as a stub for forward compatibility
	Rand       string
	Counter    int32
}

// ToString - stringifies hashcash for next sending it on TCP
func (h HashcashHeader) ToString() string {
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

func countSHA1(data string) ([]byte, error) {
	hasher := sha1.New()
	_, err := hasher.Write([]byte(data))
	if err != nil {
		return nil, err
	}
	return hasher.Sum(nil), nil
}

func IsHashCorrect(hash []byte, zerosCount int32) bool {
	if zerosCount > int32(len(hash)) {
		return false
	}
	for _, ch := range hash[:zerosCount] {
		if ch != 0 {
			return false
		}
	}
	return true
}

// bruteforce until IsHashCorrect is true
func (h *HashcashHeader) ComputeHashcash(maxCounterValue int32) (*HashcashHeader, error) {
	result := *h
	for result.Counter <= maxCounterValue {
		header := result.ToString()
		hash, err := countSHA1(header)
		if err != nil {
			return nil, err
		}
		//fmt.Println(header, hash)
		if IsHashCorrect(hash, result.ZerosCount) {
			return &result, nil
		}
		// if hash don't have needed count of leading zeros, we are increasing counter and try next hash
		result.Counter++
	}
	return nil, fmt.Errorf("max iterations exceeded")
}
