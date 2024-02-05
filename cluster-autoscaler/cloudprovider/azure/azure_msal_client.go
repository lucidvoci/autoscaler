/*
Copyright 2018 The Kubernetes Authors.

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

package azure

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"os"

	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	"github.com/Azure/go-autorest/autorest"
	"github.com/AzureAD/microsoft-authentication-library-for-go/apps/confidential"
)

type MSALBearerAuthorizer struct {
	client   confidential.Client
	tenantID string
}

// WithAuthorization returns a PrepareDecorator that adds an HTTP Authorization header whose
// value is "Bearer " followed by the token.
func (ba *MSALBearerAuthorizer) WithAuthorization() autorest.PrepareDecorator {
	return func(p autorest.Preparer) autorest.Preparer {
		return autorest.PreparerFunc(func(r *http.Request) (*http.Request, error) {
			// TODO: get refresh token when available
			token, errT := ba.client.AcquireTokenByCredential(
				context.Background(),
				[]string{"https://management.azure.com/.default"},
				confidential.WithTenantID(ba.tenantID),
			)

			if errT != nil {
				return nil, fmt.Errorf("failed to retrieve a token: %w", errT)
			}

			r, err := p.Prepare(r)
			if err == nil {
				return autorest.Prepare(r, autorest.WithHeader("Authorization", fmt.Sprintf("Bearer %s", token.AccessToken)))
			}
			return r, err
		})
	}
}

func NewMSALClient(config *Config) (confidential.Client, error) {
	b, err := os.ReadFile(config.AADClientCertPath)
	if err != nil {
		return confidential.Client{}, fmt.Errorf("couldn't read cert file %s: %w", config.AADClientCertPath, err)
	}

	cert, key, err := azidentity.ParseCertificates(b, nil)
	if err != nil {
		return confidential.Client{}, fmt.Errorf("couldn't parse certificate: %w", err)
	}

	authorityUrl, err := url.JoinPath(config.Location, config.TenantID)
	if err != nil {
		return confidential.Client{}, fmt.Errorf("failed to create authorityUrl: %w", err)
	}

	cred, err := confidential.NewCredFromCert(cert, key)
	if err != nil {
		return confidential.Client{}, fmt.Errorf("failed to create creds object from certificate+key: %w", err)
	}

	opts := []confidential.Option{
		// confidential.WithAzureRegion(confidential.AutoDetectRegion()),
		// confidential.WithKnownAuthorityHosts([]string{authorityHost}), // tmp
		// confidential.WithInstanceDiscovery(false),                     // tmp
	}

	opts = append(opts, confidential.WithX5C())

	client, err := confidential.New(authorityUrl, config.AADClientID, cred, opts...)
	if err != nil {
		return confidential.Client{}, fmt.Errorf("failed to create confidential client: %w", err)
	}
	return client, err
}
