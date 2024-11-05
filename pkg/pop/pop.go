package pop

import "github.com/Azure/kubelogin/pkg/internal/pop"

func GetUniqueSwPoPKey() (*SwKey, error) {
	key, err := pop.GetUniqueSwPoPKey()
	if err != nil {
		return nil, err
	}
	return &SwKey{SwKey: *key}, nil
}
