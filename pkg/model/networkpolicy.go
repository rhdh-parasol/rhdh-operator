package model

import (
	"strings"

	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/intstr"

	"github.com/redhat-developer/rhdh-operator/api"
	"github.com/redhat-developer/rhdh-operator/pkg/model/multiobject"
	"github.com/redhat-developer/rhdh-operator/pkg/utils"

	networkingv1 "k8s.io/api/networking/v1"
	corev1 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type NetworkPolicyFactory struct{}

func (f NetworkPolicyFactory) newBackstageObject() RuntimeObject {
	return &NetworkPolicies{}
}

type NetworkPolicies struct {
	policies *multiobject.MultiObject
	model    *BackstageModel
}

func init() {
	registerConfig(NetworkPolicyKey, NetworkPolicyFactory{}, true, nil)
}

func BackstageNetworkPolicyName(backstageName string) string {
	return utils.GenerateRuntimeObjectName(backstageName, "backstage")
}

func DbNetworkPolicyName(backstageName string) string {
	return utils.GenerateRuntimeObjectName(backstageName, "backstage-psql")
}

func (b *NetworkPolicies) Object() runtime.Object {
	if b.policies != nil && len(b.policies.Items) > 0 {
		return b.policies
	}
	return nil
}

// implementation of RuntimeObject interface
func (b *NetworkPolicies) GetKey() string {
	return NetworkPolicyKey
}

func (b *NetworkPolicies) addToModel(model *BackstageModel, backstage api.Backstage, config runtime.Object, scheme *runtime.Scheme) error {
	b.model = model

	if config != nil {
		b.policies = config.(*multiobject.MultiObject)
		b.filterAndAdjust(model.localDbEnabled)
	}

	// Always add wrapper to model (unconditional)
	model.setRuntimeObject(b)

	// Only set metadata if underlying object exists
	if b.policies != nil && len(b.policies.Items) > 0 {
		b.setMetaInfo(backstage, scheme)
	}

	return nil
}

// filterAndAdjust removes DB-specific policies when local DB is disabled,
// and adds port 5432 egress to the backend policy to allow connections
// to an external database.
func (b *NetworkPolicies) filterAndAdjust(localDbEnabled bool) {
	if b.policies == nil {
		return
	}

	if localDbEnabled {
		// All policies apply as-is
		return
	}

	// Filter out DB policies and adjust backend egress
	var filtered []client.Object
	for _, item := range b.policies.Items {
		np, ok := item.(*networkingv1.NetworkPolicy)
		if !ok {
			filtered = append(filtered, item)
			continue
		}

		if isDbNetworkPolicy(np) {
			// Skip DB policies when local DB is disabled
			continue
		}

		// Add external DB egress (port 5432) to backend policy
		addExternalDbEgress(np)
		filtered = append(filtered, item)
	}

	b.policies.Items = filtered
}

// isDbNetworkPolicy returns true if the NetworkPolicy is for the PostgreSQL database.
// It checks both the metadata name and the podSelector labels for "psql" indicators.
func isDbNetworkPolicy(np *networkingv1.NetworkPolicy) bool {
	if strings.Contains(np.Name, "psql") {
		return true
	}
	for _, v := range np.Spec.PodSelector.MatchLabels {
		if strings.Contains(v, "psql") {
			return true
		}
	}
	return false
}

// addExternalDbEgress appends a port 5432 TCP egress rule to the policy,
// allowing the backend to reach external PostgreSQL databases.
func addExternalDbEgress(np *networkingv1.NetworkPolicy) {
	np.Spec.Egress = append(np.Spec.Egress, networkingv1.NetworkPolicyEgressRule{
		Ports: []networkingv1.NetworkPolicyPort{
			{
				Port:     &intstr.IntOrString{Type: intstr.Int, IntVal: 5432},
				Protocol: protocolPtr(corev1.ProtocolTCP),
			},
		},
	})
}

func protocolPtr(p corev1.Protocol) *corev1.Protocol {
	return &p
}

func (b *NetworkPolicies) updateAndValidate(_ api.Backstage, _ *runtime.Scheme) error {
	return nil
}

func (b *NetworkPolicies) setMetaInfo(backstage api.Backstage, scheme *runtime.Scheme) {
	for _, item := range b.policies.Items {
		np, ok := item.(*networkingv1.NetworkPolicy)
		if !ok {
			continue
		}

		if isDbNetworkPolicy(np) {
			np.SetName(DbNetworkPolicyName(backstage.Name))
			utils.GenerateLabel(&np.Spec.PodSelector.MatchLabels, BackstageAppLabel, utils.BackstageDbAppLabelValue(backstage.Name))
		} else {
			np.SetName(BackstageNetworkPolicyName(backstage.Name))
			utils.GenerateLabel(&np.Spec.PodSelector.MatchLabels, BackstageAppLabel, utils.BackstageAppLabelValue(backstage.Name))
		}

		setMetaInfo(np, backstage, scheme)
	}
}

// BackendNetworkPolicy returns the backend network policy if present, or nil.
func (b *NetworkPolicies) BackendNetworkPolicy() *networkingv1.NetworkPolicy {
	if b.policies == nil {
		return nil
	}
	for _, item := range b.policies.Items {
		np, ok := item.(*networkingv1.NetworkPolicy)
		if !ok {
			continue
		}
		if !isDbNetworkPolicy(np) {
			return np
		}
	}
	return nil
}

// DbNetworkPolicy returns the DB network policy if present, or nil.
func (b *NetworkPolicies) DbNetworkPolicy() *networkingv1.NetworkPolicy {
	if b.policies == nil {
		return nil
	}
	for _, item := range b.policies.Items {
		np, ok := item.(*networkingv1.NetworkPolicy)
		if !ok {
			continue
		}
		if isDbNetworkPolicy(np) {
			return np
		}
	}
	return nil
}

// PolicyCount returns the number of network policies in the multi-object.
func (b *NetworkPolicies) PolicyCount() int {
	if b.policies == nil {
		return 0
	}
	return len(b.policies.Items)
}

// hasEgressPort checks if any egress rule in the policy contains the specified port.
func hasEgressPort(np *networkingv1.NetworkPolicy, port int32) bool {
	for _, rule := range np.Spec.Egress {
		for _, p := range rule.Ports {
			if p.Port != nil && p.Port.IntVal == port {
				return true
			}
		}
	}
	return false
}
