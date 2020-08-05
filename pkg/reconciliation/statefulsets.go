package reconciliation

import (
	"context"
	api "github.com/jsanda/cassandra-operator/api/v1alpha1"
	"github.com/jsanda/cassandra-operator/pkg/result"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
)

const (
pvcName = "server-data"
)

func (r *requestHandler) CheckStatefulSet(ctx context.Context) result.ReconcileResult {
	actualStatefulSet := &appsv1.StatefulSet{}
	dcName := "dc1"
	rackName := "rack-1"
	nsName := newNamespacedNameForStatefulSet(r.cluster, dcName, rackName)

	err := r.Get(ctx, nsName, actualStatefulSet)

	if err != nil && errors.IsNotFound(err) {
		// create the statefulset
		statefulSet, err := newStatefulSet(r.cluster, rackName)
		if err != nil {
			r.log.Error(err, "failed to create new statefulset", "StatefulSet", nsName.Name)
			return result.Error(err)
		}
		if err = r.Create(ctx, statefulSet); err != nil {
			r.log.Error(err, "failed to persist new statefulset", "StatefulSet", nsName.Name)
			return result.Error(err)
		}
		return result.Continue()
	} else if err != nil {
		r.log.Error(err, "failed to get statefulset", "StatefulSet", nsName.Name)
		return result.Error(err)
	} else {
		result.Continue()
	}

	return result.Done()
}

func newNamespacedNameForStatefulSet(cluster *api.CassandraCluster, dcName string, rackName string) types.NamespacedName {
	name := cluster.Spec.Name + "-" + dcName + "-" + rackName + "-sts"
	ns := cluster.Namespace

	return types.NamespacedName{Name: name, Namespace: ns}
}

func newStatefulSet(cluster *api.CassandraCluster, rackName string) (*appsv1.StatefulSet, error) {
	pvcLabels := cluster.GetClusterLabels()
	selectorLabels := cluster.GetClusterLabels()
	volumeClaimTemplates := []corev1.PersistentVolumeClaim{newDataVolumeClaimTemplate(pvcLabels)}
	nsName := newNamespacedNameForStatefulSet(cluster, "dc1", rackName)
	replicas := int32(3)

	podTemplateSpec, err := buildPodTemplateSpec(cluster, rackName)
	if err != nil {
		return nil, err
	}

	statefulSet := &appsv1.StatefulSet{
		ObjectMeta: metav1.ObjectMeta{
			Name: nsName.Name,
			Namespace: nsName.Namespace,
			Labels: cluster.GetClusterLabels(),
		},
		Spec: appsv1.StatefulSetSpec{
			Selector: &metav1.LabelSelector{
				MatchLabels: selectorLabels,
			},
			Replicas: &replicas,
			ServiceName: cluster.GetClusterName(),
			Template: *podTemplateSpec,
			VolumeClaimTemplates: volumeClaimTemplates,
		},
	}

	return statefulSet, nil
}

func buildPodTemplateSpec(cluster *api.CassandraCluster, rackName string) (*corev1.PodTemplateSpec, error) {
	template := &corev1.PodTemplateSpec{}

	podLabels := cluster.GetClusterLabels()
	api.AddManagedByLabel(podLabels)

	template.Labels = podLabels

	affinity := &corev1.Affinity{}
	affinity.PodAntiAffinity = calculatePodAntiAffinity()

	template.Spec.ServiceAccountName = "default"

	template.Spec.Volumes = createVolumes()

	serverConfigInitContainer, err := buildServerConfigInitContainer(cluster)
	if err != nil {
		return nil, err
	}

	template.Spec.InitContainers = []corev1.Container {*serverConfigInitContainer}

	var serverVolumeMounts []corev1.VolumeMount
	for _, c := range template.Spec.InitContainers {
		serverVolumeMounts = append(serverVolumeMounts, c.VolumeMounts...)
	}

	containers, err := buildContainers(cluster, serverVolumeMounts)
	if err != nil {
		return nil, err
	}
	template.Spec.Containers = containers

	return template, nil
}

func buildContainers(cluster *api.CassandraCluster, serverVolumeMounts []corev1.VolumeMount) ([]corev1.Container, error) {
	cassandraContainer := corev1.Container{}
	cassandraContainer.Name = "cassandra"
	cassandraContainer.Image = "cassandra:3.11.7"
	cassandraContainer.Resources = corev1.ResourceRequirements{
		Limits: corev1.ResourceList{
			"cpu": generateResourceQuantity("2"),
			"memory": generateResourceQuantity("2Gi"),
		},
		Requests: corev1.ResourceList{
			"cpu": generateResourceQuantity("2"),
			"memory": generateResourceQuantity("2Gi"),
		},
	}

	serverVolumeMounts = append(serverVolumeMounts, corev1.VolumeMount{
		Name:      pvcName,
		MountPath: "/var/lib/cassandra",
	})
	cassandraContainer.VolumeMounts = serverVolumeMounts

	return []corev1.Container{cassandraContainer}, nil
}

func newDataVolumeClaimTemplate(pvcLabels map[string]string) corev1.PersistentVolumeClaim {
	storageClassName := "server-storage"
	pvc := corev1.PersistentVolumeClaim{
		ObjectMeta: metav1.ObjectMeta{
			Name:   pvcName,
			Labels: pvcLabels,
		},
		Spec: corev1.PersistentVolumeClaimSpec{
			AccessModes: []corev1.PersistentVolumeAccessMode{
				corev1.ReadWriteOnce,
			},
			StorageClassName: &storageClassName,
			Resources: corev1.ResourceRequirements{
				Requests: corev1.ResourceList{
					"storage": generateResourceQuantity("5Gi"),
				},
			},
		},
	}
	return pvc
}

func generateResourceQuantity(qs string) resource.Quantity {
	q, _ := resource.ParseQuantity(qs)
	return q
}

func calculatePodAntiAffinity() *corev1.PodAntiAffinity {
	return &corev1.PodAntiAffinity{
		RequiredDuringSchedulingIgnoredDuringExecution: []corev1.PodAffinityTerm{
			{
				LabelSelector: &metav1.LabelSelector{
					MatchExpressions: []metav1.LabelSelectorRequirement{
						{
							Key:      api.ClusterLabel,
							Operator: metav1.LabelSelectorOpExists,
						},
					},
				},
				TopologyKey: "kubernetes.io/hostname",
			},
		},
	}
}
