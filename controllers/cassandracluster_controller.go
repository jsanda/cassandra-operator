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
	"github.com/go-logr/logr"
	api "github.com/jsanda/cassandra-operator/api/v1alpha1"
	"github.com/jsanda/cassandra-operator/pkg/reconciliation"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
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

// +kubebuilder:rbac:groups=cassandra.apache.org,namespace="cassandra-operator",resources=cassandraclusters,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=cassandra.apache.org,namespace="cassandra-operator",resources=cassandraclusters/status,verbs=get;update;patch
// +kubebuilder:rbac:groups="",namespace="cassandra-operator",resources=services,verbs=get;list;watch;create
// +kubebuilder:rbac:groups="apps",namespace="cassandra-operator",resources=statefulsets,verbs=get;list;watch;create

func (r *CassandraClusterReconciler) Reconcile(req ctrl.Request) (ctrl.Result, error) {
	ctx := context.Background()
	logger := r.Log.WithValues("cassandracluster", req.NamespacedName)

	handler := reconciliation.NewRequestHandler(&req, r.Client, r.Scheme, logger)

	return handler.HandleRequest(ctx)

	//status := cluster.Status.DeepCopy()

	//r.CheckHeadlessServices(cluster)

	return ctrl.Result{}, nil
}
