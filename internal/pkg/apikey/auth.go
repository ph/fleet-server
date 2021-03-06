// Licensed to Elasticsearch B.V. under one or more contributor
// license agreements. See the NOTICE file distributed with
// this work for additional information regarding copyright
// ownership. Elasticsearch B.V. licenses this file to you under
// the Apache License, Version 2.0 (the "License"); you may
// not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing,
// software distributed under the License is distributed on an
// "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY
// KIND, either express or implied.  See the License for the
// specific language governing permissions and limitations
// under the License.

package apikey

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/elastic/go-elasticsearch/v8"
	"github.com/elastic/go-elasticsearch/v8/esapi"
)

type SecurityInfo struct {
	UserName    string            `json:"username"`
	Roles       []string          `json:"roles"`
	FullName    string            `json:"full_name"`
	Email       string            `json:"email"`
	Metadata    json.RawMessage   `json:"metadata"`
	Enabled     bool              `json:"enabled"`
	AuthRealm   map[string]string `json:"authentication_realm"`
	LookupRealm map[string]string `json:"lookup_realm"`
}

// Kibana:
// https://github.com/elastic/kibana/blob/master/x-pack/plugins/security/server/authentication/authenticator.ts#L308
// NOTE: Bulk request currently not available.
func (k ApiKey) Authenticate(ctx context.Context, es *elasticsearch.Client) (*SecurityInfo, error) {

	// TODO: Escape request for safety.  Don't depend on ES.

	token := fmt.Sprintf("%s%s", authPrefix, k.Token())

	req := esapi.SecurityAuthenticateRequest{
		Header: map[string][]string{AuthKey: []string{token}},
	}

	res, err := req.Do(ctx, es)

	if err != nil {
		return nil, err
	}

	if res.Body != nil {
		defer res.Body.Close()
	}

	if res.IsError() {
		return nil, fmt.Errorf("Fail Auth: %s", res.String())
	}

	var info SecurityInfo
	decoder := json.NewDecoder(res.Body)
	if err := decoder.Decode(&info); err != nil {
		return nil, fmt.Errorf("Auth: error parsing response body: %s", err) // TODO: Wrap error
	}

	return &info, nil
}
