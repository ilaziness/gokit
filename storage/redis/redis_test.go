package redis

import (
	"testing"

	"github.com/ilaziness/gokit/config"
	"github.com/stretchr/testify/assert"
)

func TestLock(t *testing.T) {
	Init(&config.Redis{
		Host: "127.0.0.1",
		Port: 6379,
	}, true)

	l1, err := Lock("test", 10)
	assert.Equal(t, nil, err)
	b1, _ := l1.Unlock()
	assert.Equal(t, true, b1)

	l2, err := Lock("test", 10)
	assert.Equal(t, nil, err)
	b2, _ := l2.Unlock()
	assert.Equal(t, true, b2)

	l3, err := Lock("test", 10)
	assert.Equal(t, nil, err)
	_, err = Lock("test", 10)
	assert.NotEqual(t, nil, err)
	b3, _ := l3.Unlock()
	assert.Equal(t, true, b3)

	l5, err := Lock("test", 10)
	assert.Equal(t, nil, err)
	l6, err := Lock("test2", 10)
	assert.Equal(t, nil, err)
	b5, _ := l5.Unlock()
	assert.Equal(t, true, b5)
	b6, _ := l6.Unlock()
	assert.Equal(t, true, b6)
}
