package auth

import (
	"context"
	"encoding/base64"
	"strings"

	"github.com/go-oauth2/oauth2/v4"
	"github.com/go-oauth2/oauth2/v4/generates"
	"github.com/golang-jwt/jwt"
	"github.com/google/uuid"
	"github.com/matelang/jwt-go-aws-kms/v2/jwtkms"
)

// KMSJWTAccessGenerate allows for the generation of an access token using credentials pulled from KMS
type KMSJWTAccessGenerate struct {
	keyID  string
	method jwt.SigningMethod
	config *jwtkms.Config
}

// NewKMSJWTAccessGenerate creates a new JWT access generator with the AWS config, key ID and signing
// method provided
func NewKMSJWTAccessGenerate(client jwtkms.KMSClient, keyID string, method jwt.SigningMethod) *KMSJWTAccessGenerate {
	return &KMSJWTAccessGenerate{
		keyID:  keyID,
		method: method,
		config: jwtkms.NewKMSConfig(client, keyID, false),
	}
}

// Token generates a new access token from the OAuth2 basic generate information and a flag determining
// whether a refresh token should be generated or not
func (generate *KMSJWTAccessGenerate) Token(ctx context.Context, data *oauth2.GenerateBasic,
	isGenRefresh bool) (string, string, error) {

	// First, create the JWT access claims we'll be encoding
	claims := &generates.JWTAccessClaims{
		StandardClaims: jwt.StandardClaims{
			Audience:  data.Client.GetID(),
			Subject:   data.UserID,
			ExpiresAt: data.TokenInfo.GetAccessCreateAt().Add(data.TokenInfo.GetAccessExpiresIn()).Unix(),
		},
	}

	// Next, create a new token using the method and claims and add our key ID to the token header
	token := jwt.NewWithClaims(generate.method, claims)
	token.Header["kid"] = generate.keyID

	// Now, sign the string using our key from KMS; if this fails then return an error
	access, err := token.SignedString(generate.config.WithContext(ctx))
	if err != nil {
		return "", "", err
	}

	// Finally, if we want to add a refresh token to the access code then do so here
	var refresh string
	if isGenRefresh {
		token := uuid.NewSHA1(uuid.Must(uuid.NewRandom()), []byte(access)).String()
		refresh = base64.URLEncoding.EncodeToString([]byte(token))
		refresh = strings.ToUpper(strings.TrimRight(refresh, "="))
	}

	// Return the access token and refresh token
	return access, refresh, nil
}
