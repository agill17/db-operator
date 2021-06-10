package controllers

import (
	"context"
	"github.com/agill17/db-operator/api/v1alpha1"
	"github.com/agill17/db-operator/pkg/factory"
	aws2 "github.com/agill17/db-operator/pkg/factory/aws"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/rds"
	"github.com/go-logr/logr"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	"reflect"
	controllerruntime "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"testing"
	"time"
)

func TestDBClusterReconciler_Reconcile(t *testing.T) {
	testScheme := runtime.NewScheme()
	clientgoscheme.AddToScheme(testScheme)
	v1alpha1.AddToScheme(testScheme)
	timeNow := time.Now()

	type fields struct {
		Client           client.Client
		Log              logr.Logger
		Scheme           *runtime.Scheme
		CloudDBInterface factory.CloudDB
	}
	type args struct {
		ctx context.Context
		req controllerruntime.Request
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    controllerruntime.Result
		wantErr bool
	}{
		{
			name:    "Test-AWS DBluster - when cluster does not exist, should requeue after creating",
			want:    controllerruntime.Result{Requeue: true},
			wantErr: false,
			args: args{
				ctx: context.Background(),
				req: controllerruntime.Request{NamespacedName: types.NamespacedName{
					Namespace: "default",
					Name:      "aws-db-cluster",
				}},
			},
			fields: fields{
				Client: fake.NewFakeClientWithScheme(testScheme, &v1alpha1.DBCluster{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "aws-db-cluster",
						Namespace: "default",
					},
					Spec: v1alpha1.DBClusterSpec{
						Provider: v1alpha1.Provider{
							Type: "aws",
							SecretRef: v1.SecretReference{
								Name:      "aws-provider-secret",
								Namespace: "default",
							},
						},
						Region:            "us-east-1",
						AvailabilityZones: []string{"us-east-1a", "us-east-1b"},
						DatabaseName:      "test",
						Engine:            "aurora-mysql",
						EngineMode:        "provisioned",
						EngineVersion:     "5.7.12",
						MasterUsername:    "test",
						MasterUserPasswordSecretRef: v1alpha1.MasterUserPasswordSecretRef{
							PasswordKey: "password",
							SecretRef: v1.SecretReference{
								Name:      "dbcluster-password",
								Namespace: "default",
							},
						},
						DBClusterParameterGroupName: "default-aurora-mysql5.7",
					},
				}, &v1.Secret{
					ObjectMeta: metav1.ObjectMeta{
						Namespace: "default",
						Name:      "aws-provider-secret",
					},
					Data: map[string][]byte{
						"AWS_ACCESS_KEY_ID":     []byte("fake-id"),
						"AWS_SECRET_ACCESS_KEY": []byte("fake-access-key"),
					},
					Type: v1.SecretTypeOpaque,
				}, &v1.Secret{
					ObjectMeta: metav1.ObjectMeta{
						Namespace: "default",
						Name:      "dbcluster-password",
					},
					Data: map[string][]byte{
						"password": []byte("test"),
					},
					Type: v1.SecretTypeOpaque,
				}),
				Log:    logf.Log,
				Scheme: testScheme,
				CloudDBInterface: &factory.MockCloudDB{
					IsDBClusterUpToDateResp:     true,
					IsDBClusterUpToDateModifyIn: &rds.ModifyDBClusterInput{},
					DBClusterExistsResp:         false,
				},
			},
		},
		{
			name:    "Test-AWS DBluster - when cluster does exist and is up to date - it should not requeue",
			want:    controllerruntime.Result{},
			wantErr: false,
			args: args{
				ctx: context.Background(),
				req: controllerruntime.Request{NamespacedName: types.NamespacedName{
					Namespace: "default",
					Name:      "aws-db-cluster",
				}},
			},
			fields: fields{
				Client: fake.NewFakeClientWithScheme(testScheme, &v1alpha1.DBCluster{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "aws-db-cluster",
						Namespace: "default",
					},
					Spec: v1alpha1.DBClusterSpec{
						Provider: v1alpha1.Provider{
							Type: "aws",
							SecretRef: v1.SecretReference{
								Name:      "aws-provider-secret",
								Namespace: "default",
							},
						},
						Region:            "us-east-1",
						AvailabilityZones: []string{"us-east-1a", "us-east-1b"},
						DatabaseName:      "test",
						Engine:            "aurora-mysql",
						EngineMode:        "provisioned",
						EngineVersion:     "5.7.12",
						MasterUsername:    "test",
						MasterUserPasswordSecretRef: v1alpha1.MasterUserPasswordSecretRef{
							PasswordKey: "password",
							SecretRef: v1.SecretReference{
								Name:      "dbcluster-password",
								Namespace: "default",
							},
						},
						DBClusterParameterGroupName: "default-aurora-mysql5.7",
					},
					Status: v1alpha1.DBClusterStatus{
						Phase:                   v1alpha1.Available,
						SecretsManagerVersionID: "",
					},
				}, &v1.Secret{
					ObjectMeta: metav1.ObjectMeta{
						Namespace: "default",
						Name:      "aws-provider-secret",
					},
					Data: map[string][]byte{
						"AWS_ACCESS_KEY_ID":     []byte("fake-id"),
						"AWS_SECRET_ACCESS_KEY": []byte("fake-access-key"),
					},
					Type: v1.SecretTypeOpaque,
				}, &v1.Secret{
					ObjectMeta: metav1.ObjectMeta{
						Namespace: "default",
						Name:      "dbcluster-password",
					},
					Data: map[string][]byte{
						"password": []byte("test"),
					},
					Type: v1.SecretTypeOpaque,
				}),
				Log:    logf.Log,
				Scheme: testScheme,
				CloudDBInterface: &factory.MockCloudDB{
					IsDBClusterUpToDateResp:     true,
					IsDBClusterUpToDateModifyIn: &rds.ModifyDBClusterInput{},
					DBClusterExistsResp:         true,
					DBClusterExistsStatus:       string(v1alpha1.Available),
				},
			},
		},
		{
			name:    "Test-AWS DBluster - when cluster does exist and is NOT up to date - it should modify and requeue",
			want:    controllerruntime.Result{Requeue: true},
			wantErr: false,
			args: args{
				ctx: context.Background(),
				req: controllerruntime.Request{NamespacedName: types.NamespacedName{
					Namespace: "default",
					Name:      "aws-db-cluster",
				}},
			},
			fields: fields{
				Client: fake.NewFakeClientWithScheme(testScheme, &v1alpha1.DBCluster{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "aws-db-cluster",
						Namespace: "default",
					},
					Spec: v1alpha1.DBClusterSpec{
						Provider: v1alpha1.Provider{
							Type: "aws",
							SecretRef: v1.SecretReference{
								Name:      "aws-provider-secret",
								Namespace: "default",
							},
						},
						Region:            "us-east-1",
						AvailabilityZones: []string{"us-east-1a", "us-east-1b"},
						DatabaseName:      "test",
						Engine:            "aurora-mysql",
						EngineMode:        "provisioned",
						EngineVersion:     "5.7.12",
						MasterUsername:    "test",
						MasterUserPasswordSecretRef: v1alpha1.MasterUserPasswordSecretRef{
							PasswordKey: "password",
							SecretRef: v1.SecretReference{
								Name:      "dbcluster-password",
								Namespace: "default",
							},
						},
						DBClusterParameterGroupName: "default-aurora-mysql5.7",
					},
					Status: v1alpha1.DBClusterStatus{
						Phase:                   v1alpha1.Available,
						SecretsManagerVersionID: "",
					},
				}, &v1.Secret{
					ObjectMeta: metav1.ObjectMeta{
						Namespace: "default",
						Name:      "aws-provider-secret",
					},
					Data: map[string][]byte{
						"AWS_ACCESS_KEY_ID":     []byte("fake-id"),
						"AWS_SECRET_ACCESS_KEY": []byte("fake-access-key"),
					},
					Type: v1.SecretTypeOpaque,
				}, &v1.Secret{
					ObjectMeta: metav1.ObjectMeta{
						Namespace: "default",
						Name:      "dbcluster-password",
					},
					Data: map[string][]byte{
						"password": []byte("test"),
					},
					Type: v1.SecretTypeOpaque,
				}),
				Log:    logf.Log,
				Scheme: testScheme,
				CloudDBInterface: &factory.MockCloudDB{
					IsDBClusterUpToDateResp: false,
					IsDBClusterUpToDateModifyIn: &rds.ModifyDBClusterInput{
						DeletionProtection:  aws.Bool(true),
						DBClusterIdentifier: aws.String("default-aws-db-cluster"),
					},
					DBClusterExistsResp:   true,
					DBClusterExistsStatus: string(v1alpha1.Available),
				},
			},
		},
		{
			name:    "Test-AWS DBluster - when dbcluster object does not exist - do not requeue, no error",
			want:    controllerruntime.Result{Requeue: false},
			wantErr: false,
			args: args{
				ctx: context.Background(),
				req: controllerruntime.Request{NamespacedName: types.NamespacedName{
					Namespace: "default",
					Name:      "aws-db-cluster",
				}},
			},
			fields: fields{
				Client: fake.NewFakeClientWithScheme(testScheme),
				Log:    logf.Log,
				Scheme: testScheme,
				CloudDBInterface: &factory.MockCloudDB{
					IsDBClusterUpToDateResp: false,
					IsDBClusterUpToDateModifyIn: &rds.ModifyDBClusterInput{
						DeletionProtection:  aws.Bool(true),
						DBClusterIdentifier: aws.String("default-aws-db-cluster"),
					},
					DBClusterExistsResp:   true,
					DBClusterExistsStatus: string(v1alpha1.Available),
				},
			},
		},
		{
			name:    "Test-AWS DBluster - when provider secrets do not exist",
			want:    controllerruntime.Result{},
			wantErr: true,
			args: args{
				ctx: context.Background(),
				req: controllerruntime.Request{NamespacedName: types.NamespacedName{
					Namespace: "default",
					Name:      "aws-db-cluster",
				}},
			},
			fields: fields{
				Client: fake.NewFakeClientWithScheme(testScheme, &v1alpha1.DBCluster{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "aws-db-cluster",
						Namespace: "default",
					},
					Spec: v1alpha1.DBClusterSpec{
						Provider: v1alpha1.Provider{
							Type: "aws",
							SecretRef: v1.SecretReference{
								Name:      "aws-provider-secret",
								Namespace: "default",
							},
						},
						Region:            "us-east-1",
						AvailabilityZones: []string{"us-east-1a", "us-east-1b"},
						DatabaseName:      "test",
						Engine:            "aurora-mysql",
						EngineMode:        "provisioned",
						EngineVersion:     "5.7.12",
						MasterUsername:    "test",
						MasterUserPasswordSecretRef: v1alpha1.MasterUserPasswordSecretRef{
							PasswordKey: "password",
							SecretRef: v1.SecretReference{
								Name:      "dbcluster-password",
								Namespace: "default",
							},
						},
						DBClusterParameterGroupName: "default-aurora-mysql5.7",
					},
					Status: v1alpha1.DBClusterStatus{
						Phase:                   v1alpha1.Available,
						SecretsManagerVersionID: "",
					},
				}),
				Log:              logf.Log,
				Scheme:           testScheme,
				CloudDBInterface: &factory.MockCloudDB{},
			},
		},
		{
			name:    "Test-AWS DBluster - when master user password secret does not exist",
			want:    controllerruntime.Result{},
			wantErr: true,
			args: args{
				ctx: context.Background(),
				req: controllerruntime.Request{NamespacedName: types.NamespacedName{
					Namespace: "default",
					Name:      "aws-db-cluster",
				}},
			},
			fields: fields{
				Client: fake.NewFakeClientWithScheme(testScheme, &v1alpha1.DBCluster{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "aws-db-cluster",
						Namespace: "default",
					},
					Spec: v1alpha1.DBClusterSpec{
						Provider: v1alpha1.Provider{
							Type: "aws",
							SecretRef: v1.SecretReference{
								Name:      "aws-provider-secret",
								Namespace: "default",
							},
						},
						Region:            "us-east-1",
						AvailabilityZones: []string{"us-east-1a", "us-east-1b"},
						DatabaseName:      "test",
						Engine:            "aurora-mysql",
						EngineMode:        "provisioned",
						EngineVersion:     "5.7.12",
						MasterUsername:    "test",
						MasterUserPasswordSecretRef: v1alpha1.MasterUserPasswordSecretRef{
							PasswordKey: "password",
							SecretRef: v1.SecretReference{
								Name:      "dbcluster-password",
								Namespace: "default",
							},
						},
						DBClusterParameterGroupName: "default-aurora-mysql5.7",
					},
					Status: v1alpha1.DBClusterStatus{
						Phase:                   v1alpha1.Available,
						SecretsManagerVersionID: "",
					},
				}, &v1.Secret{
					ObjectMeta: metav1.ObjectMeta{
						Namespace: "default",
						Name:      "aws-provider-secret",
					},
					Data: map[string][]byte{
						"AWS_ACCESS_KEY_ID":     []byte("fake-id"),
						"AWS_SECRET_ACCESS_KEY": []byte("fake-access-key"),
					},
					Type: v1.SecretTypeOpaque,
				}),
				Log:              logf.Log,
				Scheme:           testScheme,
				CloudDBInterface: &factory.MockCloudDB{},
			},
		},
		{
			name:    "Test-AWS DBluster - when dbCluster CR has a deletion timestamp with deletionProtection disabled",
			want:    controllerruntime.Result{Requeue: false},
			wantErr: false,
			args: args{
				ctx: context.Background(),
				req: controllerruntime.Request{NamespacedName: types.NamespacedName{
					Namespace: "default",
					Name:      "aws-db-cluster",
				}},
			},
			fields: fields{
				Client: fake.NewFakeClientWithScheme(testScheme, &v1alpha1.DBCluster{
					ObjectMeta: metav1.ObjectMeta{
						Name:              "aws-db-cluster",
						Namespace:         "default",
						DeletionTimestamp: &metav1.Time{Time: timeNow},
					},
					Spec: v1alpha1.DBClusterSpec{
						Provider: v1alpha1.Provider{
							Type: "aws",
							SecretRef: v1.SecretReference{
								Name:      "aws-provider-secret",
								Namespace: "default",
							},
						},
						Region:             "us-east-1",
						AvailabilityZones:  []string{"us-east-1a", "us-east-1b"},
						DeletionProtection: false,
						DatabaseName:       "test",
						Engine:             "aurora-mysql",
						EngineMode:         "provisioned",
						EngineVersion:      "5.7.12",
						MasterUsername:     "test",
						MasterUserPasswordSecretRef: v1alpha1.MasterUserPasswordSecretRef{
							PasswordKey: "password",
							SecretRef: v1.SecretReference{
								Name:      "dbcluster-password",
								Namespace: "default",
							},
						},
						DBClusterParameterGroupName: "default-aurora-mysql5.7",
					},
				}, &v1.Secret{
					ObjectMeta: metav1.ObjectMeta{
						Namespace: "default",
						Name:      "aws-provider-secret",
					},
					Data: map[string][]byte{
						"AWS_ACCESS_KEY_ID":     []byte("fake-id"),
						"AWS_SECRET_ACCESS_KEY": []byte("fake-access-key"),
					},
					Type: v1.SecretTypeOpaque,
				}, &v1.Secret{
					ObjectMeta: metav1.ObjectMeta{
						Namespace: "default",
						Name:      "dbcluster-password",
					},
					Data: map[string][]byte{
						"password": []byte("test"),
					},
					Type: v1.SecretTypeOpaque,
				}),
				Log:    logf.Log,
				Scheme: testScheme,
				CloudDBInterface: &factory.MockCloudDB{
					DeleteDBClusterErr:      nil,
					IsDBClusterUpToDateResp: true,
					DBClusterExistsResp:     true,
					DBClusterExistsStatus:   string(v1alpha1.Available),
				},
			},
		},
		{
			name:    "Test-AWS DBluster - when dbCluster CR has a deletion timestamp with deletionProtection enabled",
			want:    controllerruntime.Result{},
			wantErr: true,
			args: args{
				ctx: context.Background(),
				req: controllerruntime.Request{NamespacedName: types.NamespacedName{
					Namespace: "default",
					Name:      "aws-db-cluster",
				}},
			},
			fields: fields{
				Client: fake.NewFakeClientWithScheme(testScheme, &v1alpha1.DBCluster{
					ObjectMeta: metav1.ObjectMeta{
						Name:              "aws-db-cluster",
						Namespace:         "default",
						DeletionTimestamp: &metav1.Time{Time: timeNow},
					},
					Spec: v1alpha1.DBClusterSpec{
						Provider: v1alpha1.Provider{
							Type: "aws",
							SecretRef: v1.SecretReference{
								Name:      "aws-provider-secret",
								Namespace: "default",
							},
						},
						Region:             "us-east-1",
						AvailabilityZones:  []string{"us-east-1a", "us-east-1b"},
						DeletionProtection: false,
						DatabaseName:       "test",
						Engine:             "aurora-mysql",
						EngineMode:         "provisioned",
						EngineVersion:      "5.7.12",
						MasterUsername:     "test",
						MasterUserPasswordSecretRef: v1alpha1.MasterUserPasswordSecretRef{
							PasswordKey: "password",
							SecretRef: v1.SecretReference{
								Name:      "dbcluster-password",
								Namespace: "default",
							},
						},
						DBClusterParameterGroupName: "default-aurora-mysql5.7",
					},
				}, &v1.Secret{
					ObjectMeta: metav1.ObjectMeta{
						Namespace: "default",
						Name:      "aws-provider-secret",
					},
					Data: map[string][]byte{
						"AWS_ACCESS_KEY_ID":     []byte("fake-id"),
						"AWS_SECRET_ACCESS_KEY": []byte("fake-access-key"),
					},
					Type: v1.SecretTypeOpaque,
				}, &v1.Secret{
					ObjectMeta: metav1.ObjectMeta{
						Namespace: "default",
						Name:      "dbcluster-password",
					},
					Data: map[string][]byte{
						"password": []byte("test"),
					},
					Type: v1.SecretTypeOpaque,
				}),
				Log:    logf.Log,
				Scheme: testScheme,
				CloudDBInterface: &factory.MockCloudDB{
					DeleteDBClusterErr:      aws2.ErrDBClusterDeletionProtectionEnabled{Message: "errDeletionProtectionIsEnabled"},
					IsDBClusterUpToDateResp: true,
					DBClusterExistsResp:     true,
					DBClusterExistsStatus:   string(v1alpha1.Available),
				},
			},
		},
		{
			name:    "Test-AWS DBluster - when cluster exists but is not yet available",
			want:    controllerruntime.Result{Requeue: true, RequeueAfter: 30 * time.Second},
			wantErr: false,
			args: args{
				ctx: context.Background(),
				req: controllerruntime.Request{NamespacedName: types.NamespacedName{
					Namespace: "default",
					Name:      "aws-db-cluster",
				}},
			},
			fields: fields{
				Client: fake.NewFakeClientWithScheme(testScheme, &v1alpha1.DBCluster{
					ObjectMeta: metav1.ObjectMeta{
						Name:              "aws-db-cluster",
						Namespace:         "default",
						DeletionTimestamp: &metav1.Time{Time: timeNow},
					},
					Spec: v1alpha1.DBClusterSpec{
						Provider: v1alpha1.Provider{
							Type: "aws",
							SecretRef: v1.SecretReference{
								Name:      "aws-provider-secret",
								Namespace: "default",
							},
						},
						Region:             "us-east-1",
						AvailabilityZones:  []string{"us-east-1a", "us-east-1b"},
						DeletionProtection: false,
						DatabaseName:       "test",
						Engine:             "aurora-mysql",
						EngineMode:         "provisioned",
						EngineVersion:      "5.7.12",
						MasterUsername:     "test",
						MasterUserPasswordSecretRef: v1alpha1.MasterUserPasswordSecretRef{
							PasswordKey: "password",
							SecretRef: v1.SecretReference{
								Name:      "dbcluster-password",
								Namespace: "default",
							},
						},
						DBClusterParameterGroupName: "default-aurora-mysql5.7",
					},
				}, &v1.Secret{
					ObjectMeta: metav1.ObjectMeta{
						Namespace: "default",
						Name:      "aws-provider-secret",
					},
					Data: map[string][]byte{
						"AWS_ACCESS_KEY_ID":     []byte("fake-id"),
						"AWS_SECRET_ACCESS_KEY": []byte("fake-access-key"),
					},
					Type: v1.SecretTypeOpaque,
				}, &v1.Secret{
					ObjectMeta: metav1.ObjectMeta{
						Namespace: "default",
						Name:      "dbcluster-password",
					},
					Data: map[string][]byte{
						"password": []byte("test"),
					},
					Type: v1.SecretTypeOpaque,
				}),
				Log:    logf.Log,
				Scheme: testScheme,
				CloudDBInterface: &factory.MockCloudDB{
					DBClusterExistsResp:   true,
					DBClusterExistsStatus: string(v1alpha1.Creating),
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := &DBClusterReconciler{
				Client:           tt.fields.Client,
				Log:              tt.fields.Log,
				Scheme:           tt.fields.Scheme,
				CloudDBInterface: tt.fields.CloudDBInterface,
			}
			got, err := r.Reconcile(tt.args.ctx, tt.args.req)
			if (err != nil) != tt.wantErr {
				t.Errorf("Reconcile() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Reconcile() got = %v, want %v", got, tt.want)
			}
		})
	}
}
