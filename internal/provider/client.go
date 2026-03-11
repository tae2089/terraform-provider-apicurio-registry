// Copyright IBM Corp. 2021, 2025
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"io"
	"net/http"
)

type ApicurioClient struct {
	HttpClient *http.Client
	Endpoint   string
	Token      string
}

func (c *ApicurioClient) NewRequest(ctx context.Context, method, url string, body io.Reader) (*http.Request, error) {
	req, err := http.NewRequestWithContext(ctx, method, url, body)
	if err != nil {
		return nil, err
	}

	if c.Token != "" {
		req.Header.Set("Authorization", "Bearer "+c.Token)
	}

	return req, nil
}

// v3 API Structs common to resources and data sources

type CreateArtifactRequest struct {
	ArtifactId   string                `json:"artifactId,omitempty"`
	ArtifactType string                `json:"artifactType,omitempty"`
	FirstVersion *CreateVersionRequest `json:"firstVersion,omitempty"`
}

type CreateVersionRequest struct {
	Version string           `json:"version,omitempty"`
	Content *ArtifactContent `json:"content"`
}

type CreateVersionResponse struct {
	Version  string `json:"version"`
	GlobalId int64  `json:"globalId"`
	State    string `json:"state"`
}

type ArtifactContent struct {
	Content     string `json:"content"`
	ContentType string `json:"contentType"`
}

type ArtifactMetaData struct {
	ArtifactId   string `json:"artifactId"`
	Id           string `json:"id"`
	GroupId      string `json:"groupId"`
	ArtifactType string `json:"artifactType"`
	Type         string `json:"type"`
}

type VersionMetaData struct {
	Version  string `json:"version"`
	GlobalId int64  `json:"globalId"`
	State    string `json:"state"`
}
