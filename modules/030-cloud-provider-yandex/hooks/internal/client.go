/*
Copyright 2022 Flant JSC

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package internal

import (
	"crypto/rsa"
	"crypto/x509"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
	"time"

	jose "github.com/square/go-jose/v3"

	d8http "github.com/deckhouse/deckhouse/go_lib/dependency/http"
	yandexV1 "github.com/deckhouse/deckhouse/modules/030-cloud-provider-yandex/hooks/internal/v1"
)

type YandexAPI struct {
	client d8http.Client

	iamToken string
}

func NewYandexAPI(client d8http.Client) *YandexAPI {
	return &YandexAPI{
		client: client,
	}
}

func (a *YandexAPI) Init(sa *yandexV1.ServiceAccount) error {
	bearerToken, err := generateJWTKeyForGetIAMToken(sa)
	if err != nil {
		return err
	}

	return WithRetry(3, 3*time.Second, func() error {
		iamToken, err := a.getIAMToken(bearerToken)
		if err != nil {
			return err
		}

		a.iamToken = iamToken
		return nil
	})
}

func (a *YandexAPI) CreateAPIKey(serviceAccountID string) (apiKey, apiKeyID string, err error) {
	if a.iamToken == "" {
		return "", "", fmt.Errorf("api not init")
	}

	requestBody := yandexV1.APIKeyCreationRequest{
		ServiceAccountID: serviceAccountID,
		Description:      "Auto-generated by Deckhouse for cloud metrics exporter",
	}

	requestBodyBytes, err := json.Marshal(requestBody)
	if err != nil {
		return "", "", err
	}

	req, err := http.NewRequest(
		"POST",
		"https://iam.api.cloud.yandex.net/iam/v1/apiKeys",
		strings.NewReader(string(requestBodyBytes)))
	if err != nil {
		return "", "", err
	}
	req.Header.Set("Content-Type", "application/json")
	d8http.SetBearerToken(req, a.iamToken)

	resp, err := a.client.Do(req)
	if err != nil {
		return "", "", err
	}

	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		body, _ := ioutil.ReadAll(resp.Body)
		return "", "", fmt.Errorf("cannot create API-key %s: %s", resp.Status, body)
	}

	var data yandexV1.APIKeyCreationResponse

	err = json.NewDecoder(resp.Body).Decode(&data)
	if err != nil {
		return "", "", err
	}

	return data.Secret, data.APIKey.ID, err
}

func (a *YandexAPI) DeleteAPIKey(apiKeyID string) (err error) {
	if a.iamToken == "" {
		return fmt.Errorf("api not init")
	}

	if apiKeyID == "" {
		return fmt.Errorf("api key id not found")
	}

	req, err := http.NewRequest(
		"DELETE",
		fmt.Sprintf("https://iam.api.cloud.yandex.net/iam/v1/apiKeys/%s", apiKeyID),
		nil)
	if err != nil {
		return err
	}
	d8http.SetBearerToken(req, a.iamToken)

	resp, err := a.client.Do(req)
	if err != nil {
		return err
	}

	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		body, _ := ioutil.ReadAll(resp.Body)
		return fmt.Errorf("cannot create API-key %s: %s", resp.Status, body)
	}

	return nil
}

func (a *YandexAPI) getIAMToken(jwtForExchange string) (string, error) {
	requestBody := yandexV1.IAMTokenCreationRequest{
		JWT: jwtForExchange,
	}

	requestBodyBytes, err := json.Marshal(requestBody)
	if err != nil {
		return "", err
	}

	req, err := http.NewRequest(
		"POST",
		"https://iam.api.cloud.yandex.net/iam/v1/tokens",
		strings.NewReader(string(requestBodyBytes)))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := a.client.Do(req)
	if err != nil {
		return "", err
	}

	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		body, _ := ioutil.ReadAll(resp.Body)
		return "", fmt.Errorf("cannot create IAM token %s: %s", resp.Status, body)
	}

	var data yandexV1.IAMTokenCreationResponse

	err = json.NewDecoder(resp.Body).Decode(&data)
	if err != nil {
		return "", err
	}

	return data.IAMToken, err
}

func generateJWTKeyForGetIAMToken(sa *yandexV1.ServiceAccount) (string, error) {
	privateKey, err := parsePrivateKey([]byte(sa.PrivateKey))
	if err != nil {
		return "", fmt.Errorf("cannot parse private key: %s", err)
	}
	key := jose.SigningKey{
		Key:       privateKey,
		Algorithm: jose.PS256,
	}

	now := time.Now()
	nowUnix := now.Second()
	expiredAt := now.Add(5 * time.Minute).Second()
	payload := map[string]interface{}{
		"iss": sa.ServiceAccountID,
		"aud": "https://iam.api.cloud.yandex.net/iam/v1/tokens",
		"iat": nowUnix,
		"exp": expiredAt,
	}

	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return "", err
	}

	signer, err := jose.NewSigner(key, &jose.SignerOptions{
		ExtraHeaders: map[jose.HeaderKey]interface{}{
			"kid": sa.ID,
		},
	})
	if err != nil {
		return "", fmt.Errorf("cannot create signer: %s", err)
	}

	signedPayload, err := signer.Sign(payloadBytes)
	if err != nil {
		return "", fmt.Errorf("cannot sign payload: %s", err)
	}

	return signedPayload.CompactSerialize()
}

func parsePrivateKey(key []byte) (*rsa.PrivateKey, error) {
	var err error

	var block *pem.Block
	if block, _ = pem.Decode(key); block == nil {
		return nil, fmt.Errorf("cannot decode pem block")
	}

	var parsed interface{}
	if parsed, err = x509.ParsePKCS1PrivateKey(block.Bytes); err != nil {
		if parsed, err = x509.ParsePKCS8PrivateKey(block.Bytes); err != nil {
			return nil, fmt.Errorf("cannot decode private key")
		}
	}

	var pk *rsa.PrivateKey
	var ok bool
	if pk, ok = parsed.(*rsa.PrivateKey); !ok {
		return nil, fmt.Errorf("private key is not correct")
	}

	return pk, nil
}

func WithRetry(times int, sleep time.Duration, action func() error) error {
	var lastErr error
	for i := 0; i < times; i++ {
		if lastErr != nil {
			time.Sleep(sleep)
		}

		lastErr = nil
		lastErr = action()
		if lastErr == nil {
			return nil
		}
	}

	return lastErr
}
