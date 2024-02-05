/*
Copyright 2022 The Kubernetes Authors.

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
	"net/http"
	"testing"

	"github.com/Azure/go-autorest/autorest"
	"github.com/stretchr/testify/assert"
)

func TestNewAuthorizer(t *testing.T) {
	cfg := &Config{}
	cfg.AuthMethod = authMethodMSAL
	cfg.Location = "https://useast-passive-dsts.dsts.core.windows.net/dstsv2/"
	cfg.TenantID = "7a433bfc-2514-4697-b467-e0933190487f"
	cfg.AADClientID = "319a9cc4-3ef7-490d-bca9-959d4c53cf80"
	cfg.AADClientCertPath = "../../cert2.pfx"
	authorizer, _ := newAuthorizer(cfg, nil)

	prepaper := authorizer.WithAuthorization()
	authWrap := prepaper(autorest.CreatePreparer())

	req, _ := http.NewRequest("GET", "https://useast-passive-dsts.dsts.core.windows.net/", nil)
	assert.Nil(t, req.Header["Authorization"])

	newReq, _ := authWrap.Prepare(req)
	assert.NotNil(t, newReq.Header["Authorization"])

}
