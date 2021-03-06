/**
 * Copyright 2021 Opstrace, Inc.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *      http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package controllers

import (
	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func WithPodAntiAffinity(name string) *corev1.Affinity {
	return &corev1.Affinity{
		PodAntiAffinity: &corev1.PodAntiAffinity{
			PreferredDuringSchedulingIgnoredDuringExecution: []corev1.WeightedPodAffinityTerm{
				{
					PodAffinityTerm: corev1.PodAffinityTerm{
						LabelSelector: &v1.LabelSelector{
							MatchLabels: map[string]string{"name": name},
						},
						TopologyKey: "kubernetes.io/hostname",
					},
					//
					// https://kubernetes.io/docs/concepts/scheduling-eviction/assign-pod-node/
					//
					// "The weight field in preferredDuringSchedulingIgnoredDuringExecution
					// is in the range 1-100. For each node that meets all of the scheduling
					// requirements (resource request, RequiredDuringScheduling affinity
					// expressions, etc.), the scheduler will compute a sum by iterating
					// through the elements of this field and adding "weight" to the sum if
					// the node matches the corresponding MatchExpressions. This score is
					// then combined with the scores of other priority functions for the
					// node. The node(s) with the highest total score are the most
					// preferred."
					//
					// Give the maximum weight to this rule to prevent the pod from being
					// scheduled to a node already running a pod for the given
					// deployment.
					Weight: 100,
				},
			},
		},
	}
}
