package quic

import (
	"bytes"
	"testing"
)

func TestPackCodec_Encode(t *testing.T) {
	codec := NewPackCodec()

	tests := []struct {
		name     string
		pack     *Pack
		expected []byte
	}{
		{
			name: "empty payload",
			pack: &Pack{
				Head: PackHead{
					SQID:    123,
					OpCode:  456,
					Version: 1,
				},
				Payload: nil,
			},
			expected: []byte{
				0x00, 0x00, 0x00, 0x0c, // Len: 12
				0x00, 0x00, 0x00, 0x7b, // SQID: 123
				0x01, 0xc8, // OpCode: 456
				0x00, 0x01, // Version: 1
			},
		},
		{
			name: "with payload",
			pack: &Pack{
				Head: PackHead{
					SQID:    789,
					OpCode:  101,
					Version: 1,
				},
				Payload: []byte("hello"),
			},
			expected: []byte{
				0x00, 0x00, 0x00, 0x11, // Len: 17 (12 + 5)
				0x00, 0x00, 0x03, 0x15, // SQID: 789
				0x00, 0x65, // OpCode: 101
				0x00, 0x01, // Version: 1
				'h', 'e', 'l', 'l', 'o', // Payload: "hello"
			},
		},
		{
			name: "default opcode",
			pack: &Pack{
				Head: PackHead{
					SQID:    1,
					OpCode:  0, // Should default to OpCodeResOK
					Version: 1,
				},
				Payload: []byte("test"),
			},
			expected: []byte{
				0x00, 0x00, 0x00, 0x10, // Len: 16 (12 + 4)
				0x00, 0x00, 0x00, 0x01, // SQID: 1
				0x00, 0x00, // OpCode: 0 (OpCodeResOK)
				0x00, 0x01, // Version: 1
				't', 'e', 's', 't', // Payload: "test"
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := codec.Encode(tt.pack)
			if err != nil {
				t.Fatalf("Encode() error = %v", err)
			}
			if !bytes.Equal(result, tt.expected) {
				t.Errorf("Encode() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestPackCodec_Decode(t *testing.T) {
	codec := NewPackCodec()

	tests := []struct {
		name     string
		data     []byte
		expected *Pack
		wantErr  bool
	}{
		{
			name: "empty payload",
			data: []byte{
				0x00, 0x00, 0x00, 0x0c, // Len: 12
				0x00, 0x00, 0x00, 0x7b, // SQID: 123
				0x01, 0xc8, // OpCode: 456
				0x00, 0x01, // Version: 1
			},
			expected: &Pack{
				Head: PackHead{
					Len:     12,
					SQID:    123,
					OpCode:  456,
					Version: 1,
				},
				Payload: nil,
			},
		},
		{
			name: "with payload",
			data: []byte{
				0x00, 0x00, 0x00, 0x11, // Len: 17
				0x00, 0x00, 0x03, 0x15, // SQID: 789
				0x00, 0x65, // OpCode: 101
				0x00, 0x01, // Version: 1
				'h', 'e', 'l', 'l', 'o', // Payload: "hello"
			},
			expected: &Pack{
				Head: PackHead{
					Len:     17,
					SQID:    789,
					OpCode:  101,
					Version: 1,
				},
				Payload: []byte("hello"),
			},
		},
		{
			name:    "packet too small",
			data:    []byte{0x00, 0x00, 0x00}, // Only 3 bytes
			wantErr: true,
		},
		{
			name: "length mismatch",
			data: []byte{
				0x00, 0x00, 0x00, 0x20, // Len: 32 (but actual data is shorter)
				0x00, 0x00, 0x00, 0x01, // SQID: 1
				0x00, 0x01, // OpCode: 1
				0x00, 0x01, // Version: 1
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := codec.Decode(tt.data)
			if tt.wantErr {
				if err == nil {
					t.Errorf("Decode() expected error but got none")
				}
				return
			}
			if err != nil {
				t.Fatalf("Decode() error = %v", err)
			}
			if result.Head.Len != tt.expected.Head.Len ||
				result.Head.SQID != tt.expected.Head.SQID ||
				result.Head.OpCode != tt.expected.Head.OpCode ||
				result.Head.Version != tt.expected.Head.Version {
				t.Errorf("Decode() head = %+v, want %+v", result.Head, tt.expected.Head)
			}
			if !bytes.Equal(result.Payload, tt.expected.Payload) {
				t.Errorf("Decode() payload = %v, want %v", result.Payload, tt.expected.Payload)
			}
		})
	}
}

func TestPackCodec_EncodeDecodeRoundTrip(t *testing.T) {
	codec := NewPackCodec()

	original := &Pack{
		Head: PackHead{
			SQID:    12345,
			OpCode:  9999,
			Version: 1,
		},
		Payload: []byte("This is a test payload with some data"),
	}

	// Encode
	encoded, err := codec.Encode(original)
	if err != nil {
		t.Fatalf("Encode() error = %v", err)
	}

	// Decode
	decoded, err := codec.Decode(encoded)
	if err != nil {
		t.Fatalf("Decode() error = %v", err)
	}

	// Compare
	if decoded.Head.SQID != original.Head.SQID ||
		decoded.Head.OpCode != original.Head.OpCode ||
		decoded.Head.Version != original.Head.Version {
		t.Errorf("Round trip head mismatch: got %+v, want %+v", decoded.Head, original.Head)
	}
	if !bytes.Equal(decoded.Payload, original.Payload) {
		t.Errorf("Round trip payload mismatch: got %v, want %v", decoded.Payload, original.Payload)
	}
}
