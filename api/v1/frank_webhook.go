/*
Copyright 2023.

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

package v1

import (
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/util/validation/field"
	ctrl "sigs.k8s.io/controller-runtime"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/webhook"
)

// log is for logging in this package.
var franklog = logf.Log.WithName("frank-resource")

func (r *Frank) SetupWebhookWithManager(mgr ctrl.Manager) error {
	return ctrl.NewWebhookManagedBy(mgr).
		For(r).
		Complete()
}

// TODO(user): EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!

//+kubebuilder:webhook:path=/mutate-apps-frank-com-v1-frank,mutating=true,failurePolicy=fail,sideEffects=None,groups=apps.frank.com,resources=franks,verbs=create;update,versions=v1,name=mfrank.kb.io,admissionReviewVersions=v1

var _ webhook.Defaulter = &Frank{}

// Default implements webhook.Defaulter so a webhook will be registered for the type
func (r *Frank) Default() {
	franklog.Info("default", "name", r.Name)
	if r.Spec.Replica == nil {
		r.Spec.Replica = new(int32)
		*r.Spec.Replica = 2
	}
	// TODO(user): fill in your defaulting logic.
}

// TODO(user): change verbs to "verbs=create;update;delete" if you want to enable deletion validation.
//+kubebuilder:webhook:path=/validate-apps-frank-com-v1-frank,mutating=false,failurePolicy=fail,sideEffects=None,groups=apps.frank.com,resources=franks,verbs=create;update,versions=v1,name=vfrank.kb.io,admissionReviewVersions=v1

var _ webhook.Validator = &Frank{}

// ValidateCreate implements webhook.Validator so a webhook will be registered for the type
func (r *Frank) ValidateCreate() error {
	franklog.Info("validate create", "name", r.Name)

	// TODO(user): fill in your validation logic upon object creation.
	return r.ValidateVerification()
}

// ValidateUpdate implements webhook.Validator so a webhook will be registered for the type
func (r *Frank) ValidateUpdate(old runtime.Object) error {
	franklog.Info("validate update", "name", r.Name)

	// TODO(user): fill in your validation logic upon object update.
	return r.ValidateVerification()
}

// ValidateDelete implements webhook.Validator so a webhook will be registered for the type
func (r *Frank) ValidateDelete() error {
	franklog.Info("validate delete", "name", r.Name)

	// TODO(user): fill in your validation logic upon object deletion.
	return nil
}

func (r *Frank) ValidateVerification() error {
	var allErrs field.ErrorList
	if r.Spec.Image == nil {
		err := field.Invalid(field.NewPath("spec").Child("image"),
			r.Spec.Image,
			"The value cannot be empty, please check your value")
		allErrs = append(allErrs, err)
	}
	if r.Spec.Pdl == nil {
		err := field.Invalid(field.NewPath("spec").Child("pdl"),
			r.Spec.Pdl,
			"The value cannot be empty, please check your value")
		allErrs = append(allErrs, err)
	}
	if len(allErrs) == 0 {
		return nil
	}
	return apierrors.NewInvalid(
		schema.GroupKind{Group: "apps.frank.com", Kind: "Frank"},
		r.Name, allErrs)
}
