package utils

import (
	"context"
	"k8s.io/apimachinery/pkg/api/meta"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func ListContainsString(list []string, lookupStr string) (bool, int) {
	for idx, e := range list {
		if e == lookupStr {
			return true, idx
		}
	}
	return false, -1
}

func AddFinalizer(finalizer string, client client.Client, object client.Object) error {
	metaObj, err := meta.Accessor(object)
	if err != nil {
		return err
	}
	// if obj is marked as getting deleted, do not try to add the finalizer
	if metaObj.GetDeletionTimestamp() != nil {
		return nil
	}

	currentFinalizers := metaObj.GetFinalizers()
	if ok, _ := ListContainsString(currentFinalizers, finalizer); !ok {
		currentFinalizers = append(currentFinalizers, finalizer)
		metaObj.SetFinalizers(currentFinalizers)
		return client.Update(context.TODO(), object)
	}
	return nil
}

func RemoveFinalizer(finalizer string, client client.Client, object client.Object) error {
	metaObj, err := meta.Accessor(object)
	if err != nil {
		return err
	}
	currentFinalizers := metaObj.GetFinalizers()
	if ok, idx := ListContainsString(currentFinalizers, finalizer); ok {
		currentFinalizers = append(currentFinalizers[:idx], currentFinalizers[idx+1:]...)
		metaObj.SetFinalizers(currentFinalizers)
		return client.Update(context.TODO(), object)
	}
	return nil
}
