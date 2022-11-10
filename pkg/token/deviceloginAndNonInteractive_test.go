package token

import (
	"os"
	"testing"
)

func TestDeviceloginAndNonInteractive(t *testing.T) {
	testData := []struct {
		name            string
		execInfoEnvTest string
		options         Options
		expectedError   string
	}{
		{
			name:            "KUBERNETES_EXEC_INFO.spec.interactive: false and login mode is devicelogin",
			execInfoEnvTest: `{"kind":"ExecCredential","apiVersion":"client.authentication.k8s.io/v1beta1","spec":{"interactive":false}}`,
			options: Options{
				LoginMethod: DeviceCodeLogin,
			},
			expectedError: "devicelogin is not supported if interactiveMode is 'never'",
		},
	}

	for _, data := range testData {
		t.Run(data.name, func(t *testing.T) {
			os.Setenv("KUBERNETES_EXEC_INFO", data.execInfoEnvTest)
			defer os.Unsetenv("KUBERNETES_EXEC_INFO")
			ecp, err := New(&data.options)
			if ecp != nil || err == nil || err.Error() != data.expectedError {
				t.Fatalf("expected: return defined error, actual: did not return expected error")
			}
		})
	}
}
