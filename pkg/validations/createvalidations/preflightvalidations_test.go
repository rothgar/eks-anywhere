package createvalidations_test

import (
	"context"
	"testing"

	"github.com/golang/mock/gomock"
	. "github.com/onsi/gomega"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/aws/eks-anywhere/internal/test"
	"github.com/aws/eks-anywhere/pkg/api/v1alpha1"
	anywherev1 "github.com/aws/eks-anywhere/pkg/api/v1alpha1"
	"github.com/aws/eks-anywhere/pkg/cluster"
	"github.com/aws/eks-anywhere/pkg/constants"
	"github.com/aws/eks-anywhere/pkg/types"
	"github.com/aws/eks-anywhere/pkg/validations"
	"github.com/aws/eks-anywhere/pkg/validations/createvalidations"
	"github.com/aws/eks-anywhere/pkg/validations/mocks"
	releasev1alpha1 "github.com/aws/eks-anywhere/release/api/v1alpha1"
)

type preflightValidationsTest struct {
	*WithT
	ctx context.Context
	k   *mocks.MockKubectlClient
	c   *createvalidations.CreateValidations
}

func newPreflightValidationsTest(t *testing.T) *preflightValidationsTest {
	ctrl := gomock.NewController(t)
	k := mocks.NewMockKubectlClient(ctrl)
	c := &types.Cluster{
		KubeconfigFile: "kubeconfig",
	}
	clusterSpec := test.NewClusterSpec(func(s *cluster.Spec) {
		s.Cluster.Spec.GitOpsRef = &v1alpha1.Ref{
			Name: "gitops",
		}
	})
	opts := &validations.Opts{
		Kubectl:           k,
		Spec:              clusterSpec,
		WorkloadCluster:   c,
		ManagementCluster: c,
	}
	return &preflightValidationsTest{
		WithT: NewWithT(t),
		ctx:   context.Background(),
		k:     k,
		c:     createvalidations.New(opts),
	}
}

func TestPreFlightValidationsGitProvider(t *testing.T) {
	tt := newPreflightValidationsTest(t)
	tt.Expect(validations.ProcessValidationResults(tt.c.PreflightValidations(tt.ctx))).To(Succeed())
}

func TestPreFlightValidationsWorkloadCluster(t *testing.T) {
	tt := newPreflightValidationsTest(t)
	mgmtClusterName := "mgmt-cluster"
	tt.c.Opts.Spec.Cluster.SetManagedBy(mgmtClusterName)
	tt.c.Opts.Spec.Cluster.Spec.ManagementCluster.Name = mgmtClusterName
	tt.c.Opts.ManagementCluster.Name = mgmtClusterName

	mgmt := &v1alpha1.Cluster{
		ObjectMeta: v1.ObjectMeta{
			Name: "mgmt-cluster",
		},
		Spec: v1alpha1.ClusterSpec{
			ManagementCluster: v1alpha1.ManagementCluster{
				Name: "mgmt-cluster",
			},
			BundlesRef: &anywherev1.BundlesRef{
				Name:      "bundles-29",
				Namespace: constants.EksaSystemNamespace,
			},
		},
	}

	mgmtBundle := &releasev1alpha1.Bundles{
		Spec: releasev1alpha1.BundlesSpec{
			Number: tt.c.Opts.Spec.Bundles.Spec.Number + 1,
		},
	}

	tt.k.EXPECT().GetClusters(tt.ctx, tt.c.Opts.WorkloadCluster).Return(nil, nil)
	tt.k.EXPECT().ValidateClustersCRD(tt.ctx, tt.c.Opts.WorkloadCluster).Return(nil)
	tt.k.EXPECT().ValidateEKSAClustersCRD(tt.ctx, tt.c.Opts.WorkloadCluster).Return(nil)
	tt.k.EXPECT().GetEksaCluster(tt.ctx, tt.c.Opts.ManagementCluster, mgmtClusterName).Return(mgmt, nil)
	tt.k.EXPECT().GetEksaCluster(tt.ctx, tt.c.Opts.ManagementCluster, mgmtClusterName).Return(mgmt, nil)
	tt.k.EXPECT().GetBundles(tt.ctx, tt.c.Opts.ManagementCluster.KubeconfigFile, mgmt.Spec.BundlesRef.Name, mgmt.Spec.BundlesRef.Namespace).Return(mgmtBundle, nil)

	tt.Expect(validations.ProcessValidationResults(tt.c.PreflightValidations(tt.ctx))).To(Succeed())
}
