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

func TestDefaultNetworkPoliciesLocalDbEnabled(t *testing.T) {
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

	npObj := model.GetRuntimeObject(NetworkPolicyKey)
	assert.NotNil(t, npObj, "NetworkPolicies should be in the model")

	nps := npObj.(*NetworkPolicies)
	assert.Equal(t, 2, nps.PolicyCount(), "Should have 2 policies (backend + db) when local DB is enabled")

	// Verify backend policy
	backendNP := nps.BackendNetworkPolicy()
	assert.NotNil(t, backendNP, "Backend NetworkPolicy should exist")
	assert.Equal(t, BackstageNetworkPolicyName(bs.Name), backendNP.Name)
	assert.Equal(t, "backstage-test-np", backendNP.Spec.PodSelector.MatchLabels[BackstageAppLabel])
	assert.Contains(t, backendNP.Spec.PolicyTypes, networkingv1.PolicyTypeIngress)
	assert.Contains(t, backendNP.Spec.PolicyTypes, networkingv1.PolicyTypeEgress)
	assert.NotEmpty(t, backendNP.Spec.Ingress, "Backend should have ingress rules")
	assert.NotEmpty(t, backendNP.Spec.Egress, "Backend should have egress rules")

	// Backend should NOT have port 5432 egress when local DB is enabled
	assert.False(t, hasEgressPort(backendNP, 5432), "Backend should not have port 5432 egress when local DB is enabled")

	// Verify DB policy
	dbNP := nps.DbNetworkPolicy()
	assert.NotNil(t, dbNP, "DB NetworkPolicy should exist when local DB is enabled")
	assert.Equal(t, DbNetworkPolicyName(bs.Name), dbNP.Name)
	assert.Equal(t, "backstage-psql-test-np", dbNP.Spec.PodSelector.MatchLabels[BackstageAppLabel])
	assert.Contains(t, dbNP.Spec.PolicyTypes, networkingv1.PolicyTypeIngress)
	assert.Contains(t, dbNP.Spec.PolicyTypes, networkingv1.PolicyTypeEgress)
	assert.NotEmpty(t, dbNP.Spec.Ingress, "DB should have ingress rules")
	assert.NotEmpty(t, dbNP.Spec.Egress, "DB should have egress rules")
}

func TestNetworkPoliciesLocalDbDisabled(t *testing.T) {
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

	testObj := createBackstageTest(bs).withDefaultConfig(true)
	model, err := InitObjects(context.TODO(), bs, testObj.externalConfig, platform.Default, testObj.scheme)
	assert.NoError(t, err)

	npObj := model.GetRuntimeObject(NetworkPolicyKey)
	assert.NotNil(t, npObj, "NetworkPolicies should still be in the model (backend only)")

	nps := npObj.(*NetworkPolicies)
	assert.Equal(t, 1, nps.PolicyCount(), "Should have 1 policy (backend only) when local DB is disabled")

	// Verify DB policy is filtered out
	dbNP := nps.DbNetworkPolicy()
	assert.Nil(t, dbNP, "DB NetworkPolicy should not exist when local DB is disabled")

	// Verify backend policy has port 5432 egress for external DB
	backendNP := nps.BackendNetworkPolicy()
	assert.NotNil(t, backendNP, "Backend NetworkPolicy should exist")
	assert.True(t, hasEgressPort(backendNP, 5432), "Backend should have port 5432 egress for external DB when local DB is disabled")
}

func TestNetworkPoliciesNotConfigured(t *testing.T) {
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

	npObj := model.GetRuntimeObject(NetworkPolicyKey)
	assert.Nil(t, npObj, "NetworkPolicies should not be returned when not configured")
}

func TestNetworkPoliciesNamespace(t *testing.T) {
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

	npObj := model.GetRuntimeObject(NetworkPolicyKey)
	assert.NotNil(t, npObj)

	nps := npObj.(*NetworkPolicies)
	backendNP := nps.BackendNetworkPolicy()
	assert.NotNil(t, backendNP)
	assert.Equal(t, "my-namespace", backendNP.Namespace)

	dbNP := nps.DbNetworkPolicy()
	assert.NotNil(t, dbNP)
	assert.Equal(t, "my-namespace", dbNP.Namespace)
}

func TestNetworkPoliciesMultipleCRs(t *testing.T) {
	// Verify that different CR names produce independent policy names
	bs1 := api.Backstage{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "app-one",
			Namespace: "ns123",
		},
		Spec: api.BackstageSpec{
			Application: &api.Application{},
		},
	}

	bs2 := api.Backstage{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "app-two",
			Namespace: "ns123",
		},
		Spec: api.BackstageSpec{
			Application: &api.Application{},
		},
	}

	testObj1 := createBackstageTest(bs1).withDefaultConfig(true)
	model1, err := InitObjects(context.TODO(), bs1, testObj1.externalConfig, platform.Default, testObj1.scheme)
	assert.NoError(t, err)

	testObj2 := createBackstageTest(bs2).withDefaultConfig(true)
	model2, err := InitObjects(context.TODO(), bs2, testObj2.externalConfig, platform.Default, testObj2.scheme)
	assert.NoError(t, err)

	nps1 := model1.GetRuntimeObject(NetworkPolicyKey).(*NetworkPolicies)
	nps2 := model2.GetRuntimeObject(NetworkPolicyKey).(*NetworkPolicies)

	// Backend policies should have different names
	assert.NotEqual(t, nps1.BackendNetworkPolicy().Name, nps2.BackendNetworkPolicy().Name)
	assert.Equal(t, "backstage-app-one", nps1.BackendNetworkPolicy().Name)
	assert.Equal(t, "backstage-app-two", nps2.BackendNetworkPolicy().Name)

	// DB policies should have different names
	assert.NotEqual(t, nps1.DbNetworkPolicy().Name, nps2.DbNetworkPolicy().Name)
	assert.Equal(t, "backstage-psql-app-one", nps1.DbNetworkPolicy().Name)
	assert.Equal(t, "backstage-psql-app-two", nps2.DbNetworkPolicy().Name)
}
