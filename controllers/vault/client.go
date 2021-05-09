package vault

import (
	"fmt"
	vaultApi "github.com/hashicorp/vault/api"
	"os"
	"strconv"
	"time"
)

const (
	saJwtFile          = "/var/run/secrets/kubernetes.io/serviceaccount/token"
	K8sAuthBackendPath = "VAULT_K8S_AUTH_BACKEND_PATH"
	K8sAuthBackendRole = "VAULT_K8S_AUTH_BACKEND_ROLE"
)

type VClient struct {
	Client                   *vaultApi.Client
	LogicalClient            *vaultApi.Logical
	RoleLeaseDuration        int
	ExpectedLeaseToEndAtTime time.Time
}

func getVaultAuthInfo(providerCredentials map[string][]byte) (string, string, error) {
	authBackendPath, authBackendPathFound := providerCredentials[K8sAuthBackendPath]
	if !authBackendPathFound {
		return "", "", ErrProviderMissingVaultAuthBackendPath{
			Message: fmt.Sprintf("provider.credentials is missing %v key", K8sAuthBackendPath),
		}
	}
	authBackendRole, authBackendRoleFound := providerCredentials[K8sAuthBackendRole]
	if !authBackendRoleFound {
		return "", "", ErrProviderMissingAuthBackendRole{
			Message: fmt.Sprintf("provider.credentials is missing %v key", K8sAuthBackendRole),
		}
	}
	return string(authBackendPath), string(authBackendRole), nil
}

func NewVaultClient(providerCredentials map[string][]byte) (*VClient, error) {

	strAuthBackendPath, strAuthBackendRole, err := getVaultAuthInfo(providerCredentials)
	if err != nil {
		return nil, err
	}

	// 0. setup a vault client ( a non-authenticated one )
	vClient, errCreatingClient := vaultApi.NewClient(getVaultCfg(providerCredentials))
	if errCreatingClient != nil {
		return nil, errCreatingClient
	}

	// 1. get current pod SA jwt
	rawSAJWTFileContents, err := os.ReadFile(saJwtFile)
	if err != nil {
		return nil, err
	}
	strSAJwtFileContents := string(rawSAJWTFileContents)

	// 2. login with authBackendPath + authBackendRole
	vaultLoginData := map[string]interface{}{}
	vaultLoginData["role"] = strAuthBackendRole
	vaultLoginData["jwt"] = strSAJwtFileContents
	authBackendLoginPath := fmt.Sprintf("%s/login", strAuthBackendPath)
	loginSecret, err := vClient.Logical().Write(authBackendLoginPath, vaultLoginData)
	if err != nil {
		return nil, err
	}

	// update non-authenticated vault client with client token so its a authenticated vault client
	vClient.SetToken(loginSecret.Auth.ClientToken)

	c := &VClient{
		Client:                   vClient,
		LogicalClient:            vClient.Logical(),
		RoleLeaseDuration:        loginSecret.Auth.LeaseDuration,
		ExpectedLeaseToEndAtTime: time.Now().Add(time.Second * time.Duration(loginSecret.Auth.LeaseDuration)),
	}
	return c, nil

}

// getVaultCfg is a helper func that returns a vault config to setup a vault client
// If the input map has VAULT_ADDR and VAULT_SKIP_VERIFY, the default config is overridden with those values
func getVaultCfg(authInfo map[string][]byte) *vaultApi.Config {
	vaultCfg := vaultApi.DefaultConfig()
	if val, ok := authInfo[vaultApi.EnvVaultAddress]; ok {
		vaultCfg.Address = string(val)
	}
	if val, ok := authInfo[vaultApi.EnvVaultInsecure]; ok {
		b, _ := strconv.ParseBool(string(val))
		_ = vaultCfg.ConfigureTLS(&vaultApi.TLSConfig{Insecure: b})
	}
	return vaultCfg
}
