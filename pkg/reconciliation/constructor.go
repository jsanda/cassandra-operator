package reconciliation

import (
	api "github.com/jsanda/cassandra-operator/api/v1alpha1"
	corev1 "k8s.io/api/core/v1"
)

func buildServerConfigInitContainer(cluster *api.CassandraCluster) (*corev1.Container, error) {
	serverCfg := corev1.Container{}
	serverCfg.Name = "server-config-init"
	serverCfg.Image = cluster.GetConfigBuilderImage()
	serverCfgMount := corev1.VolumeMount{
		Name:      "server-config",
		MountPath: "/config",
	}
	serverCfg.VolumeMounts = []corev1.VolumeMount{serverCfgMount}

	useHostIpForBroadcast := "false"

	rackName := "rack1"
	serverVersion := "3.11.6"
	serverType := "cassandra"

	configData, err := cluster.GetConfigAsJSON()
	if err != nil {
		return nil, err
	}
	serverCfg.Env = []corev1.EnvVar{
		{Name: "CONFIG_FILE_DATA", Value: configData},
		{Name: "POD_IP", ValueFrom: selectorFromFieldPath("status.podIP")},
		{Name: "HOST_IP", ValueFrom: selectorFromFieldPath("status.hostIP")},
		{Name: "USE_HOST_IP_FOR_BROADCAST", Value: useHostIpForBroadcast},
		{Name: "RACK_NAME", Value: rackName},
		{Name: "PRODUCT_VERSION", Value: serverVersion},
		{Name: "PRODUCT_NAME", Value: serverType},
		// TODO remove this post 1.0
		{Name: "DSE_VERSION", Value: serverVersion},
	}

	return &serverCfg, nil
}

func selectorFromFieldPath(fieldPath string) *corev1.EnvVarSource {
	return &corev1.EnvVarSource{
		FieldRef: &corev1.ObjectFieldSelector{
			FieldPath: fieldPath,
		},
	}
}

func createVolumes() []corev1.Volume {
	serverConfig := corev1.Volume{}
	serverConfig.Name = "server-config"
	serverConfig.VolumeSource = corev1.VolumeSource{
		EmptyDir: &corev1.EmptyDirVolumeSource{},
	}

	serverLogs := corev1.Volume{}
	serverLogs.Name = "server-logs"
	serverLogs.VolumeSource = corev1.VolumeSource{
		EmptyDir: &corev1.EmptyDirVolumeSource{},
	}

	return []corev1.Volume{serverConfig, serverLogs}
}
