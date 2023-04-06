package token

import (
	"encoding/base64"
	"time"

	"github.com/quay/zlog"
	"github.com/urfave/cli/v2"
	"gopkg.in/square/go-jose.v2"
	"gopkg.in/square/go-jose.v2/jwt"
)

const (
	TokenIssuer          = "clairctl"
	TokenValidityPeriod  = time.Hour * 24 * 7
)

var CreateTokenCmd = &cli.Command{
	Name:        "createtoken",
	Description: "Creates a JWT token given a psk",
	Usage:       "createtoken --key sdfvevefr==",
	Action:      createTokenAction,
	Flags: []cli.Flag{
		&cli.StringFlag{
			Name:    "key",
			Usage:   "--key ddsdfsdfsfsd==",
			Value:   "",
			EnvVars: []string{"PSK_KEY"},
		},
	},
}

func createTokenAction(c *cli.Context) error {
	ctx := c.Context
	key := c.String("key")
	zlog.Debug(ctx).Str("key", key).Msg("got md5 key")
	tok, err := CreateToken(key)
	if err != nil {
		return err
	}
	zlog.Info(ctx).Msg(tok)
	return nil
}

func CreateToken(key string) (tok string, err error) {
	decKey, err := getSigningKey(key)
	if err != nil {
		return "", err
	}
	sk := jose.SigningKey{
		Algorithm: getSigningAlgorithm(),
		Key:       decKey,
	}
	s, err := jose.NewSigner(sk, nil)
	if err != nil {
		return "", err
	}
	now := time.Now()

	// Mint the jwt.
	return jwt.Signed(s).Claims(&jwt.Claims{
		Issuer:    TokenIssuer,
		Expiry:    jwt.NewNumericDate(now.Add(TokenValidityPeriod)),
		IssuedAt:  jwt.NewNumericDate(now),
		NotBefore: jwt.NewNumericDate(now),
	}).CompactSerialize()

}

func getSigningAlgorithm() jose.SignatureAlgorithm {
	return jose.HS256
}

func getSigningKey(key string) ([]byte, error) {
	return base64.StdEncoding.DecodeString(key)
}