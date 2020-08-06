package reconciliation

import (
	"crypto/sha256"
	"encoding/base64"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/kubernetes/pkg/util/hash"
)

const resourceHashAnnotationKey = "cassandra.datastax.com/resource-hash"

// Source: http://github.com/datastax/cass-operator/blob/master/operator/pkg/reconciliation/hash_annotation.go#L20-L20
func resourcesHaveSameHash(r1, r2 metav1.ObjectMeta) bool {
	a1 := r1.GetAnnotations()
	a2 := r2.GetAnnotations()
	if a1 == nil || a2 == nil {
		return false
	}
	return a1[resourceHashAnnotationKey] == a2[resourceHashAnnotationKey]
}

// Source: http://github.com/datastax/cass-operator/blob/master/operator/pkg/reconciliation/hash_annotation.go#L29-L29
func addHashAnnotation(r metav1.ObjectMeta) {
	hash := deepHashString(r)
	m := r.GetAnnotations()
	if m == nil {
		m = map[string]string{}
	}
	m[resourceHashAnnotationKey] = hash
	r.SetAnnotations(m)
}

// Source: http://github.com/datastax/cass-operator/blob/master/operator/pkg/reconciliation/hash_annotation.go#L39-L39
func deepHashString(obj interface{}) string {
	hasher := sha256.New()
	hash.DeepHashObject(hasher, obj)
	hashBytes := hasher.Sum([]byte{})
	b64Hash := base64.StdEncoding.EncodeToString(hashBytes)
	return b64Hash
}
