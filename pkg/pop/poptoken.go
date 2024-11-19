package pop

import "github.com/Azure/kubelogin/pkg/internal/pop"

func GetSwPoPKey() (*SwKey, error) {
	key, err := pop.GetSwPoPKey()
	if err != nil {
		return nil, err
	}
	return &SwKey{SwKey: *key}, nil
}
