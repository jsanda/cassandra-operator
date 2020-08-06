package reconciliation

import (
	"context"
	api "github.com/jsanda/cassandra-operator/api/v1alpha1"
	"github.com/jsanda/cassandra-operator/pkg/result"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

func (r *requestHandler) CheckHeadlessServices(ctx context.Context) result.ReconcileResult {
	allPodsService := newAllPodsServiceForCassandraCluster(r.cluster)
	seedsService := newSeedsServiceForCassandraCluster(r.cluster)

	services := []*corev1.Service{seedsService, allPodsService}

	for idx := range services {
		desiredSvc := services[idx]

		err := controllerutil.SetControllerReference(r.cluster, desiredSvc, r.scheme)
		if err != nil {
			r.log.Error(err, "could not set controller reference for headless desiredSvc", "Service", desiredSvc.Name)
			return result.Error(err)
		}

		actualSvc := &corev1.Service{}
		err = r.Get(ctx, types.NamespacedName{Namespace: desiredSvc.Namespace, Name: desiredSvc.Name}, actualSvc)
		if err != nil && errors.IsNotFound(err) {
			if err = r.Create(context.TODO(), desiredSvc); err != nil {
				r.log.Error(err, "failed to create headless service", "Service", desiredSvc.Name)
				return result.Error(err)
			}
		} else if err != nil {
			r.log.Error(err, "could not get headless service", "Service", desiredSvc.Name)
			return result.Error(err)
		} else {
			// TODO Check to see if the service needs to be updated
			return result.Continue()
		}
	}

	return result.Continue()
}

func newAllPodsServiceForCassandraCluster(cluster *api.CassandraCluster) *corev1.Service {
	service := makeGenericHeadlessService(cluster)
	service.ObjectMeta.Name = cluster.GetAllPodsServiceName()
	service.Spec.PublishNotReadyAddresses = true

	addHashAnnotation(service.ObjectMeta)

	return service
}

func newSeedsServiceForCassandraCluster(cluster *api.CassandraCluster) *corev1.Service {
	service := makeGenericHeadlessService(cluster)
	service.ObjectMeta.Name = cluster.GetSeedsServiceName()

	labels := cluster.GetClusterLabels()
	service.ObjectMeta.Labels = labels

	// Commenting out the call to buildLabelSelectorForSeedService for now because we are not
	// currently applying additional labels to pods other than what is done in the
	// PodTemplateSpec.
	//
	//service.Spec.Selector = buildLabelSelectorForSeedService(cluster)
	service.Spec.PublishNotReadyAddresses = true

	addHashAnnotation(service.ObjectMeta)

	return service
}

// makeGenericHeadlessService returns a fresh k8s headless (aka ClusterIP equals "None") Service
// struct that has the same namespace as the CassandraDatacenter argument, and proper labels for the DC.
// The caller needs to fill in the ObjectMeta.Name value, at a minimum, before it can be created
// inside the k8s cluster.
//
// Source: http://github.com/datastax/cass-operator/blob/master/operator/pkg/reconciliation/constructor.go#L179-L179
func makeGenericHeadlessService(cluster *api.CassandraCluster) *corev1.Service {
	labels := cluster.GetClusterLabels()
	api.AddManagedByLabel(labels)

	var service corev1.Service
	service.ObjectMeta.Namespace = cluster.Namespace
	service.ObjectMeta.Labels = labels
	service.Spec.Selector = cluster.GetClusterLabels()
	service.Spec.Type = "ClusterIP"
	service.Spec.ClusterIP = "None"

	return &service
}

func buildLabelSelectorForSeedService(cluster *api.CassandraCluster) map[string]string {
	labels := cluster.GetClusterLabels()

	// narrow selection to just the seed nodes
	labels[api.SeedNodeLabel] = "true"

	return labels
}
