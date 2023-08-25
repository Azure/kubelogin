// Copyright (c) Microsoft Corporation.
// Licensed under the MIT license.

package pop

import (
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
)

type PopAuthenticationScheme struct {
	// host is the u claim we will add on the pop token
	Host   string
	PoPKey PoPKey
}

func (as *PopAuthenticationScheme) TokenRequestParams() map[string]string {
	return map[string]string{
		"token_type": popTokenType,
		"req_cnf":    as.PoPKey.ReqCnf(),
	}
}

func (as *PopAuthenticationScheme) KeyID() string {
	return as.PoPKey.KeyID()
}

func (as *PopAuthenticationScheme) FormatAccessToken(accessToken string) (string, error) {
	ts := time.Now().Unix()
	nonce := uuid.New().String()
	nonce = strings.ReplaceAll(nonce, "-", "")
	header := fmt.Sprintf(`{"typ":"%s","alg":"%s","kid":"%s"}`, popTokenType, as.PoPKey.Alg(), as.PoPKey.KeyID())
	headerB64 := base64.RawURLEncoding.EncodeToString([]byte(header))
	payload := fmt.Sprintf(`{"at":"%s","ts":%d,"u":"%s","cnf":{"jwk":%s},"nonce":"%s"}`, accessToken, ts, as.Host, as.PoPKey.JWK(), nonce)
	payloadB64 := base64.RawURLEncoding.EncodeToString([]byte(payload))
	h256 := sha256.Sum256([]byte(headerB64 + "." + payloadB64))
	signature, err := as.PoPKey.Sign(h256[:])
	if err != nil {
		return "", err
	}
	signatureB64 := base64.RawURLEncoding.EncodeToString(signature)

	return headerB64 + "." + payloadB64 + "." + signatureB64, nil
}

func (as *PopAuthenticationScheme) AccessTokenType() string {
	return popTokenType
}
