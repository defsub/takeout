// Copyright (C) 2021 The Takeout Authors.
//
// This file is part of Takeout.
//
// Takeout is free software: you can redistribute it and/or modify it under the
// terms of the GNU Affero General Public License as published by the Free
// Software Foundation, either version 3 of the License, or (at your option)
// any later version.
//
// Takeout is distributed in the hope that it will be useful, but WITHOUT ANY
// WARRANTY; without even the implied warranty of MERCHANTABILITY or FITNESS
// FOR A PARTICULAR PURPOSE.  See the GNU Affero General Public License for
// more details.
//
// You should have received a copy of the GNU Affero General Public License
// along with Takeout.  If not, see <https://www.gnu.org/licenses/>.

package token

import (
	"crypto/rsa"
	"encoding/base64"
	"encoding/binary"
	"errors"
	"fmt"
	"math/big"

	"github.com/defsub/takeout/config"
	"github.com/defsub/takeout/lib/client"
	"github.com/golang-jwt/jwt"
)

var (
	ErrExpectedRSA256     = errors.New("Token signature must be RS256")
	ErrKeyNotFound        = errors.New("Key not found")
	ErrSigningKeyRequired = errors.New("Signing key required")
	ErrAudience           = errors.New("Audience mismatch")
)

const (
	GoogleOpenIDConfigurationURI = "https://accounts.google.com/.well-known/openid-configuration"
	GoogleJWKSURI                = "https://www.googleapis.com/oauth2/v3/certs"

	UseSignature = "sig"

	HeaderAlgorithm = "alg"
	HeaderKeyID     = "kid"

	ClaimIssuer     = "iss"
	ClaimAudience   = "aud"
	ClaimSubject    = "sub"
	ClaimExpiration = "exp"
	ClaimNotBefore  = "nbf"
	ClaimIssuedAt   = "iat"
)

type OpenIDConfiguration struct {
	Issuer                            string   `json:"issuer"`
	AuthorizationEndpoint             string   `json:"authorization_endpoint"`
	DeviceAuthorizationEndpoint       string   `json:"device_authorization_endpoint"`
	TokenEndpoint                     string   `json:"token_endpoint"`
	UserInfoEndpoint                  string   `json:"userinfo_endpoint"`
	RevocationEndpoint                string   `json:"revocation_endpoint"`
	JWKS_URI                          string   `json:"jwks_uri"`
	ResponseTypesSupported            []string `json:"response_types_supported"`
	SubjectTypesSupported             []string `json:"subject_types_supported"`
	IdTokenSigningAlgValuesSupported  []string `json:"id_token_signing_alg_values_supported"`
	ScopesSupported                   []string `json:"scopes_supported"`
	TokenEndpointAuthMethodsSupported []string `json:"code_challenge_methods_supported"`
	GrantTypesSupported               []string `json:"grant_types_supported"`
}

type JSONWebKey struct {
	KeyID     string `json:"kid"`
	KeyType   string `json:"kty"`
	Algorithm string `json:"alg"`
	Use       string `json:"use"`
	N         string `json:"n"` // modulus
	E         string `json:"e"` // public exponent
}

type JWKS struct {
	Keys []JSONWebKey `json:"keys"`
}

// See https://ldapwiki.com/wiki/Openid-configuration
// https://[base-server-url]/.well-known/openid-configuration
func DiscoverConfiguration(config *config.Config, url string) (OpenIDConfiguration, error) {
	var result OpenIDConfiguration
	c := client.NewClient(config)
	err := c.GetJson(url, &result)
	return result, err
}

func GetJWKS(config *config.Config, url string) (JWKS, error) {
	var result JWKS
	c := client.NewClient(config)
	err := c.GetJson(url, &result)
	return result, err
}

func GoogleWebKey(config *config.Config, kid string) (JSONWebKey, error) {
	var result JSONWebKey
	cfg, err := DiscoverConfiguration(config, GoogleOpenIDConfigurationURI)
	if err != nil {
		return result, err
	}
	jwks, err := GetJWKS(config, cfg.JWKS_URI)
	if err != nil {
		return result, err
	}
	for _, k := range jwks.Keys {
		if k.KeyID == kid {
			return k, nil
		}
	}
	return result, ErrKeyNotFound
}

func (k JSONWebKey) PublicKey() (*rsa.PublicKey, error) {
	// exponent
	ebytes, err := base64.RawURLEncoding.DecodeString(k.E)
	if err != nil {
		return nil, err
	}
	for {
		if len(ebytes) >= 4 {
			break
		}
		ebytes = append([]byte{0}, ebytes...)
	}
	e := int(binary.BigEndian.Uint32(ebytes))

	// modulus
	nbytes, err := base64.RawURLEncoding.DecodeString(k.N)
	if err != nil {
		return nil, err
	}
	n := new(big.Int)
	n.SetBytes(nbytes)

	return &rsa.PublicKey{E: e, N: n}, nil
}

// This should be a very small cache, mapping JWT kid to an RSA public key
// fetched dynamically from the network. For now this is used to validate
// Google JWT request tokens to ensure they come from Google servers.
var keyCache map[string]*rsa.PublicKey = make(map[string]*rsa.PublicKey)

func ValidateGoogleToken(config *config.Config, tokenString, audience string) error {
	// first parse to find the public key
	var claims jwt.StandardClaims
	token, _, err := new(jwt.Parser).ParseUnverified(tokenString, &claims)

	// alg: RS256
	// kid: c3104c688c15e6b8e58e67a328780952b2174017
	// typ: JWT
	if token.Header[HeaderAlgorithm] != "RS256" {
		return ErrExpectedRSA256
	}

	// also check the audience ahead of time
	if claims.Audience != audience {
		return ErrAudience
	}

	// get the public key using the kid
	kid := token.Header[HeaderKeyID].(string)
	pub, ok := keyCache[kid]
	if !ok {
		key, err := GoogleWebKey(config, kid)
		if err != nil {
			return err
		}
		if key.Use != UseSignature {
			return ErrSigningKeyRequired
		}
		// now parse again with the public key to verify the signature
		pub, err = key.PublicKey()
		if err != nil {
			return err
		}
		// cache key for next time
		keyCache[kid] = pub
	}

	// now fully verify the token
	token, err = jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodRSA); !ok {
			return nil, fmt.Errorf("Unexpected signing method: %v", token.Header["alg"])
		}
		return pub, nil
	})
	if err != nil {
		return err
	}
	// looks legit!
	return nil
}
