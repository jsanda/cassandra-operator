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

package v1alpha1

import (
	"encoding/json"
	"github.com/Jeffail/gabs"
	"github.com/pkg/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	serverconfig "github.com/datastax/cass-operator/operator/pkg/serverconfig"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

const (
	// ClusterLabel is the operator's label for the cluster name
	ClusterLabel = "cassandra.apache.org/cluster"

	ManagedByLabel = "app.kubernetes.io/managed-by"

	ManagedByLabelValue = "cassandra-operator"

	// SeedNodeLabel is the operator's label for the seed node state
	SeedNodeLabel = "cassandra.apache.org/seed-node"

	defaultConfigBuilderImage = "datastax/cass-config-builder:1.0.1"
)

type Rack struct {
	Name string `json:"name,omitempty"`
}

type Datacenter struct {
	Name string `json:"name,omitempty"`

	NodesPerRack int32 `json:"nodesPerRack,omitempty"`

	Racks []Rack `json:"racks,omitempty"`
}

// CassandraClusterSpec defines the desired state of CassandraCluster
type CassandraClusterSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	Name string `json:"name"`

	Datacenters []Datacenter `json:"datacenters,omitempty"`

	// +kubebuilder:validation:PreserveUnknownFields=true
	Config json.RawMessage `json:"config,omitempty"`
}

// CassandraClusterStatus defines the observed state of CassandraCluster
type CassandraClusterStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status

// CassandraCluster is the Schema for the cassandraclusters API
type CassandraCluster struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   CassandraClusterSpec   `json:"spec,omitempty"`
	Status CassandraClusterStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// CassandraClusterList contains a list of CassandraCluster
type CassandraClusterList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []CassandraCluster `json:"items"`
}

func (c *CassandraCluster) GetClusterLabels() map[string]string {
	return map[string]string{
		ClusterLabel: c.Spec.Name,
	}
}

func (c *CassandraCluster) GetAllPodsServiceName() string {
	return c.Spec.Name + "-all-pods-service"
}

func (c *CassandraCluster) GetSeedsServiceName() string {
	return c.Spec.Name + "-seed-service"
}

func AddManagedByLabel(m map[string]string) {
	m[ManagedByLabel] = ManagedByLabelValue
}

func HasManagedByCassandraOperatorLabel(m map[string]string) bool {
	v, ok := m[ManagedByLabel]
	return ok && v == ManagedByLabelValue
}

func (c *CassandraCluster) GetConfigBuilderImage() string {
	return defaultConfigBuilderImage
}

// GetConfigAsJSON gets a JSON-encoded string suitable for passing to configBuilder
func (c *CassandraCluster) GetConfigAsJSON() (string, error) {
	// We use the cluster seed-service name here for the seed list as it will
	// resolve to the seed nodes. This obviates the need to update the
	// cassandra.yaml whenever the seed nodes change.
	seeds := []string{c.GetSeedsServiceName()}

	//cql := 0
	//cqlSSL := 0
	//broadcast := 0
	//broadcastSSL := 0

	modelValues := serverconfig.GetModelValues(seeds, c.Spec.Name, c.Spec.Name,)

	var modelBytes []byte

	modelBytes, err := json.Marshal(modelValues)
	if err != nil {
		return "", err
	}

	// Combine the model values with the user-specified values

	modelParsed, err := gabs.ParseJSON([]byte(modelBytes))
	if err != nil {
		return "", errors.Wrap(err, "Model information for CassandraCluster resource was not properly configured")
	}

	if c.Spec.Config != nil {
		configParsed, err := gabs.ParseJSON([]byte(c.Spec.Config))
		if err != nil {
			return "", errors.Wrap(err, "Error parsing Spec.Config for CassandraCluster resource")
		}

		if err := modelParsed.Merge(configParsed); err != nil {
			return "", errors.Wrap(err, "Error merging Spec.Config for CassandraDatacenter resource")
		}
	}

	return modelParsed.String(), nil
}

func init() {
	SchemeBuilder.Register(&CassandraCluster{}, &CassandraClusterList{})
}
