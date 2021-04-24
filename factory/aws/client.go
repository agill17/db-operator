package aws

import (
	"fmt"
	"github.com/agill17/db-operator/vault"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/rds"
	"github.com/aws/aws-sdk-go/service/rds/rdsiface"
	"os"
	"regexp"
	"strings"
	"sync"
)

const (
	EnvMockRDSEndpoint = "MOCK_RDS_ENDPOINT"
)

type RDSClient struct {
	rdsClient rdsiface.RDSAPI
	cacheKeyName string
}

const (
	VaultRegexp        = `^vault[a-zA-Z0-9\W].*#[a-zA-Z0-9\W].*$`
	AccessKeyIdVar     = "AWS_ACCESS_KEY_ID"
	SecretAccessKeyVar = "AWS_SECRET_ACCESS_KEY"

)

var rdsClientCache sync.Map

func getRDSClientCacheKey(pName, region string) string {
	return fmt.Sprintf("aws-%v-%v", pName, region)
}

func NewRDSClient(region string, pName string, providerCredentials map[string][]byte) (*RDSClient, error) {
	cacheKeyName := getRDSClientCacheKey(pName, region)
	cachedRDSClient, ok := rdsClientCache.Load(cacheKeyName)
	if !ok {
		sess := session.Must(session.NewSession())
		creds, err := getAwsCredentials(providerCredentials)
		if err != nil {
			return nil, err
		}
		rdsClientCfg := &aws.Config{
			CredentialsChainVerboseErrors:     aws.Bool(true),
			Region: aws.String(region),
			Credentials: creds,
		}

		if val, ok := os.LookupEnv(EnvMockRDSEndpoint); ok {
			rdsClientCfg.Endpoint = aws.String(val)
		}
		r := &RDSClient{
			rdsClient:    rds.New(sess, rdsClientCfg),
			cacheKeyName: cacheKeyName,
		}
		// cache key is deleted when access key id is no longer valid
		rdsClientCache.Store(cacheKeyName, r)
		return r, nil
	}
	return cachedRDSClient.(*RDSClient), nil
}

func getAwsCredentials(providerCredentials map[string][]byte) (*credentials.Credentials, error) {
	accessKey, accessKeyFound := providerCredentials[AccessKeyIdVar]
	if !accessKeyFound {
		return nil, ErrorProviderMissingAwsAccessKeyID{Message: fmt.Sprintf("provider.credentials is missing %v key", AccessKeyIdVar)}
	}
	strAccessKey := string(accessKey)
	secretKey, secretKeyFound := providerCredentials[SecretAccessKeyVar]
	if !secretKeyFound {
		return nil, ErrorProviderMissingAwsSecretAccessKey{Message: fmt.Sprintf("provider.credentials is missing %v key", SecretAccessKeyVar)}
	}
	strSecretKey := string(secretKey)

	vaultRegexp, err := regexp.Compile(VaultRegexp)
	if err != nil {
		return nil, err
	}

	if vaultRegexp.MatchString(strAccessKey) &&
		vaultRegexp.MatchString(strSecretKey) {

		vClient, err := vault.NewVaultClient(providerCredentials)
		if err != nil {
			return nil, err
		}

		// sample vault path
		// vault:kv/data/foo#key
		// remove vault prefix
		strAccessIdPath := strings.Replace(strAccessKey, "vault:", "", 1)
		strSecretKeyPath := strings.Replace(strSecretKey, "vault:", "", 1)

		// separate path and key by splitting at #
		accessIDPathKeySplit := strings.Split(strAccessIdPath, "#")
		secretKeyPathSplit := strings.Split(strSecretKeyPath, "#")

		accessKeyFromVault, err := vClient.ReadVaultSecretPath(accessIDPathKeySplit[0], accessIDPathKeySplit[1])
		if err != nil {
			return nil, err
		}
		secretAccessKeyFromVault, err := vClient.ReadVaultSecretPath(secretKeyPathSplit[0], secretKeyPathSplit[1])
		if err != nil {
			return nil, err
		}

		return credentials.NewStaticCredentials(accessKeyFromVault, secretAccessKeyFromVault, ""), nil
	}

	return credentials.NewStaticCredentials(strAccessKey, strSecretKey, ""), nil

}


