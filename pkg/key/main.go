package key

import (
	"bytes"
	"crypto/rand"
	"encoding/ascii85"

	"github.com/martinlindhe/base36"
	"github.com/pkg/errors"
	"golang.org/x/crypto/ed25519"
)

const seedLength = 40

type Public struct {
	raw ed25519.PublicKey
	str string
}

func (p *Public) Verify(message, sig []byte) bool {
	return ed25519.Verify(p.raw, message, sig)
}

func (p *Public) String() string {
	return p.str
}

func (p *Public) Ed25519() ed25519.PublicKey {
	return p.raw
}

type Private struct {
	raw ed25519.PrivateKey
	str string
}

func (p *Private) Sign(message []byte) []byte {
	return ed25519.Sign(p.raw, message)
}

func (p *Private) String() string {
	return p.str
}

func (p *Private) Ed25519() ed25519.PrivateKey {
	return p.raw
}

func Generate() (public Public, private Private) {
	public.raw, private.raw, _ = ed25519.GenerateKey(rand.Reader)
	public.str = base36.EncodeBytes([]byte(public.raw))

	privateBase85 := make([]byte, ascii85.MaxEncodedLen(len(private.raw[:ed25519.PublicKeySize])))
	ascii85.Encode(privateBase85, private.raw[:ed25519.PublicKeySize])
	private.str = string(privateBase85) + public.str

	return public, private
}

func ParsePublic(str string) (Public, error) {
	publicBytes := base36.DecodeToBytes(str)
	if len(publicBytes) != ed25519.PublicKeySize {
		return Public{}, errors.New("invalid key size")
	}

	return Public{
		raw: ed25519.PublicKey(publicBytes),
		str: str,
	}, nil
}

func ParsePrivate(str string) (Private, error) {
	seed := make([]byte, ed25519.SeedSize)
	if _, _, err := ascii85.Decode(seed, []byte(str[:seedLength]), true); err != nil {
		return Private{}, errors.Wrap(err, "invalid seed")
	}

	if len(seed) != ed25519.SeedSize {
		return Private{}, errors.New("invalid seed size")
	}

	privateKey := ed25519.NewKeyFromSeed(seed)
	publicBytes := base36.DecodeToBytes(str[seedLength:])

	if bytes.Compare(privateKey[ed25519.SeedSize:], publicBytes) != 0 {
		return Private{}, errors.New("public part does not match")
	}

	return Private{
		raw: privateKey,
		str: str,
	}, nil
}
