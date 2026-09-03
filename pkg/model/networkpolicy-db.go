package model

import (
	"fmt"

	"k8s.io/apimachinery/pkg/runtime"

	"github.com/redhat-developer/rhdh-operator/api"
	"github.com/redhat-developer/rhdh-operator/pkg/utils"

	networkingv1 "k8s.io/api/networking/v1"
)

type DbNetworkPolicyFactory struct{}

func (f DbNetworkPolicyFactory) newBackstageObject() RuntimeObject {
	return &DbNetworkPolicy{}
}

type DbNetworkPolicy struct {
	networkPolicy *networkingv1.NetworkPolicy
	model         *BackstageModel
}

func init() {
	registerConfig(NetworkPolicyDbKey, DbNetworkPolicyFactory{}, false, nil)
}

func DbNetworkPolicyName(backstageName string) string {
	return utils.GenerateRuntimeObjectName(backstageName, "backstage-psql")
}

func (b *DbNetworkPolicy) Object() runtime.Object {
	if b.networkPolicy == nil {
		return nil
	}
	return b.networkPolicy
}

// implementation of RuntimeObject interface
func (b *DbNetworkPolicy) GetKey() string {
	return NetworkPolicyDbKey
}

func (b *DbNetworkPolicy) addToModel(model *BackstageModel, backstage api.Backstage, config runtime.Object, scheme *runtime.Scheme) error {
	b.model = model

	// Only set networkPolicy if localDb is enabled
	if model.localDbEnabled {
		if config == nil {
			return fmt.Errorf("local database is enabled but networkpolicy-db.yaml config is missing")
		}
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

func (b *DbNetworkPolicy) updateAndValidate(_ api.Backstage, _ *runtime.Scheme) error {
	return nil
}

func (b *DbNetworkPolicy) setMetaInfo(backstage api.Backstage, scheme *runtime.Scheme) {
	b.networkPolicy.SetName(DbNetworkPolicyName(backstage.Name))
	utils.GenerateLabel(&b.networkPolicy.Spec.PodSelector.MatchLabels, BackstageAppLabel, utils.BackstageDbAppLabelValue(backstage.Name))
	setMetaInfo(b.networkPolicy, backstage, scheme)
}
