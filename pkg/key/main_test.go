package key_test

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/xdarksome/scp/pkg/key"
)

func TestKeys(t *testing.T) {
	public, private := key.Generate()
	fmt.Printf("public: %s\n", public.String())
	fmt.Printf("private: %s\n", private.String())

	parsedPub, err := key.ParsePublic(public.String())
	require.NoError(t, err)
	require.Equal(t, public, parsedPub)

	parsedPrivate, err := key.ParsePrivate(private.String())
	require.NoError(t, err)
	require.Equal(t, private, parsedPrivate)

	message := []byte("Sign me")
	signature := private.Sign(message)
	ok := public.Verify(message, signature)
	require.True(t, ok)
}
