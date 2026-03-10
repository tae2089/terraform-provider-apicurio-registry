// Copyright IBM Corp. 2021, 2025
// SPDX-License-Identifier: MPL-2.0

package provider

import "net/http"

type ApicurioClient struct {
	HttpClient *http.Client
	Endpoint   string
}
