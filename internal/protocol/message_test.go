package protocol

import (
	"bufio"
	"bytes"
	"errors"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestReadMessage(t *testing.T) {
	testCases := []struct {
		name    string
		input   string
		want    *Message
		wantErr error
	}{
		{
			name:  "empty message",
			input: "\n",
			want: &Message{
				MessageType: ForceQuit,
				Payload:     "",
			},
			wantErr: errors.New("unexpected end of JSON input"),
		},
		{
			name:  "only message type is provided",
			input: "{\"type\":\"Force quit\"}\n",
			want: &Message{
				MessageType: ForceQuit,
				Payload:     "",
			},
		},
		{
			name:  "message type and payload is present",
			input: "{\"type\":\"Resource response\",\"payload\":\"abracadabra\"}\n",
			want: &Message{
				MessageType: ResourceResponse,
				Payload:     "abracadabra",
			},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			bufReader := bufio.NewReader(strings.NewReader(testCase.input))

			got, err := Read(bufReader)
			if testCase.wantErr != nil {
				if err == nil {
					t.Errorf("Unexpected output while reading the Message. Got %v, expected %v", got, testCase.wantErr)
					return
				}
				if err.Error() != testCase.wantErr.Error() {
					t.Errorf("Unexpected error while reading the Message. Got %v, expected %v", err, testCase.wantErr)
				}
				return
			}
			if err != nil {
				t.Errorf("Unexpected error while reading the Message: %v", err)
			}
			if diff := cmp.Diff(got, testCase.want); diff != "" {
				t.Errorf("Unexpected output occured while reading the Message: %v", err)
			}
		})
	}
}

func TestWriteMessage(t *testing.T) {
	testCases := []struct {
		name    string
		input   *Message
		want    string
		wantErr error
	}{
		{
			name:    "empty message",
			input:   nil,
			want:    "null\n",
			wantErr: errors.New("message is nil"),
		},
		{
			name: "only message type is provided",
			input: &Message{
				MessageType: ForceQuit,
			},
			want: "{\"type\":\"Force quit\"}\n",
		},
		{
			name: "message type and payload is present",
			input: &Message{
				MessageType: ResourceResponse,
				Payload:     "abracadabra",
			},
			want: "{\"type\":\"Resource response\",\"payload\":\"abracadabra\"}\n",
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			writer := bytes.Buffer{}

			connBufWriter := bufio.NewWriter(&writer)

			err := Write(connBufWriter, testCase.input)
			if testCase.wantErr != nil {
				if err == nil {
					t.Errorf("Unexpected output while writing the Message. Got %v, expected error %v", string(writer.Bytes()), testCase.wantErr)
					return
				}
				if err.Error() != testCase.wantErr.Error() {
					t.Errorf("Unexpected error while writing the Message. Got %v, expected %v", err, testCase.wantErr)
				}
				return
			}
			if err != nil {
				t.Errorf("Unexpected error while writing the Message: %v", err)
			}
			if diff := cmp.Diff(string(writer.Bytes()), testCase.want); diff != "" {
				t.Errorf("Unexpected output while writing the Message: %v", err)
			}
		})
	}
}
