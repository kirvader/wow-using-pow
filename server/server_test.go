package server

import (
	"bufio"
	"bytes"
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/kirvader/wow-using-pow/internal/protocol"
	"github.com/kirvader/wow-using-pow/pkg/clock"
	pow "github.com/kirvader/wow-using-pow/pkg/proof_of_work"
	"github.com/kirvader/wow-using-pow/pkg/storage"
)

func TestServer_sendChallengeToClient(t *testing.T) {
	defaultTestDuration := time.Minute

	type ServiceSetup struct {
		storageSet       storage.StorageSet
		clock            clock.Clock
		challengeBuilder pow.POWChallengeBuilder
	}
	testCases := []struct {
		name            string
		serviceSetup    *ServiceSetup
		clientId        string
		wantSentMessage *protocol.Message
		wantErr         error
	}{
		{
			name: "storage set insert failed",
			serviceSetup: &ServiceSetup{
				storageSet: &storage.StorageSetMock{
					InsertClientTokenFunc: func(_ context.Context, _, _ string, _ time.Duration) error {
						return errors.New("storage set internal error")
					},
				},
				clock: &clock.ClockMock{
					Counter: 0,
				},
				challengeBuilder: &pow.POWChallengeBuilderMock{
					GenerateRandomChallengeFunc: func(currentTime *time.Time, resource string) (pow.POWChallenge, string) {
						return &pow.Hashcash{}, "random value"
					},
					GetChallengeDurationFunc: func() *time.Duration {
						return &defaultTestDuration
					},
				},
			},
			clientId:        "client",
			wantSentMessage: nil,
			wantErr:         fmt.Errorf("inserting into storage set failed: %w", errors.New("storage set internal error")),
		},
		{
			name: "success",
			serviceSetup: &ServiceSetup{
				storageSet: &storage.StorageSetMock{
					InsertClientTokenFunc: func(_ context.Context, _, _ string, _ time.Duration) error {
						return nil
					},
				},
				clock: &clock.ClockMock{
					Counter: 0,
				},
				challengeBuilder: &pow.POWChallengeBuilderMock{
					GenerateRandomChallengeFunc: func(currentTime *time.Time, resource string) (pow.POWChallenge, string) {
						return &pow.Hashcash{}, "random value"
					},
					GetChallengeDurationFunc: func() *time.Duration {
						return &defaultTestDuration
					},
				},
			},
			clientId: "client",
			wantSentMessage: &protocol.Message{
				MessageType: protocol.ChallengeResponse,
				Payload:     `{"Version":0,"ZerosCount":0,"Date":null,"Resource":"","Extension":"","Rand":"","Counter":0}`,
			},
			wantErr: nil,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			srv := &Server{
				storageSet:       testCase.serviceSetup.storageSet,
				clock:            testCase.serviceSetup.clock,
				challengeBuilder: testCase.serviceSetup.challengeBuilder,
			}

			connMock := bytes.Buffer{}

			err := srv.sendChallengeToClient(context.Background(), &connMock, testCase.clientId)
			if testCase.wantErr != nil {
				if err == nil {
					t.Errorf("Server.sendChallengeToClient() - unexpected ouput. Got %v, expected error %v", string(connMock.Bytes()), testCase.wantErr)
					return
				}
				if err.Error() != testCase.wantErr.Error() {
					t.Errorf("Server.sendChallengeToClient() - unexpected error. Got %v, expected %v", err, testCase.wantErr)
				}
				return
			}
			if err != nil {
				t.Errorf("Server.sendChallengeToClient() - unexpected error: %v", err)
				return
			}

			msg, err := protocol.Read(bufio.NewReader(&connMock))
			if err != nil {
				t.Errorf("Server.sendChallengeToClient() - unparsable message sent: %v", err)
				return
			}
			if diff := cmp.Diff(msg, testCase.wantSentMessage); diff != "" {
				t.Errorf("Server.sendChallengeToClient() - unexpected output. Diff: %v", diff)
			}
		})
	}
}

func TestServer_checkClientChallengeSolution(t *testing.T) {
	defaultTestDuration := time.Minute

	longPastTime := time.Unix(0, 0)

	type ServiceSetup struct {
		storageSet       storage.StorageSet
		clock            clock.Clock
		challengeBuilder pow.POWChallengeBuilder
	}
	testCases := []struct {
		name              string
		serviceSetup      *ServiceSetup
		clientId          string
		challengeSolution pow.POWChallenge
		want              bool
		wantErr           error
	}{
		{
			name:         "clientId doesn't correspond to solution's resource",
			serviceSetup: &ServiceSetup{},
			clientId:     "wrong client",
			challengeSolution: &pow.POWChallengeMock{
				Resource: "client",
			},
			want:    false,
			wantErr: errors.New("invalid resource"),
		},
		{
			name: "fetching client token from storage set failed",
			serviceSetup: &ServiceSetup{
				storageSet: &storage.StorageSetMock{
					GetClientTokenFunc: func(_ context.Context, _ string) (string, error) {
						return "", errors.New("storage set error")
					},
				},
			},
			clientId: "client",
			challengeSolution: &pow.POWChallengeMock{
				Resource: "client",
			},
			want:    false,
			wantErr: fmt.Errorf("client token doesn't exist: %w", errors.New("storage set error")),
		},
		{
			name: "wrong client token",
			serviceSetup: &ServiceSetup{
				storageSet: &storage.StorageSetMock{
					GetClientTokenFunc: func(_ context.Context, _ string) (string, error) {
						return "wrong token", nil
					},
				},
			},
			clientId: "client",
			challengeSolution: &pow.POWChallengeMock{
				Resource: "client",
				RandVal:  "token",
			},
			want:    false,
			wantErr: errors.New("wrong client token"),
		},
		{
			name: "deleting from storage set failed",
			serviceSetup: &ServiceSetup{
				storageSet: &storage.StorageSetMock{
					GetClientTokenFunc: func(_ context.Context, _ string) (string, error) {
						return "token", nil
					},
					DeleteFunc: func(_ context.Context, _ string) error {
						return errors.New("storage set error")
					},
				},
			},
			clientId: "client",
			challengeSolution: &pow.POWChallengeMock{
				Resource: "client",
				RandVal:  "token",
			},
			want:    false,
			wantErr: fmt.Errorf("deleting from storage set failed: %w", errors.New("storage set error")),
		},
		{
			name: "challenge expired",
			serviceSetup: &ServiceSetup{
				storageSet: &storage.StorageSetMock{
					GetClientTokenFunc: func(_ context.Context, _ string) (string, error) {
						return "token", nil
					},
					DeleteFunc: func(_ context.Context, _ string) error {
						return nil
					},
				},
				challengeBuilder: &pow.POWChallengeBuilderMock{
					GetChallengeDurationFunc: func() *time.Duration {
						return &defaultTestDuration
					},
				},
				clock: &clock.ClockMock{
					Counter: 1231270371, // very big time frame
				},
			},
			clientId: "client",
			challengeSolution: &pow.POWChallengeMock{
				Resource: "client",
				RandVal:  "token",
				Date:     &longPastTime,
			},
			want:    false,
			wantErr: errors.New("challenge expired"),
		},
		{
			name: "verification failed with error",
			serviceSetup: &ServiceSetup{
				storageSet: &storage.StorageSetMock{
					GetClientTokenFunc: func(_ context.Context, _ string) (string, error) {
						return "token", nil
					},
					DeleteFunc: func(_ context.Context, _ string) error {
						return nil
					},
				},
				challengeBuilder: &pow.POWChallengeBuilderMock{
					GetChallengeDurationFunc: func() *time.Duration {
						return &defaultTestDuration
					},
				},
				clock: &clock.ClockMock{
					Counter: longPastTime.Unix(),
				},
			},
			clientId: "client",
			challengeSolution: &pow.POWChallengeMock{
				Resource: "client",
				RandVal:  "token",
				Date:     &longPastTime,
				VerifyFunc: func() (bool, error) {
					return false, errors.New("verification error")
				},
			},
			want:    false,
			wantErr: fmt.Errorf("verifying challenge solution failed: %w", errors.New("verification error")),
		},
		{
			name: "verification failed - bad solution",
			serviceSetup: &ServiceSetup{
				storageSet: &storage.StorageSetMock{
					GetClientTokenFunc: func(_ context.Context, _ string) (string, error) {
						return "token", nil
					},
					DeleteFunc: func(_ context.Context, _ string) error {
						return nil
					},
				},
				challengeBuilder: &pow.POWChallengeBuilderMock{
					GetChallengeDurationFunc: func() *time.Duration {
						return &defaultTestDuration
					},
				},
				clock: &clock.ClockMock{
					Counter: longPastTime.Unix(),
				},
			},
			clientId: "client",
			challengeSolution: &pow.POWChallengeMock{
				Resource: "client",
				RandVal:  "token",
				Date:     &longPastTime,
				VerifyFunc: func() (bool, error) {
					return false, nil
				},
			},
			want: false,
		},
		{
			name: "correct solution",
			serviceSetup: &ServiceSetup{
				storageSet: &storage.StorageSetMock{
					GetClientTokenFunc: func(_ context.Context, _ string) (string, error) {
						return "token", nil
					},
					DeleteFunc: func(_ context.Context, _ string) error {
						return nil
					},
				},
				challengeBuilder: &pow.POWChallengeBuilderMock{
					GetChallengeDurationFunc: func() *time.Duration {
						return &defaultTestDuration
					},
				},
				clock: &clock.ClockMock{
					Counter: longPastTime.Unix(),
				},
			},
			clientId: "client",
			challengeSolution: &pow.POWChallengeMock{
				Resource: "client",
				RandVal:  "token",
				Date:     &longPastTime,
				VerifyFunc: func() (bool, error) {
					return true, nil
				},
			},
			want: true,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			srv := &Server{
				storageSet:       testCase.serviceSetup.storageSet,
				clock:            testCase.serviceSetup.clock,
				challengeBuilder: testCase.serviceSetup.challengeBuilder,
			}

			connMock := bytes.Buffer{}

			got, err := srv.checkClientChallengeSolution(context.Background(), testCase.clientId, testCase.challengeSolution)
			if testCase.wantErr != nil {
				if err == nil {
					t.Errorf("Server.checkClientChallengeSolution() - unexpected ouput. Got %v, expected error %v", string(connMock.Bytes()), testCase.wantErr)
					return
				}
				if err.Error() != testCase.wantErr.Error() {
					t.Errorf("Server.checkClientChallengeSolution() - unexpected error. Got %v, expected %v", err, testCase.wantErr)
				}
				return
			}
			if err != nil {
				t.Errorf("Server.checkClientChallengeSolution() - unexpected error: %v", err)
				return
			}
			if got != testCase.want {
				t.Errorf("Server.checkClientChallengeSolution() - unexpected output. Got: %t, expected: %t", got, testCase.want)
			}
		})
	}
}
