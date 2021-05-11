package aws

import (
	"fmt"
	vault2 "github.com/agill17/db-operator/controllers/vault"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/credentials/stscreds"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/rds"
	"github.com/aws/aws-sdk-go/service/rds/rdsiface"
	"github.com/aws/aws-sdk-go/service/secretsmanager"
	"github.com/aws/aws-sdk-go/service/secretsmanager/secretsmanageriface"
	"math"
	"os"
	"strings"
	"sync"
)

const (
	MockAwsEndpoint = "MOCK_AWS_ENDPOINT"
)

type InternalAwsClients struct {
	rdsClient    rdsiface.RDSAPI
	smClient     secretsmanageriface.SecretsManagerAPI
	cacheKeyName string
	creds        *credentials.Credentials
}

const (
	VaultRegexp        = `^vault[a-zA-Z0-9\W].*#[a-zA-Z0-9\W].*$`
	AccessKeyIdVar     = "AWS_ACCESS_KEY_ID"
	SecretAccessKeyVar = "AWS_SECRET_ACCESS_KEY"
	RoleArnVar         = "AWS_ROLE_ARN"
)

var awsClientCache sync.Map

func getAwsClientCacheKey(pName, region string) string {
	return fmt.Sprintf("aws-%v-%v", pName, region)
}

func NewInternalAwsClient(region string, pName string, providerCredentials map[string][]byte) (*InternalAwsClients, error) {
	cacheKeyName := getAwsClientCacheKey(pName, region)
	cachedInternalAwsClient, ok := awsClientCache.Load(cacheKeyName)
	if !ok {
		sess := session.Must(session.NewSession())
		creds, err := getAwsCredentials(providerCredentials)
		if err != nil {
			return nil, err
		}
		awsClientCfg := &aws.Config{
			CredentialsChainVerboseErrors: aws.Bool(true),
			Region:                        aws.String(region),
			Credentials:                   creds,
			MaxRetries:                    aws.Int(math.MaxInt32),
		}

		if val, ok := os.LookupEnv(MockAwsEndpoint); ok {
			awsClientCfg.Endpoint = aws.String(val)
		}

		r := &InternalAwsClients{
			rdsClient:    rds.New(sess, awsClientCfg),
			smClient:     secretsmanager.New(sess, awsClientCfg),
			cacheKeyName: cacheKeyName,
			creds:        creds,
		}
		// cache key is deleted when access key id is no longer valid
		awsClientCache.Store(cacheKeyName, r)
		return r, nil
	}
	if cachedInternalAwsClient.(*InternalAwsClients).creds.IsExpired() {
		awsClientCache.Delete(cacheKeyName)
		return nil, ErrRequeueNeeded{Message: "AWS Credentials expired, requeue needed."}
	}
	return cachedInternalAwsClient.(*InternalAwsClients), nil
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

/*
	TODO: revisit,redesign : tabling this for now because
	vault api does not provide a simple way
	to find if a path is kv engine or aws secrets engine
*/
func getCredsFromVault(providerCredentials map[string][]byte, strAccessKey, strSecretKey string) (string, string, error) {
	vClient, err := vault2.NewVaultClient(providerCredentials)
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
