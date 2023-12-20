package testutils

import (
	"net/http"
	"os"
	"strings"

	"gopkg.in/dnaeon/go-vcr.v3/cassette"
	"gopkg.in/dnaeon/go-vcr.v3/recorder"
)

const (
	tenantUUID        = "AZURE_TENANT_ID"
	vcrMode           = "VCR_MODE"
	vcrModeRecordOnly = "RecordOnly"
	redactionToken    = "[REDACTED]"
	testToken         = "TEST_ACCESS_TOKEN"
)

// GetVCRHttpClient setup Go-vcr
func GetVCRHttpClient(path string, token string) (*recorder.Recorder, *http.Client) {
	if len(path) == 0 || path == "" {
		return nil, nil
	}

	opts := &recorder.Options{
		CassetteName: path,
		Mode:         getVCRMode(),
	}
	rec, _ := recorder.NewWithOptions(opts)

	hook := func(i *cassette.Interaction) error {
		var detectedClientID, detectedClientSecret, detectedClientAssertion, detectedScope, detectedReqCnf string
		// Delete sensitive content
		delete(i.Response.Headers, "Set-Cookie")
		delete(i.Response.Headers, "X-Ms-Request-Id")
		if i.Request.Form["client_id"] != nil {
			detectedClientID = i.Request.Form["client_id"][0]
			i.Request.Form["client_id"] = []string{redactionToken}
		}
		if i.Request.Form["client_secret"] != nil && i.Request.Form["client_secret"][0] != BadSecret {
			detectedClientSecret = i.Request.Form["client_secret"][0]
			i.Request.Form["client_secret"] = []string{redactionToken}
		}
		if i.Request.Form["client_assertion"] != nil {
			detectedClientAssertion = i.Request.Form["client_assertion"][0]
			i.Request.Form["client_assertion"] = []string{redactionToken}
		}
		if i.Request.Form["scope"] != nil {
			detectedScope = i.Request.Form["scope"][0][:strings.IndexByte(i.Request.Form["scope"][0], '/')]
			i.Request.Form["scope"] = []string{redactionToken + "/.default openid offline_access profile"}
		}
		if i.Request.Form["req_cnf"] != nil {
			detectedScope = i.Request.Form["req_cnf"][0]
			i.Request.Form["req_cnf"] = []string{redactionToken}
		}

		if os.Getenv(tenantUUID) != "" {
			i.Request.URL = strings.ReplaceAll(i.Request.URL, os.Getenv(tenantUUID), tenantUUID)
			i.Response.Body = strings.ReplaceAll(i.Response.Body, os.Getenv(tenantUUID), tenantUUID)
		}

		if detectedClientID != "" {
			i.Request.Body = strings.ReplaceAll(i.Request.Body, detectedClientID, redactionToken)
		}
		if detectedClientSecret != "" {
			i.Request.Body = strings.ReplaceAll(i.Request.Body, detectedClientSecret, redactionToken)
		}
		if detectedClientAssertion != "" {
			i.Request.Body = strings.ReplaceAll(i.Request.Body, detectedClientAssertion, redactionToken)
		}
		if detectedScope != "" {
			i.Request.Body = strings.ReplaceAll(i.Request.Body, detectedScope, redactionToken)
		}
		if detectedReqCnf != "" {
			i.Request.Body = strings.ReplaceAll(i.Request.Body, detectedReqCnf, redactionToken)
		}

		if strings.Contains(i.Response.Body, "access_token") {
			i.Response.Body = `{"token_type":"Bearer","expires_in":86399,"ext_expires_in":86399,"access_token":"` + testToken + `"}`
		}

		if strings.Contains(i.Response.Body, "Invalid client secret provided") {
			i.Response.Body = `{"error":"invalid_client","error_description":"AADSTS7000215: Invalid client secret provided. Ensure the secret being sent in the request is the client secret value, not the client secret ID, for a secret added to app ''[REDACTED]''.\r\nTrace ID: [REDACTED]\r\nCorrelation ID: [REDACTED]\r\nTimestamp: 2023-06-02 21:00:26Z","error_codes":[7000215],"timestamp":"2023-06-02 21:00:26Z","trace_id":"[REDACTED]","correlation_id":"[REDACTED]","error_uri":"https://login.microsoftonline.com/error?code=7000215"}`
		}
		return nil
	}
	rec.AddHook(hook, recorder.BeforeSaveHook)

	playbackHook := func(i *cassette.Interaction) error {
		// Return a verifiable unique token on each test
		if strings.Contains(i.Response.Body, "access_token") {
			i.Response.Body = strings.ReplaceAll(i.Response.Body, testToken, token)
		}
		return nil
	}
	rec.AddHook(playbackHook, recorder.BeforeResponseReplayHook)

	rec.SetMatcher(customMatcher)
	rec.SetReplayableInteractions(true)

	return rec, rec.GetDefaultClient()
}

func customMatcher(r *http.Request, i cassette.Request) bool {
	id := os.Getenv(tenantUUID)
	if id == "" {
		id = "00000000-0000-0000-0000-000000000000"
	}
	switch os.Getenv(vcrMode) {
	case vcrModeRecordOnly:
	default:
		r.URL.Path = strings.ReplaceAll(r.URL.Path, id, tenantUUID)
	}
	return cassette.DefaultMatcher(r, i)
}

// Get go-vcr record mode from environment variable
func getVCRMode() recorder.Mode {
	switch os.Getenv(vcrMode) {
	case vcrModeRecordOnly:
		return recorder.ModeRecordOnly
	default:
		return recorder.ModeReplayOnly
	}
}
