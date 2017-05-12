package predicates

import (
	"reflect"
	"testing"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/kubernetes/pkg/api/v1"
	"k8s.io/kubernetes/plugin/pkg/scheduler/algorithm"
	"k8s.io/kubernetes/plugin/pkg/scheduler/schedulercache"
	schedulertesting "k8s.io/kubernetes/plugin/pkg/scheduler/testing"
)

func TestBUG2676(t *testing.T) {
	podLabelA := map[string]string{
		"app": "nginx",
	}

	labelM1 := map[string]string{
		"kubernetes.io/hostname": "machine1",
	}

	labelM2 := map[string]string{
		"kubernetes.io/hostname": "machine2",
	}

	labelM3 := map[string]string{
		"kubernetes.io/hostname": "machine3",
	}

	tests := []struct {
		pod    *v1.Pod
		pods   []*v1.Pod
		nodes  []v1.Node
		fits   map[string]bool
		test   string
		nometa bool
	}{
		{
			pod: &v1.Pod{
				Spec: v1.PodSpec{
					Affinity: &v1.Affinity{
						PodAntiAffinity: &v1.PodAntiAffinity{
							RequiredDuringSchedulingIgnoredDuringExecution: []v1.PodAffinityTerm{
								{
									LabelSelector: &metav1.LabelSelector{
										MatchExpressions: []metav1.LabelSelectorRequirement{
											{
												Key:      "app",
												Operator: metav1.LabelSelectorOpIn,
												Values:   []string{"nginx"},
											},
										},
									},
									TopologyKey: "kubernetes.io/hostname",
								},
							},
						},
					},
				},
				ObjectMeta: metav1.ObjectMeta{Labels: podLabelA, Namespace: "NS", Name: "P4"},
			},
			pods: []*v1.Pod{
				{Spec: v1.PodSpec{
					Affinity: &v1.Affinity{
						PodAntiAffinity: &v1.PodAntiAffinity{
							RequiredDuringSchedulingIgnoredDuringExecution: []v1.PodAffinityTerm{
								{
									LabelSelector: &metav1.LabelSelector{
										MatchExpressions: []metav1.LabelSelectorRequirement{
											{
												Key:      "app",
												Operator: metav1.LabelSelectorOpIn,
												Values:   []string{"nginx"},
											},
										},
									},
									TopologyKey: "kubernetes.io/hostname",
								},
							},
						},
					},
					NodeName: "machine1",
					Hostname: "machine1",
				}, ObjectMeta: metav1.ObjectMeta{Name: "P1", Labels: podLabelA, Namespace: "NS"}},
				{Spec: v1.PodSpec{
					Affinity: &v1.Affinity{
						PodAntiAffinity: &v1.PodAntiAffinity{
							RequiredDuringSchedulingIgnoredDuringExecution: []v1.PodAffinityTerm{
								{
									LabelSelector: &metav1.LabelSelector{
										MatchExpressions: []metav1.LabelSelectorRequirement{
											{
												Key:      "app",
												Operator: metav1.LabelSelectorOpIn,
												Values:   []string{"nginx"},
											},
										},
									},
									TopologyKey: "hostname",
								},
							},
						},
					},
					NodeName: "machine2",
					Hostname: "machine2",
				}, ObjectMeta: metav1.ObjectMeta{Name: "P2", Labels: podLabelA}},
				{Spec: v1.PodSpec{
					Affinity: &v1.Affinity{
						PodAntiAffinity: &v1.PodAntiAffinity{
							RequiredDuringSchedulingIgnoredDuringExecution: []v1.PodAffinityTerm{
								{
									LabelSelector: &metav1.LabelSelector{
										MatchExpressions: []metav1.LabelSelectorRequirement{
											{
												Key:      "app",
												Operator: metav1.LabelSelectorOpIn,
												Values:   []string{"nginx"},
											},
										},
									},
									TopologyKey: "kubernetes.io/hostname",
								},
							},
						},
					},
					NodeName: "machine3",
					Hostname: "machine3",
				}, ObjectMeta: metav1.ObjectMeta{Name: "P3", Labels: podLabelA}},
			},
			nodes: []v1.Node{
				{ObjectMeta: metav1.ObjectMeta{Name: "machine1", Labels: labelM1}},
				{ObjectMeta: metav1.ObjectMeta{Name: "machine2", Labels: labelM2}},
				{ObjectMeta: metav1.ObjectMeta{Name: "machine3", Labels: labelM3}},
			},
			fits: map[string]bool{
				"machine1": false,
				"machine2": true,
				"machine3": true,
			},
			test:   "A pod can be scheduled onto all the nodes that have the same topology key & label value with one of them has an existing pod that match the affinity rules",
			nometa: false,
		},
	}
	affinityExpectedFailureReasons := []algorithm.PredicateFailureReason{ErrPodAffinityNotMatch}

	for _, test := range tests {
		nodeListInfo := FakeNodeListInfo(test.nodes)
		for _, node := range test.nodes {
			var podsOnNode []*v1.Pod
			for _, pod := range test.pods {
				if pod.Spec.NodeName == node.Name {
					podsOnNode = append(podsOnNode, pod)
				}
			}

			testFit := PodAffinityChecker{
				info:      nodeListInfo,
				podLister: schedulertesting.FakePodLister(test.pods),
			}
			nodeInfo := schedulercache.NewNodeInfo(podsOnNode...)
			nodeInfo.SetNode(&node)
			nodeInfoMap := map[string]*schedulercache.NodeInfo{node.Name: nodeInfo}
			var meta interface{}

			if !test.nometa {
				meta = PredicateMetadata(test.pod, nodeInfoMap)
			}

			fits, reasons, err := testFit.InterPodAffinityMatches(test.pod, meta, nodeInfo)
			if err != nil {
				t.Errorf("%s: unexpected error %v", test.test, err)
			}
			if !fits && !reflect.DeepEqual(reasons, affinityExpectedFailureReasons) {
				t.Errorf("%s: unexpected failure reasons: %v", test.test, reasons)
			}

			if fits != test.fits[node.Name] {
				t.Errorf("%s: expected %v for %s got %v", test.test, test.fits[node.Name], node.Name, fits)
			}
		}
	}
}
