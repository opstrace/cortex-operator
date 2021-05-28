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

package v1alpha1

import (
	"gopkg.in/yaml.v2"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/webhook"

	"github.com/cortexproject/cortex/pkg/cortex"
	"github.com/cortexproject/cortex/pkg/util/flagext"
)

// log is for logging in this package.
var cortexlog = logf.Log.WithName("cortex-resource")

func (r *Cortex) SetupWebhookWithManager(mgr ctrl.Manager) error {
	cortexlog.Info("webhook manager set up")
	return ctrl.NewWebhookManagedBy(mgr).
		For(r).
		Complete()
}

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!

//+kubebuilder:webhook:webhookVersions={v1beta1},path=/mutate-cortex-opstrace-io-v1alpha1-cortex,mutating=true,failurePolicy=fail,sideEffects=None,groups=cortex.opstrace.io,resources=cortices,verbs=create;update,versions=v1alpha1,name=mcortex.kb.io,admissionReviewVersions={v1,v1beta1}

var _ webhook.Defaulter = &Cortex{}

// Default implements webhook.Defaulter so a webhook will be registered for the type
func (r *Cortex) Default() {
	cortexlog.Info("default", "name", r.Name)

	// TODO(user): fill in your defaulting logic.
}

// TODO(user): change verbs to "verbs=create;update;delete" if you want to enable deletion validation.
//+kubebuilder:webhook:webhookVersions={v1beta1},path=/validate-cortex-opstrace-io-v1alpha1-cortex,mutating=false,failurePolicy=fail,sideEffects=None,groups=cortex.opstrace.io,resources=cortices,verbs=create;update,versions=v1alpha1,name=vcortex.kb.io,admissionReviewVersions={v1,v1beta1}

var _ webhook.Validator = &Cortex{}

// ValidateCreate implements webhook.Validator so a webhook will be registered for the type
func (r *Cortex) ValidateCreate() error {
	cortexlog.Info("validate create", "name", r.Name)

	cfg := &cortex.Config{}
	flagext.DefaultValues(cfg)

	err := yaml.UnmarshalStrict(r.Spec.Config.Raw, cfg)
	if err != nil {
		return err
	}

	return cfg.Validate(nil)
}

// ValidateUpdate implements webhook.Validator so a webhook will be registered for the type
func (r *Cortex) ValidateUpdate(old runtime.Object) error {
	cortexlog.Info("validate update", "name", r.Name)

	// TODO(user): fill in your validation logic upon object update.
	return nil
}

// ValidateDelete implements webhook.Validator so a webhook will be registered for the type
func (r *Cortex) ValidateDelete() error {
	cortexlog.Info("validate delete", "name", r.Name)

	// TODO(user): fill in your validation logic upon object deletion.
	return nil
}
