package aws

import (
	"fmt"
	"github.com/agill17/db-operator/vault"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/credentials/stscreds"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/rds"
	"github.com/aws/aws-sdk-go/service/rds/rdsiface"
	"math"
	"os"
	"strings"
	"sync"
)

const (
	EnvMockRDSEndpoint = "MOCK_RDS_ENDPOINT"
)

type RDSClient struct {
	rdsClient    rdsiface.RDSAPI
	cacheKeyName string
	creds        *credentials.Credentials
}

const (
	VaultRegexp        = `^vault[a-zA-Z0-9\W].*#[a-zA-Z0-9\W].*$`
	AccessKeyIdVar     = "AWS_ACCESS_KEY_ID"
	SecretAccessKeyVar = "AWS_SECRET_ACCESS_KEY"
	RoleArnVar         = "AWS_ROLE_ARN"
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
			CredentialsChainVerboseErrors: aws.Bool(true),
			Region:                        aws.String(region),
			Credentials:                   creds,
			MaxRetries:                    aws.Int(math.MaxInt32),
		}

		if val, ok := os.LookupEnv(EnvMockRDSEndpoint); ok {
			rdsClientCfg.Endpoint = aws.String(val)
		}
		r := &RDSClient{
			rdsClient:    rds.New(sess, rdsClientCfg),
			cacheKeyName: cacheKeyName,
			creds:        creds,
		}
		// cache key is deleted when access key id is no longer valid
		rdsClientCache.Store(cacheKeyName, r)
		return r, nil
	}
	if cachedRDSClient.(*RDSClient).creds.IsExpired() {
		rdsClientCache.Delete(cacheKeyName)
		return nil, ErrRequeueNeeded{Message: "AWS Credentials expired, requeue needed."}
	}
	return cachedRDSClient.(*RDSClient), nil
}

/**
Precedence
1. AWS_ROLE_ARN -- assuming the underlying pod has permissions to assume another role
2. Static creds ( AWS_ACCESS_KEY_ID and AWS_SECRET_ACCESS_KEY )
*/
func getAwsCredentials(providerCredentials map[string][]byte) (*credentials.Credentials, error) {
	roleArn, hasRoleArn := providerCredentials[RoleArnVar]
	if hasRoleArn {
		sess := session.Must(session.NewSession())
		return stscreds.NewCredentials(sess, string(roleArn)), nil
	}
	accessKeyId, hasAccessKeyId := providerCredentials[AccessKeyIdVar]
	if !hasAccessKeyId {
		return nil, ErrorProviderMissingAwsAccessKeyID{Message: "AWS provider credentials missing access key id"}
	}
	secretAccessKey, hasSecretKey := providerCredentials[SecretAccessKeyVar]
	if !hasSecretKey {
		return nil, ErrorProviderMissingAwsSecretAccessKey{Message: "AWS provider credentials missing seceret access key"}
	}
	return credentials.NewStaticCredentials(string(accessKeyId), string(secretAccessKey), ""), nil

}

// TODO: tabling this for now
func getCredsFromVault(providerCredentials map[string][]byte, strAccessKey, strSecretKey string) (string, string, error) {
	vClient, err := vault.NewVaultClient(providerCredentials)
	if err != nil {
		return "", "", err
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
		return "", "", err
	}
	secretAccessKeyFromVault, err := vClient.ReadVaultSecretPath(secretKeyPathSplit[0], secretKeyPathSplit[1])
	if err != nil {
		return "", "", err
	}

	return accessKeyFromVault, secretAccessKeyFromVault, nil
}
