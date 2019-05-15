package key_test

import (
	"fmt"
	"github.com/xdarksome/scp"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestKeys(t *testing.T) {
	public, private := scp.Generate()
	fmt.Printf("public: %s\n", public.String())
	fmt.Printf("private: %s\n", private.String())

	parsedPub, err := scp.ParsePublic(public.String())
	require.NoError(t, err)
	require.Equal(t, public, parsedPub)

	parsedPrivate, err := scp.ParsePrivate(private.String())
	require.NoError(t, err)
	require.Equal(t, private, parsedPrivate)

	message := []byte("Sign me")
	signature := private.Sign(message)
	ok := public.Verify(message, signature)
	require.True(t, ok)
}
