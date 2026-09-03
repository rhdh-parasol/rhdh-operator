package model

import (
	"context"
	"testing"

	"github.com/redhat-developer/rhdh-operator/api"
	"github.com/redhat-developer/rhdh-operator/pkg/platform"

	networkingv1 "k8s.io/api/networking/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/utils/ptr"

	"github.com/stretchr/testify/assert"
)

func TestDefaultBackstageNetworkPolicy(t *testing.T) {
	bs := api.Backstage{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-np",
			Namespace: "ns123",
		},
		Spec: api.BackstageSpec{
			Application: &api.Application{},
		},
	}

	testObj := createBackstageTest(bs).withDefaultConfig(true)
	model, err := InitObjects(context.TODO(), bs, testObj.externalConfig, platform.Default, testObj.scheme)
	assert.NoError(t, err)

	npObj := model.GetRuntimeObject(NetworkPolicyBackendKey)
	assert.NotNil(t, npObj, "Backend NetworkPolicy should be in the model")

	np := npObj.(*BackstageNetworkPolicy).networkPolicy
	assert.Equal(t, BackstageNetworkPolicyName(bs.Name), np.Name)

	// Verify pod selector targets backstage app pods
	assert.Equal(t, "backstage-test-np", np.Spec.PodSelector.MatchLabels[BackstageAppLabel])

	// Verify policy types
	assert.Contains(t, np.Spec.PolicyTypes, networkingv1.PolicyTypeIngress)
	assert.Contains(t, np.Spec.PolicyTypes, networkingv1.PolicyTypeEgress)

	// Verify ingress rules exist
	assert.NotEmpty(t, np.Spec.Ingress, "Backend NetworkPolicy should have ingress rules")

	// Verify egress rules exist
	assert.NotEmpty(t, np.Spec.Egress, "Backend NetworkPolicy should have egress rules")
}

func TestBackstageNetworkPolicyNotConfigured(t *testing.T) {
	bs := api.Backstage{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-np",
			Namespace: "ns123",
		},
		Spec: api.BackstageSpec{
			Database:    &api.Database{EnableLocalDb: ptr.To(false)},
			Application: &api.Application{},
		},
	}

	// Use a config path that has required configs but not the network policy config
	testObj := createBackstageTest(bs).withConfigPath("testdata/testflavours")
	model, err := InitObjects(context.TODO(), bs, testObj.externalConfig, platform.Default, testObj.scheme)
	assert.NoError(t, err)

	npObj := model.GetRuntimeObject(NetworkPolicyBackendKey)
	assert.Nil(t, npObj, "Backend NetworkPolicy should not be returned when not configured")
}

func TestBackstageNetworkPolicyNamespace(t *testing.T) {
	bs := api.Backstage{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-np",
			Namespace: "my-namespace",
		},
		Spec: api.BackstageSpec{
			Application: &api.Application{},
		},
	}

	testObj := createBackstageTest(bs).withDefaultConfig(true)
	model, err := InitObjects(context.TODO(), bs, testObj.externalConfig, platform.Default, testObj.scheme)
	assert.NoError(t, err)

	npObj := model.GetRuntimeObject(NetworkPolicyBackendKey)
	assert.NotNil(t, npObj)

	np := npObj.(*BackstageNetworkPolicy).networkPolicy
	assert.Equal(t, "my-namespace", np.Namespace)
}
