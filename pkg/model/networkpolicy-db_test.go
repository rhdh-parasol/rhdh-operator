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

func TestDefaultDbNetworkPolicy(t *testing.T) {
	bs := api.Backstage{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-db-np",
			Namespace: "ns123",
		},
		Spec: api.BackstageSpec{
			Database:    &api.Database{},
			Application: &api.Application{},
		},
	}

	testObj := createBackstageTest(bs).withDefaultConfig(true)
	model, err := InitObjects(context.TODO(), bs, testObj.externalConfig, platform.Default, testObj.scheme)
	assert.NoError(t, err)

	npObj := model.GetRuntimeObject(NetworkPolicyDbKey)
	assert.NotNil(t, npObj, "DB NetworkPolicy should be in the model when local DB is enabled")

	np := npObj.(*DbNetworkPolicy).networkPolicy
	assert.Equal(t, DbNetworkPolicyName(bs.Name), np.Name)

	// Verify pod selector targets DB pods
	assert.Equal(t, "backstage-psql-test-db-np", np.Spec.PodSelector.MatchLabels[BackstageAppLabel])

	// Verify policy types
	assert.Contains(t, np.Spec.PolicyTypes, networkingv1.PolicyTypeIngress)
	assert.Contains(t, np.Spec.PolicyTypes, networkingv1.PolicyTypeEgress)

	// Verify ingress rules exist (PostgreSQL port)
	assert.NotEmpty(t, np.Spec.Ingress, "DB NetworkPolicy should have ingress rules")

	// Verify egress rules exist (DNS only)
	assert.NotEmpty(t, np.Spec.Egress, "DB NetworkPolicy should have egress rules")
}

func TestDbNetworkPolicyLocalDbDisabled(t *testing.T) {
	bs := api.Backstage{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-db-np",
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

	npObj := model.GetRuntimeObject(NetworkPolicyDbKey)
	assert.Nil(t, npObj, "DB NetworkPolicy should not be returned when local DB is disabled")
}

func TestDbNetworkPolicyNamespace(t *testing.T) {
	bs := api.Backstage{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-db-np",
			Namespace: "my-namespace",
		},
		Spec: api.BackstageSpec{
			Database:    &api.Database{},
			Application: &api.Application{},
		},
	}

	testObj := createBackstageTest(bs).withDefaultConfig(true)
	model, err := InitObjects(context.TODO(), bs, testObj.externalConfig, platform.Default, testObj.scheme)
	assert.NoError(t, err)

	npObj := model.GetRuntimeObject(NetworkPolicyDbKey)
	assert.NotNil(t, npObj)

	np := npObj.(*DbNetworkPolicy).networkPolicy
	assert.Equal(t, "my-namespace", np.Namespace)
}
