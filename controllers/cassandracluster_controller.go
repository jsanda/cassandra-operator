/*


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

package controllers

import (
	"context"
	api "github.com/jsanda/cassandra-operator/api/v1alpha1"
	"github.com/jsanda/cassandra-operator/pkg/result"
	"github.com/jsanda/cassandra-operator/pkg/reconciliation"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"

	"github.com/go-logr/logr"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	corev1 "k8s.io/api/core/v1"
)

// CassandraClusterReconciler reconciles a CassandraCluster object
type CassandraClusterReconciler struct {
	client.Client
	Log    logr.Logger
	Scheme *runtime.Scheme
}

func (r *CassandraClusterReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&api.CassandraCluster{}).
		Complete(r)
}

// +kubebuilder:rbac:groups=cassandra.apache.org,resources=cassandraclusters,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=cassandra.apache.org,resources=cassandraclusters/status,verbs=get;update;patch

func (r *CassandraClusterReconciler) Reconcile(req ctrl.Request) (ctrl.Result, error) {
	ctx := context.Background()
	_ = r.Log.WithValues("cassandracluster", req.NamespacedName)

	// your logic here
	cluster := &api.CassandraCluster{}
	err := r.Client.Get(ctx, req.NamespacedName, cluster)
	if err != nil {
		if errors.IsNotFound(err) {
			return ctrl.Result{}, nil
		} else {
			return ctrl.Result{}, err
		}
	}

	//status := cluster.Status.DeepCopy()

	r.CheckHeadlessServices(cluster)

	return ctrl.Result{}, nil
}

func (r *CassandraClusterReconciler) CheckHeadlessServices(cluster *api.CassandraCluster) result.ReconcileResult {
	allPodsService := newAllPodsServiceForCassandraDatacenter(cluster)

	err := controllerutil.SetControllerReference(cluster, allPodsService, r.Scheme)
	if err != nil {
		r.Log.Error(err, "could not set controller reference", "AllPodsService", allPodsService.Name)
		return result.Error(err)
	}

	svc := &corev1.Service{}
	err = r.Get(context.TODO(), types.NamespacedName{Namespace: allPodsService.Namespace, Name: allPodsService.Name}, svc)
	if err != nil && errors.IsNotFound(err) {
		return r.CreateHeadlessService(allPodsService)
	} else if err != nil {
		r.Log.Error(err, "failed to get headless service", "AllPodsService", allPodsService.Name)
		return result.Error(err)
	} else {
		// TODO Check to see if the service needs to be updated
		return result.Continue()
	}
}

func (r *CassandraClusterReconciler) CreateHeadlessService(svc *corev1.Service) result.ReconcileResult {
	r.Log.Info("creating headless service", "Service.Namespace", svc.Namespace, "Service.Name", svc.Name)

	if err := r.Create(context.TODO(), svc); err != nil {
		r.Log.Error(err, "could not create headless service")
		return result.Error(err)
	}

	return result.Continue()
}

func newAllPodsServiceForCassandraDatacenter(cluster *api.CassandraCluster) *corev1.Service {
	service := makeGenericHeadlessService(cluster)
	service.ObjectMeta.Name = cluster.GetAllPodsServiceName()
	service.Spec.PublishNotReadyAddresses = true

	reconciliation.AddHashAnnotation(service.ObjectMeta)

	return service
}

// makeGenericHeadlessService returns a fresh k8s headless (aka ClusterIP equals "None") Service
// struct that has the same namespace as the CassandraDatacenter argument, and proper labels for the DC.
// The caller needs to fill in the ObjectMeta.Name value, at a minimum, before it can be created
// inside the k8s cluster. This is copied from http://github.com/jsanda/cass-operator/blob/master/operator/pkg/reconciliation/constructor.go#L179-L179
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
