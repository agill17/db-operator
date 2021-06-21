/*
Copyright 2021 agill17.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package v1alpha1

import (
	"errors"
	"fmt"
	"k8s.io/apimachinery/pkg/runtime"
	"reflect"
	ctrl "sigs.k8s.io/controller-runtime"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/webhook"
	"strings"
)

const (
	optioanlEngineTagKey          = "applicable-for-engines"
	requiredFieldsPerEngineTagKey = "required-for-engines"
)

// log is for logging in this package.
var dbinstancelog = logf.Log.WithName("dbinstance-webhook")

func (r *DBInstance) SetupWebhookWithManager(mgr ctrl.Manager) error {
	return ctrl.NewWebhookManagedBy(mgr).
		For(r).
		Complete()
}

// +kubebuilder:webhook:path=/validate-agill-apps-db-operator-v1alpha1-dbinstance,mutating=false,failurePolicy=fail,sideEffects=None,groups=agill.apps.db-operator,resources=dbinstances,verbs=create;update,versions=v1alpha1,name=vdbinstance.kb.io,admissionReviewVersions={v1,v1beta1}
var _ webhook.Validator = &DBInstance{}

// ValidateCreate implements webhook.Validator so a webhook will be registered for the type
func (r *DBInstance) ValidateCreate() error {
	namespacedName := fmt.Sprintf("%s/%s", r.GetNamespace(), r.GetName())
	dbinstancelog.Info(fmt.Sprintf("%s - validating create", namespacedName))
	if err := r.validateRequiredFieldsPerEngine(); err != nil {
		return err
	}
	return nil
}

// ValidateUpdate implements webhook.Validator so a webhook will be registered for the type
func (r *DBInstance) ValidateUpdate(old runtime.Object) error {
	namespacedName := fmt.Sprintf("%s/%s", r.GetNamespace(), r.GetName())
	dbinstancelog.Info(fmt.Sprintf("%s - validating update", namespacedName))
	if err := r.validateRequiredFieldsPerEngine(); err != nil {
		return err
	}
	return nil
}

// ValidateDelete implements webhook.Validator so a webhook will be registered for the type
func (r *DBInstance) ValidateDelete() error {
	return nil
}

func (r *DBInstance) validateRequiredFieldsPerEngine() error {
	namespacedName := fmt.Sprintf("%s/%s", r.GetNamespace(), r.GetName())
	spec := r.Spec
	var errs []string
	requiredFieldsForEngine := getRequiredFieldsPerEngine(spec)
	requiredFields, ok := requiredFieldsForEngine[spec.Engine]
	if !ok {
		dbinstancelog.Info(fmt.Sprintf("%s - '%s' engine does not have any required fields", namespacedName, spec.Engine))
		return nil
	}
	for _, field := range requiredFields {
		requiredFieldValueIsZero := reflect.ValueOf(spec).FieldByName(field).IsZero()
		if requiredFieldValueIsZero {
			msg := fmt.Sprintf("%s - '%s' field is not defined and is required by '%s' engine", namespacedName, field, spec.Engine)
			errs = append(errs, msg)
		}
	}
	if errs != nil && len(errs) > 0 {
		return errors.New(strings.Join(errs, ","))
	}
	return nil
}

// { engineName1: [requiredField1, requiredField2], engineName2: [requiredField1, requiredField2] }
func getRequiredFieldsPerEngine(in DBInstanceSpec) map[string][]string {
	rt := reflect.TypeOf(in)
	if rt.Kind() != reflect.Struct {
		panic("error, expected struct but got something else")
	}

	result := map[string][]string{}
	for i := 0; i < rt.NumField(); i++ {
		field := rt.Field(i)                                                   // FieldVar dataType `tags`
		fieldName := field.Name                                                // FieldVar
		fieldRequiredByEngines := field.Tag.Get(requiredFieldsPerEngineTagKey) // values for TagKey called requiredFieldsPerEngineTagKey
		if fieldRequiredByEngines == "" {
			continue
		}
		engineSplit := strings.Split(fieldRequiredByEngines, ",")
		for j := 0; j < len(engineSplit); j++ {
			result[engineSplit[j]] = append(result[engineSplit[j]], fieldName)
		}
	}
	return result
}
