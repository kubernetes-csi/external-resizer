/*
Copyright The Kubernetes Authors.

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

// Code generated by client-gen. DO NOT EDIT.

package fake

import (
	v1 "k8s.io/api/core/v1"
	corev1 "k8s.io/client-go/applyconfigurations/core/v1"
	gentype "k8s.io/client-go/gentype"
	typedcorev1 "k8s.io/client-go/kubernetes/typed/core/v1"
)

// fakeResourceQuotas implements ResourceQuotaInterface
type fakeResourceQuotas struct {
	*gentype.FakeClientWithListAndApply[*v1.ResourceQuota, *v1.ResourceQuotaList, *corev1.ResourceQuotaApplyConfiguration]
	Fake *FakeCoreV1
}

func newFakeResourceQuotas(fake *FakeCoreV1, namespace string) typedcorev1.ResourceQuotaInterface {
	return &fakeResourceQuotas{
		gentype.NewFakeClientWithListAndApply[*v1.ResourceQuota, *v1.ResourceQuotaList, *corev1.ResourceQuotaApplyConfiguration](
			fake.Fake,
			namespace,
			v1.SchemeGroupVersion.WithResource("resourcequotas"),
			v1.SchemeGroupVersion.WithKind("ResourceQuota"),
			func() *v1.ResourceQuota { return &v1.ResourceQuota{} },
			func() *v1.ResourceQuotaList { return &v1.ResourceQuotaList{} },
			func(dst, src *v1.ResourceQuotaList) { dst.ListMeta = src.ListMeta },
			func(list *v1.ResourceQuotaList) []*v1.ResourceQuota { return gentype.ToPointerSlice(list.Items) },
			func(list *v1.ResourceQuotaList, items []*v1.ResourceQuota) {
				list.Items = gentype.FromPointerSlice(items)
			},
		),
		fake,
	}
}
