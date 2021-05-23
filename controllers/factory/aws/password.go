package aws

import (
	"context"
	"errors"
	"fmt"
	"github.com/agill17/db-operator/api/v1alpha1"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/secretsmanager"
	v1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/json"
	"math/rand"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

const (
	DefaultKeyForPasswordInSecret = "password"
)

type SmStore struct {
	Password string `json:"password"`
}

func (i InternalAwsClients) GetOrSetMasterPassword(Obj client.Object, client client.Client, scheme *runtime.Scheme) (string, error) {
	useExistingPassword, secretName, secretNs, key, err := i.UseExistingSecretForPassword(Obj)
	if err != nil {
		return "", err
	}

	if useExistingPassword {
		existingSecret := &v1.Secret{}
		if err := client.Get(context.TODO(), types.NamespacedName{
			Namespace: secretNs,
			Name:      secretName,
		}, existingSecret); err != nil {
			return "", err
		}

		return string(existingSecret.Data[key]), nil
	}

	// else generate a secret and toss it in the creds in secrets manager
	generatedSecretName := fmt.Sprintf("%s-%s-master-password", Obj.GetNamespace(), Obj.GetName())
	generatedSecret := &v1.Secret{}
	if err := client.Get(context.TODO(), types.NamespacedName{
		Namespace: Obj.GetNamespace(),
		Name:      generatedSecretName,
	}, generatedSecret); err != nil {
		if apierrors.IsNotFound(err) {
			// generate password + store in SM + create secret
			generatedPassword := randStringRunes(16)
			_, err := i.StoreGeneratedPassword(generatedSecretName, generatedPassword, Obj.GetNamespace(), Obj, client, scheme)
			if err != nil {
				return "", err
			}
			return generatedPassword, nil
		}
		return "", err
	}
	return string(generatedSecret.Data[DefaultKeyForPasswordInSecret]), nil

	//TODO: move this to Is<TYPE>UpToDate func
	//// if the generated secret was already created ( by previous reconciles )
	//// ensure the password in secret matches the password in SM ( SM becomes the desired state )
	//generatedPasswordFromSM, err := i.smClient.GetSecretValue(&secretsmanager.GetSecretValueInput{
	//	SecretId:     aws.String(generatedSecretName),
	//})
	//if err != nil {
	//	return "", err
	//}
	//smStore := &SmStore{}
	//if err := json.Unmarshal([]byte(*generatedPasswordFromSM.SecretString), smStore); err != nil {
	//	return "", err
	//}
	//
	//// compare sm password with secretPassword
	//if smStore.Password != string(generatedSecret.Data[DefaultKeyForPasswordInSecret]) {
	//	generatedSecret.Data[DefaultKeyForPasswordInSecret] = []byte(smStore.Password)
	//	if err := client.Update(context.TODO(), generatedSecret); err != nil {
	//		return "", err
	//	}
	//}
	//
	//return smStore.Password, nil
}

// returns, shouldUseExistingSecret, secretName, secretNs, secretKey, error
func (i InternalAwsClients) UseExistingSecretForPassword(obj client.Object) (bool, string, string, string, error) {
	dbClusterCR, isDBCluster := obj.(*v1alpha1.DBCluster)
	_, isDBInstance := obj.(*v1alpha1.DBInstance)

	if !isDBCluster && !isDBInstance {
		return false, "", "", "", errors.New(fmt.Sprintf("Err%TPasswordRefIsNotYetSupported", obj))
	}

	if isDBCluster {
		if dbClusterCR.Spec.MasterUserPasswordSecretRef.PasswordKey != "" {
			return true, dbClusterCR.Spec.MasterUserPasswordSecretRef.SecretRef.Name,
				dbClusterCR.Spec.MasterUserPasswordSecretRef.SecretRef.Namespace,
				dbClusterCR.Spec.MasterUserPasswordSecretRef.PasswordKey, nil
		}
	}

	//TODO: check for DBInstance and return appropriate values

	return false, "", "", "", nil
}

func (i InternalAwsClients) StoreGeneratedPassword(secretName, secretVal, secretNs string, owner metav1.Object, client client.Client, scheme *runtime.Scheme) (*v1.Secret, error) {
	sm := SmStore{Password: secretVal}
	jsonSMRaw, err := json.Marshal(sm)
	if err != nil {
		return nil, err
	}
	jsonSMStr := string(jsonSMRaw)
	_, errCreatingSM := i.smClient.CreateSecret(&secretsmanager.CreateSecretInput{
		Name:         aws.String(secretName),
		SecretString: aws.String(jsonSMStr),
	})
	if errCreatingSM != nil {
		return nil, errCreatingSM
	}

	secret := &v1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      secretName,
			Namespace: secretNs,
		},
		Data: map[string][]byte{
			DefaultKeyForPasswordInSecret: []byte(secretVal),
		},
		Type: v1.SecretTypeOpaque,
	}
	_, errCreatingSecret := controllerutil.CreateOrUpdate(context.TODO(), client, secret, func() error {
		return controllerutil.SetOwnerReference(owner, secret, scheme)
	})
	return nil, errCreatingSecret
}

var letterRunes = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")

func randStringRunes(n int) string {
	b := make([]rune, n)
	for i := range b {
		b[i] = letterRunes[rand.Intn(len(letterRunes))]
	}
	return string(b)
}
