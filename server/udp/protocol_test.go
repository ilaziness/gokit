package udp

import (
	"testing"
)

func TestPackCodec_Encode(t *testing.T) {
	codec := NewPackCodec()

	tests := []struct {
		name    string
		pack    *Pack
		wantErr bool
	}{
		{
			name: "empty payload",
			pack: &Pack{
				Head: PackHead{
					SQID:    123,
					OpCode:  1,
					Version: Version1,
				},
				Payload: nil,
			},
			wantErr: false,
		},
		{
			name: "with payload",
			pack: &Pack{
				Head: PackHead{
					SQID:    456,
					OpCode:  2,
					Version: Version1,
				},
				Payload: []byte("hello world"),
			},
			wantErr: false,
		},
		{
			name: "large payload exceeding UDP limit",
			pack: &Pack{
				Head: PackHead{
					SQID:    789,
					OpCode:  3,
					Version: Version1,
				},
				Payload: make([]byte, MaxUDPSize),
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data, err := codec.Encode(tt.pack)
			if (err != nil) != tt.wantErr {
				t.Errorf("PackCodec.Encode() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr {
				if len(data) != int(tt.pack.Head.Len) {
					t.Errorf("encoded data length = %d, want %d", len(data), tt.pack.Head.Len)
				}
			}
		})
	}
}

func TestPackCodec_Decode(t *testing.T) {
	codec := NewPackCodec()

	// 先编码一个包
	originalPack := &Pack{
		Head: PackHead{
			SQID:    123,
			OpCode:  1,
			Version: Version1,
		},
		Payload: []byte("test payload"),
	}

	data, err := codec.Encode(originalPack)
	if err != nil {
		t.Fatalf("failed to encode pack: %v", err)
	}

	tests := []struct {
		name    string
		data    []byte
		wantErr bool
	}{
		{
			name:    "valid packet",
			data:    data,
			wantErr: false,
		},
		{
			name:    "packet too small",
			data:    []byte{1, 2, 3},
			wantErr: true,
		},
		{
			name:    "empty data",
			data:    []byte{},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pack, err := codec.Decode(tt.data)
			if (err != nil) != tt.wantErr {
				t.Errorf("PackCodec.Decode() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr {
				if pack.Head.SQID != originalPack.Head.SQID {
					t.Errorf("decoded SQID = %d, want %d", pack.Head.SQID, originalPack.Head.SQID)
				}
				if pack.Head.OpCode != originalPack.Head.OpCode {
					t.Errorf("decoded OpCode = %d, want %d", pack.Head.OpCode, originalPack.Head.OpCode)
				}
				if string(pack.Payload) != string(originalPack.Payload) {
					t.Errorf("decoded payload = %s, want %s", pack.Payload, originalPack.Payload)
				}
			}
		})
	}
}

func TestPackCodec_EncodeDecodeRoundTrip(t *testing.T) {
	codec := NewPackCodec()

	originalPack := &Pack{
		Head: PackHead{
			SQID:    999,
			OpCode:  uint16(OpCodePing),
			Version: Version1,
		},
		Payload: []byte("ping test"),
	}

	// 编码
	data, err := codec.Encode(originalPack)
	if err != nil {
		t.Fatalf("encode failed: %v", err)
	}

	// 解码
	decodedPack, err := codec.Decode(data)
	if err != nil {
		t.Fatalf("decode failed: %v", err)
	}

	// 验证
	if decodedPack.Head.SQID != originalPack.Head.SQID {
		t.Errorf("SQID mismatch: got %d, want %d", decodedPack.Head.SQID, originalPack.Head.SQID)
	}
	if decodedPack.Head.OpCode != originalPack.Head.OpCode {
		t.Errorf("OpCode mismatch: got %d, want %d", decodedPack.Head.OpCode, originalPack.Head.OpCode)
	}
	if decodedPack.Head.Version != originalPack.Head.Version {
		t.Errorf("Version mismatch: got %d, want %d", decodedPack.Head.Version, originalPack.Head.Version)
	}
	if string(decodedPack.Payload) != string(originalPack.Payload) {
		t.Errorf("Payload mismatch: got %s, want %s", decodedPack.Payload, originalPack.Payload)
	}
}
