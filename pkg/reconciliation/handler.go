package reconciliation

import (
	"context"
	"github.com/go-logr/logr"
	api "github.com/jsanda/cassandra-operator/api/v1alpha1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"time"
)

type RequestHandler interface {
	HandleRequest(ctx context.Context) (reconcile.Result, error)
}

const (
	k8sRequestTimeout = time.Second * 5
)

type requestHandler struct {
	request *reconcile.Request
	client.Client
	scheme *runtime.Scheme
	log logr.Logger
	cluster *api.CassandraCluster
}

func NewRequestHandler(request *reconcile.Request, client client.Client, scheme *runtime.Scheme, log logr.Logger) RequestHandler {
	return &requestHandler{
		request: request,
		Client: client,
		scheme: scheme,
		log: log,
	}
}

func (r *requestHandler) Get(ctx context.Context, key types.NamespacedName, obj runtime.Object) error {
	requestCtx, cancel := context.WithTimeout(ctx, k8sRequestTimeout)
	defer cancel()
	return r.Client.Get(requestCtx, key, obj)
}

func (r *requestHandler) HandleRequest(ctx context.Context) (reconcile.Result, error) {
	cluster := &api.CassandraCluster{}
	err := r.Get(ctx, r.request.NamespacedName, cluster)
	if err != nil {
		if errors.IsNotFound(err) {
			return ctrl.Result{}, nil
		} else {
			return ctrl.Result{}, err
		}
	}
	r.cluster = cluster

	if result := r.CheckHeadlessServices(ctx); result.Completed() {
		return result.Output()
	}

	if result := r.CheckStatefulSet(ctx); result.Completed() {
		return result.Output()
	}

	return reconcile.Result{}, nil
}
