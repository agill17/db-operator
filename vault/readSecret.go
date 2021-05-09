package vault

import (
	"errors"
	"fmt"
	"github.com/spf13/cast"
)

func (v VClient) ReadVaultSecretPath(path string, key string) (string, error) {
	secret, err := v.LogicalClient.ReadWithData(path, map[string][]string{"version": {"-1"}})
	if err != nil {
		return "", err
	}

	secretData := secret.Data["data"]
	secretStringMap := cast.ToStringMap(secretData)
	val, found := secretStringMap[key]
	if !found {
		return "", errors.New("ErrVaultKeyNotFound")
	}

	return fmt.Sprint(val), nil
}
