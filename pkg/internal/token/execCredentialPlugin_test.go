package token

import (
	"os"
	"testing"
)

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
