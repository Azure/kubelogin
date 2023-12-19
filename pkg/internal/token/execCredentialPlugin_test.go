package token

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/Azure/go-autorest/autorest/adal"
	"github.com/Azure/kubelogin/pkg/internal/token/mock_token"
	"go.uber.org/mock/gomock"
)

func TestExecCredentialPlugin(t *testing.T) {
	const (
		cacheFile = "cacheFile"
	)
	type testContext struct {
		tokenCache    *mock_token.MockTokenCache
		tokenProvider *mock_token.MockTokenProvider
		pluginWriter  *mock_token.MockExecCredentialWriter
	}
	testData := []struct {
		name              string
		options           *Options
		setupExpectations func(tc testContext)
		expectedError     string
	}{
		{
			name: "fail to read token cache",
			options: &Options{
				tokenCacheFile: cacheFile,
			},
			setupExpectations: func(tc testContext) {
				tc.tokenCache.EXPECT().Read(cacheFile).Return(adal.Token{}, errors.New("fail"))
			},
			expectedError: "unable to read from token cache: cacheFile, err: fail",
		},
		{
			name: "reading empty token from cache should invoke token flow",
			options: &Options{
				tokenCacheFile: cacheFile,
			},
			setupExpectations: func(tc testContext) {
				tc.tokenCache.EXPECT().Read(cacheFile).Return(adal.Token{}, nil)
				tc.tokenProvider.EXPECT().Token(gomock.Any()).Return(adal.Token{}, nil)
				tc.tokenCache.EXPECT().Write(cacheFile, adal.Token{}).Return(nil)
				tc.pluginWriter.EXPECT().Write(adal.Token{}, os.Stdout)
			},
		},
		{
			name: "when cached token is still valid, token provider is not invoked",
			options: &Options{
				tokenCacheFile: cacheFile,
				ServerID:       "apiServer",
			},
			setupExpectations: func(tc testContext) {
				cachedToken := adal.Token{
					Resource:  "apiServer",
					ExpiresOn: json.Number(fmt.Sprintf("%d", time.Now().AddDate(1, 0, 0).Unix())),
				}
				tc.tokenCache.EXPECT().Read(cacheFile).Return(cachedToken, nil)
				tc.pluginWriter.EXPECT().Write(cachedToken, os.Stdout)
			},
		},
		{
			name: "in legacy mode, when cached token is still valid, token provider is not invoked",
			options: &Options{
				tokenCacheFile: cacheFile,
				ServerID:       "apiServer",
				IsLegacy:       true,
			},
			setupExpectations: func(tc testContext) {
				cachedToken := adal.Token{
					Resource:  "spn:apiServer",
					ExpiresOn: json.Number(fmt.Sprintf("%d", time.Now().AddDate(1, 0, 0).Unix())),
				}
				tc.tokenCache.EXPECT().Read(cacheFile).Return(cachedToken, nil)
				tc.pluginWriter.EXPECT().Write(cachedToken, os.Stdout)
			},
		},
		{
			name: "when token expires and there is no refresh token, need to invoke token provider",
			options: &Options{
				tokenCacheFile: cacheFile,
				ServerID:       "apiServer",
			},
			setupExpectations: func(tc testContext) {
				cachedToken := adal.Token{
					Resource:  "apiServer",
					ExpiresOn: json.Number(fmt.Sprintf("%d", time.Now().AddDate(-1, 0, 0).Unix())),
				}
				refreshedToken := adal.Token{
					Resource:  "apiServer",
					ExpiresOn: json.Number(fmt.Sprintf("%d", time.Now().AddDate(1, 0, 0).Unix())),
				}
				tc.tokenCache.EXPECT().Read(cacheFile).Return(cachedToken, nil)
				tc.tokenProvider.EXPECT().Token(gomock.Any()).Return(refreshedToken, nil)
				tc.tokenCache.EXPECT().Write(cacheFile, refreshedToken).Return(nil)
				tc.pluginWriter.EXPECT().Write(refreshedToken, os.Stdout)
			},
		},
	}

	for _, data := range testData {
		t.Run(data.name, func(t *testing.T) {
			ctrl, tokenCache, tokenProvider, pluginWriter := setupMocks(t)
			defer ctrl.Finish()

			tc := testContext{
				tokenCache:    tokenCache,
				tokenProvider: tokenProvider,
				pluginWriter:  pluginWriter,
			}

			data.setupExpectations(tc)

			plugin := execCredentialPlugin{
				o:                    data.options,
				tokenCache:           tokenCache,
				provider:             tokenProvider,
				execCredentialWriter: pluginWriter,
			}

			ctx := context.TODO()
			errMessage := ""
			if err := plugin.Do(ctx); err != nil {
				errMessage = err.Error()
			}
			if errMessage != data.expectedError {
				t.Fatalf("expectedError: %s, actual: %s", data.expectedError, errMessage)
			}
		})
	}
}

func setupMocks(t *testing.T) (*gomock.Controller, *mock_token.MockTokenCache, *mock_token.MockTokenProvider, *mock_token.MockExecCredentialWriter) {
	ctrl := gomock.NewController(t)
	tokenCache := mock_token.NewMockTokenCache(ctrl)
	tokenProvider := mock_token.NewMockTokenProvider(ctrl)
	pluginWriter := mock_token.NewMockExecCredentialWriter(ctrl)

	return ctrl, tokenCache, tokenProvider, pluginWriter
}

func TestKUBERNETES_EXEC_INFOIsEmpty(t *testing.T) {
	testData := []struct {
		name            string
		execInfoEnvTest string
		options         Options
	}{
		{
			name:            "KUBERNETES_EXEC_INFO is empty",
			execInfoEnvTest: "",
			options: Options{
				LoginMethod: DeviceCodeLogin,
				ClientID:    "clientID",
				ServerID:    "serverID",
				TenantID:    "tenantID",
			},
		},
	}

	for _, data := range testData {
		t.Run(data.name, func(t *testing.T) {
			os.Setenv("KUBERNETES_EXEC_INFO", data.execInfoEnvTest)
			defer os.Unsetenv("KUBERNETES_EXEC_INFO")
			ecp, err := New(&data.options)
			if ecp == nil || err != nil {
				t.Fatalf("expected: return execCredentialPlugin and nil error, actual: did not return execCredentialPlugin or did not return expected error")
			}
		})
	}
}
