/*
 * Copyright (C) 2019  Rohith Jayawardene <gambol99@gmail.com>
 *
 * This program is free software; you can redistribute it and/or
 * modify it under the terms of the GNU General Public License
 * as published by the Free Software Foundation; either version 2
 * of the License, or (at your option) any later version.
 *
 * This program is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 * GNU General Public License for more details.
 *
 * You should have received a copy of the GNU General Public License
 * along with this program.  If not, see <http://www.gnu.org/licenses/>.
 */

package main

import (
	"context"
	"fmt"

	gcpv1alpha1 "github.com/appvia/gcp-operator/pkg/apis/gcp/v1alpha1"
	"github.com/appvia/gcp-operator/pkg/apis/schema"

	configv1 "github.com/appvia/hub-apis/pkg/apis/config/v1"
	"github.com/appvia/hub-apis/pkg/publish"
	hschema "github.com/appvia/hub-apis/pkg/schema"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/rest"
)

var (
	// gkeClass is the provider class published into the hub
	gkeClass = configv1.Class{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "gcp",
			Namespace: "hub",
		},
		Spec: configv1.ClassSpec{
			APIVersion:    gcpv1alpha1.SchemeGroupVersion.String(),
			AutoProvision: false,
			Category:      "accounts",
			Description:   "Google Compute Services provides a means to provision the accounts and projects for cluster",
			DisplayName:   "Google Compute",
			Requires: metav1.GroupVersionKind{
				Group:   gcpv1alpha1.SchemeGroupVersion.Group,
				Kind:    "GCPAdminProject",
				Version: gcpv1alpha1.SchemeGroupVersion.Version,
			},
			Plans: []string{},
			Resources: []configv1.ClassResource{
				{
					Group:            gcpv1alpha1.SchemeGroupVersion.Group,
					Kind:             "GCPProject",
					Plans:            []string{},
					DisplayName:      "GCP Project",
					ShortDescription: "Provisions the GCP Project and permissions for cluster building",
					LongDescription:  "Provides the ability to provision the Google Compute project and permissions",
					Scope:            configv1.TeamScope,
					Version:          gcpv1alpha1.SchemeGroupVersion.Version,
				},
				{
					Group:            gcpv1alpha1.SchemeGroupVersion.Group,
					Kind:             "GCPCredentials",
					Plans:            []string{},
					DisplayName:      "GCP Credentials",
					ShortDescription: "The Google Compute credentials which can be used to provision clusters",
					LongDescription:  "The credentials used to provision cluters from",
					Scope:            configv1.TeamScope,
					Version:          gcpv1alpha1.SchemeGroupVersion.Version,
				},
			},
			Schemas: schema.ConvertToJSON(),
		},
	}
)

// publishOperator is responsible for injecting the classes configuration
// into the hub-api and crds
func publishOperator(cfg *rest.Config) error {
	// @step: publish the CRDs in the hub
	ac, err := publish.NewExtentionsAPIClient(cfg)
	if err != nil {
		return err
	}

	if publishCRDs {
		if err := publish.ApplyCustomResourceDefinitions(ac, schema.GetCustomResourceDefinitions()); err != nil {
			return fmt.Errorf("failed to register the operator crds: %s", err)
		}
	}

	if publishClasses {
		c, err := hschema.NewClient(cfg)
		if err != nil {
			return err
		}

		return hschema.PublishAll(context.TODO(), c, gkeClass, []configv1.Plan{})
	}

	return nil
}
