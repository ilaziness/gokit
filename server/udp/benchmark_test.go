package udp

import (
	"testing"
)

func BenchmarkPackCodec_Encode(b *testing.B) {
	codec := NewPackCodec()
	pack := &Pack{
		Head: PackHead{
			SQID:    123,
			OpCode:  1000,
			Version: Version1,
		},
		Payload: []byte("benchmark test payload data"),
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := codec.Encode(pack)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkPackCodec_Decode(b *testing.B) {
	codec := NewPackCodec()
	pack := &Pack{
		Head: PackHead{
			SQID:    123,
			OpCode:  1000,
			Version: Version1,
		},
		Payload: []byte("benchmark test payload data"),
	}

	data, err := codec.Encode(pack)
	if err != nil {
		b.Fatal(err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := codec.Decode(data)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkPackCodec_EncodeDecodeRoundTrip(b *testing.B) {
	codec := NewPackCodec()
	pack := &Pack{
		Head: PackHead{
			SQID:    123,
			OpCode:  1000,
			Version: Version1,
		},
		Payload: []byte("benchmark test payload data"),
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		data, err := codec.Encode(pack)
		if err != nil {
			b.Fatal(err)
		}
		_, err = codec.Decode(data)
		if err != nil {
			b.Fatal(err)
		}
	}
}
