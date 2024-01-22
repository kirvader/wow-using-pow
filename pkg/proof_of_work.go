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
func ComputeHashcash(header *HashcashHeader, maxCounterValue int32) (*HashcashHeader, error) {
	for header.Counter <= maxCounterValue {
		strHeader := header.ToString()
		hash, err := countSHA1(strHeader)
		if err != nil {
			return nil, err
		}
		if IsHashCorrect(hash, header.ZerosCount) {
			return header, nil
		}
		header.Counter++
	}
	return nil, fmt.Errorf("max iterations exceeded")
}