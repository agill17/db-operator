package v1alpha1

import "fmt"

func (in *DBInstance) GetDBInstanceID() string {
	out := fmt.Sprintf("%s-%s", in.GetNamespace(), in.GetName())
	if in.Spec.DBInstanceIdentifierOverride != "" {
		out = in.Spec.DBInstanceIdentifierOverride
	}
	return out
}
