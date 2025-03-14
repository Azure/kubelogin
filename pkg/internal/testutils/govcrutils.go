package testutils

import (
	"encoding/json"
	"net/http"
	"regexp"
	"strings"

	"gopkg.in/dnaeon/go-vcr.v4/pkg/cassette"
	"gopkg.in/dnaeon/go-vcr.v4/pkg/recorder"
)

const (
	redactedToken = "[REDACTED]"
	TestToken     = "TEST_ACCESS_TOKEN"
	TestUsername  = "user@example.com"
	TestTenantID  = "00000000-0000-0000-0000-000000000000"
	TestClientID  = "80faf920-1908-4b52-b5ef-a8e7bedfc67a"
	TestServerID  = "6dae42f8-4368-4678-94ff-3960e28e3630"
)

const (
	mockClientInfo = "eyJ1aWQiOiJjNzNjNmYyOC1hZTVmLTQxM2QtYTlhMi1lMTFlNWFmNjY4ZjgiLCJ1dGlkIjoiZTBiZDIzMjEtMDdmYS00Y2YwLTg3YjgtMDBhYTJhNzQ3MzI5In0"
	mockIDT        = "eyJ0eXAiOiJKV1QiLCJhbGciOiJSUzI1NiIsImtpZCI6Imwzc1EtNTBjQ0g0eEJWWkxIVEd3blNSNzY4MCJ9.eyJhdWQiOiIwNGIwNzc5NS04ZGRiLTQ2MWEtYmJlZS0wMmY5ZTFiZjdiNDYiLCJpc3MiOiJodHRwczovL2xvZ2luLm1pY3Jvc29mdG9ubGluZS5jb20vYzU0ZmFjODgtM2RkMy00NjFmLWE3YzQtOGEzNjhlMDM0MGIzL3YyLjAiLCJpYXQiOjE2MzcxOTEyMTIsIm5iZiI6MTYzNzE5MTIxMiwiZXhwIjoxNjM3MTk1MTEyLCJhaW8iOiJBVVFBdS84VEFBQUFQMExOZGNRUXQxNmJoSkFreXlBdjFoUGJuQVhtT0o3RXJDVHV4N0hNTjhHd2VMb2FYMWR1cDJhQ2Y0a0p5bDFzNmovSzF5R05DZmVIQlBXM21QUWlDdz09IiwiaWRwIjoiaHR0cHM6Ly9zdHMud2luZG93cy5uZXQvZTBiZDIzMjEtMDdmYS00Y2YwLTg3YjgtMDBhYTJhNzQ3MzI5LyIsIm5hbWUiOiJJZGVudGl0eSBUZXN0IFVzZXIiLCJwcmVmZXJyZWRfdXNlcm5hbWUiOiJpZGVudGl0eXRlc3R1c2VyQGF6dXJlc2Rrb3V0bG9vay5vbm1pY3Jvc29mdC5jb20iLCJyaCI6IjAuQVMwQWlLeFB4ZE05SDBhbnhJbzJqZ05BczVWM3NBVGJqUnBHdS00Qy1lR19lMFl0QUxFLiIsInN1YiI6ImMxYTBsY2xtbWxCYW9wc0MwVmlaLVpPMjFCT2dSUXE3SG9HRUtOOXloZnMiLCJ0aWQiOiJjNTRmYWM4OC0zZGQzLTQ2MWYtYTdjNC04YTM2OGUwMzQwYjMiLCJ1dGkiOiI5TXFOSWI5WjdrQy1QVHRtai11X0FBIiwidmVyIjoiMi4wIn0.hh5Exz9MBjTXrTuTZnz7vceiuQjcC_oRSTeBIC9tYgSO2c2sqQRpZi91qBZFQD9okayLPPKcwqXgEJD9p0-c4nUR5UQN7YSeDLmYtZUYMG79EsA7IMiQaiy94AyIe2E-oBDcLwFycGwh1iIOwwOwjbanmu2Dx3HfQx831lH9uVjagf0Aow0wTkTVCsedGSZvG-cRUceFLj-kFN-feFH3NuScuOfLR2Magf541pJda7X7oStwL_RNUFqjJFTdsiFV4e-VHK5qo--3oPU06z0rS9bosj0pFSATIVHrrS4gY7jiSvgMbG837CDBQkz5b08GUN5GlLN9jlygl1plBmbgww"
)

var emailRegex = regexp.MustCompile(`[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}`)

func GetVCRHttpClient(path, tenantID string) (*recorder.Recorder, error) {
	deviceCodePendingCount := 0
	beforeSaveHook := func(i *cassette.Interaction) error {
		// in device code login, since the client polls for the completion of the login
		// we only record it once to speed up the replay
		if strings.Contains(i.Response.Body, "AADSTS70016") {
			if deviceCodePendingCount > 0 {
				i.DiscardOnSave = true
				return nil
			}
			deviceCodePendingCount++
		}
		var detectedClientID,
			detectedClientSecret,
			detectedClientAssertion,
			detectedScope,
			detectedReqCnf,
			detectedPassword,
			detectedUsername,
			detectedDeviceCode string
		// Delete sensitive content
		delete(i.Response.Headers, "Set-Cookie")
		delete(i.Response.Headers, "X-Ms-Request-Id")
		if i.Request.Form["client_id"] != nil {
			detectedClientID = i.Request.Form["client_id"][0]
			i.Request.Form["client_id"] = []string{redactedToken}
		}
		if i.Request.Form["client_secret"] != nil && i.Request.Form["client_secret"][0] != BadSecret {
			detectedClientSecret = i.Request.Form["client_secret"][0]
			i.Request.Form["client_secret"] = []string{redactedToken}
		}
		if i.Request.Form["client_assertion"] != nil {
			detectedClientAssertion = i.Request.Form["client_assertion"][0]
			i.Request.Form["client_assertion"] = []string{redactedToken}
		}
		if i.Request.Form["req_cnf"] != nil {
			detectedScope = i.Request.Form["req_cnf"][0]
			i.Request.Form["req_cnf"] = []string{redactedToken}
		}
		if i.Request.Form["password"] != nil && i.Request.Form["password"][0] != BadSecret {
			detectedPassword = i.Request.Form["password"][0]
			i.Request.Form["password"] = []string{redactedToken}
		}
		if i.Request.Form["username"] != nil {
			detectedUsername = i.Request.Form["username"][0]
			i.Request.Form["username"] = []string{redactedToken}
		}
		if i.Request.Form["device_code"] != nil {
			detectedDeviceCode = i.Request.Form["device_code"][0]
			i.Request.Form["device_code"] = []string{redactedToken}
		}

		i.Request.URL = redactURL(i.Request.URL, tenantID)
		i.Response.Body = strings.ReplaceAll(i.Response.Body, tenantID, TestTenantID)

		if detectedClientID != "" {
			i.Request.Body = strings.ReplaceAll(i.Request.Body, detectedClientID, redactedToken)
		}
		if detectedClientSecret != "" {
			i.Request.Body = ReplaceSecretValuesIncludingURLEscaped(i.Request.Body, detectedClientSecret, redactedToken)
		}
		if detectedClientAssertion != "" {
			i.Request.Body = strings.ReplaceAll(i.Request.Body, detectedClientAssertion, redactedToken)
		}
		if detectedScope != "" {
			i.Request.Body = strings.ReplaceAll(i.Request.Body, detectedScope, redactedToken)
		}
		if detectedReqCnf != "" {
			i.Request.Body = strings.ReplaceAll(i.Request.Body, detectedReqCnf, redactedToken)
		}
		if detectedPassword != "" {
			i.Request.Body = ReplaceSecretValuesIncludingURLEscaped(i.Request.Body, detectedPassword, redactedToken)
		}
		if detectedUsername != "" {
			i.Request.Body = ReplaceSecretValuesIncludingURLEscaped(i.Request.Body, detectedUsername, TestUsername)
			i.Request.URL = ReplaceSecretValuesIncludingURLEscaped(i.Request.URL, detectedUsername, TestUsername)
		}
		if detectedDeviceCode != "" {
			i.Request.Body = strings.ReplaceAll(i.Request.Body, detectedDeviceCode, redactedToken)
		}

		if strings.Contains(i.Response.Body, "access_token") || strings.Contains(i.Response.Body, "device_code") {
			redacted, err := redactToken(i.Response.Body)
			if err != nil {
				return err
			}
			i.Response.Body = redacted
		}

		if strings.Contains(i.Response.Body, "Invalid client secret provided") {
			i.Response.Body = `{"error":"invalid_client","error_description":"AADSTS7000215: Invalid client secret provided. Ensure the secret being sent in the request is the client secret value, not the client secret ID, for a secret added to app ''[REDACTED]''.\r\nTrace ID: [REDACTED]\r\nCorrelation ID: [REDACTED]\r\nTimestamp: 2023-06-02 21:00:26Z","error_codes":[7000215],"timestamp":"2023-06-02 21:00:26Z","trace_id":"[REDACTED]","correlation_id":"[REDACTED]","error_uri":"https://login.microsoftonline.com/error?code=7000215"}`
		}
		return nil
	}

	playbackHook := func(i *cassette.Interaction) error {
		if strings.Contains(i.Response.Body, "access_token") {
			redacted, err := redactToken(i.Response.Body)
			if err != nil {
				return err
			}
			i.Response.Body = redacted
		}
		return nil
	}

	matcher := func(r *http.Request, i cassette.Request) bool {
		url := redactURL(r.URL.String(), tenantID)
		if r.Method != i.Method || url != i.URL {
			return false
		}
		_ = r.ParseForm()
		requestFormValues := r.Form
		isPop := i.Form["token_type"] != nil && i.Form["token_type"][0] == "pop"

		for k, v := range i.Form {
			if requestFormValues[k][0] != v[0] {
				// if recorded value is redaction token and request value is empty, then it is a mismatch
				if v[0] == redactedToken {
					if len(requestFormValues[k][0]) == 0 {
						return false
					}
					continue
				}
				// saml assertion is not relevant for the test
				if isPop && k == "assertion" {
					continue
				}
				return false
			}
		}

		return true
	}

	recOpts := []recorder.Option{
		recorder.WithHook(beforeSaveHook, recorder.BeforeSaveHook),
		recorder.WithHook(playbackHook, recorder.BeforeResponseReplayHook),
		recorder.WithMatcher(matcher),
		recorder.WithSkipRequestLatency(true),
	}

	return recorder.New(path, recOpts...)
}

func redactURL(url, tenantID string) string {
	if strings.Contains(url, "UserRealm") {
		url = emailRegex.ReplaceAllString(url, TestUsername)
	}
	return strings.ReplaceAll(url, tenantID, TestTenantID)
}

func redactToken(body string) (string, error) {
	var data map[string]interface{}
	err := json.Unmarshal([]byte(body), &data)
	if err != nil {
		return "", err
	}

	if _, ok := data["access_token"]; ok {
		data["access_token"] = TestToken
	}

	if _, ok := data["refresh_token"]; ok {
		data["refresh_token"] = TestToken
	}

	if _, ok := data["id_token"]; ok {
		data["id_token"] = mockIDT
	}

	if _, ok := data["client_info"]; ok {
		data["client_info"] = mockClientInfo
	}

	if _, ok := data["device_code"]; ok {
		data["device_code"] = redactedToken
	}

	// Marshal the map back to a JSON string
	redactedJSON, err := json.Marshal(data)
	if err != nil {
		return "", err
	}

	return string(redactedJSON), nil
}
