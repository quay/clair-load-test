package main

import (
	"encoding/base64"
	"time"

	"github.com/quay/zlog"
	"github.com/urfave/cli/v2"
	"gopkg.in/square/go-jose.v2"
	"gopkg.in/square/go-jose.v2/jwt"
)

// Constants defined here.
const (
	TokenIssuer         = "clairctl"
	TokenValidityPeriod = time.Hour * 24 * 7
)

// CreateTokenCmd handles createtoken CLI.
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

// createTokenAction to drive token action from the CLI options.
// It returns an error if any during the execution.
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

// CreateToken creates a token from the input PSK key.
// It returns a token string and an error if any during the execution.
func CreateToken(key string) (tok string, err error) {
	decKey, err := base64.StdEncoding.DecodeString(key)
	if err != nil {
		return "", err
	}
	sk := jose.SigningKey{
		Algorithm: jose.HS256,
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
