package quic

import (
	"testing"
)

func BenchmarkPackCodec_Encode(b *testing.B) {
	codec := NewPackCodec()
	pack := &Pack{
		Head: PackHead{
			SQID:    12345,
			OpCode:  1000,
			Version: 1,
		},
		Payload: make([]byte, 1024), // 1KB payload
	}

	b.ResetTimer()
	b.ReportAllocs()

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
			SQID:    12345,
			OpCode:  1000,
			Version: 1,
		},
		Payload: make([]byte, 1024), // 1KB payload
	}

	data, err := codec.Encode(pack)
	if err != nil {
		b.Fatal(err)
	}

	b.ResetTimer()
	b.ReportAllocs()

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
			SQID:    12345,
			OpCode:  1000,
			Version: 1,
		},
		Payload: make([]byte, 1024), // 1KB payload
	}

	b.ResetTimer()
	b.ReportAllocs()

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

func BenchmarkPackCodec_SmallPayload(b *testing.B) {
	codec := NewPackCodec()
	pack := &Pack{
		Head: PackHead{
			SQID:    12345,
			OpCode:  1000,
			Version: 1,
		},
		Payload: []byte("hello world"), // Small payload
	}

	b.ResetTimer()
	b.ReportAllocs()

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

func BenchmarkPackCodec_LargePayload(b *testing.B) {
	codec := NewPackCodec()
	pack := &Pack{
		Head: PackHead{
			SQID:    12345,
			OpCode:  1000,
			Version: 1,
		},
		Payload: make([]byte, 64*1024), // 64KB payload
	}

	b.ResetTimer()
	b.ReportAllocs()

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
