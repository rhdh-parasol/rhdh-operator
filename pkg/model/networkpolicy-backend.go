package model

import (
	"k8s.io/apimachinery/pkg/runtime"

	"github.com/redhat-developer/rhdh-operator/api"
	"github.com/redhat-developer/rhdh-operator/pkg/utils"

	networkingv1 "k8s.io/api/networking/v1"
)

type BackstageNetworkPolicyFactory struct{}

func (f BackstageNetworkPolicyFactory) newBackstageObject() RuntimeObject {
	return &BackstageNetworkPolicy{}
}

type BackstageNetworkPolicy struct {
	networkPolicy *networkingv1.NetworkPolicy
	model         *BackstageModel
}

func init() {
	registerConfig(NetworkPolicyBackendKey, BackstageNetworkPolicyFactory{}, false, nil)
}

func BackstageNetworkPolicyName(backstageName string) string {
	return utils.GenerateRuntimeObjectName(backstageName, "backstage")
}

func (b *BackstageNetworkPolicy) Object() runtime.Object {
	if b.networkPolicy == nil {
		return nil
	}
	return b.networkPolicy
}

// implementation of RuntimeObject interface
func (b *BackstageNetworkPolicy) GetKey() string {
	return NetworkPolicyBackendKey
}

func (b *BackstageNetworkPolicy) addToModel(model *BackstageModel, backstage api.Backstage, config runtime.Object, scheme *runtime.Scheme) error {
	b.model = model

	if config != nil {
		b.networkPolicy = config.(*networkingv1.NetworkPolicy)
	}

	// Always add wrapper to model (unconditional)
	model.setRuntimeObject(b)

	// Only set metadata if underlying object exists
	if b.networkPolicy != nil {
		b.setMetaInfo(backstage, scheme)
	}

	return nil
}

func (b *BackstageNetworkPolicy) updateAndValidate(_ api.Backstage, _ *runtime.Scheme) error {
	return nil
}

func (b *BackstageNetworkPolicy) setMetaInfo(backstage api.Backstage, scheme *runtime.Scheme) {
	b.networkPolicy.SetName(BackstageNetworkPolicyName(backstage.Name))
	utils.GenerateLabel(&b.networkPolicy.Spec.PodSelector.MatchLabels, BackstageAppLabel, utils.BackstageAppLabelValue(backstage.Name))
	setMetaInfo(b.networkPolicy, backstage, scheme)
}
