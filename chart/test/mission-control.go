package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"

	"github.com/flanksource/commons-db/kubernetes"
	"github.com/flanksource/commons/http"
)

type MissionControl struct {
	HTTP      *http.Client
	ConfigDB  *http.Client
	URL       string
	Username  string
	Password  string
	Namespace string
	*kubernetes.Client
	DB *sql.DB
}

func (mc *MissionControl) POST(path string, body interface{}) (*http.Response, error) {
	return mc.HTTP.R(context.TODO()).Post(path, body)
}

type Scraper struct {
	mc   *MissionControl
	Name string
	Id   string
}

type ScrapeResults struct {
	Response *http.Response
	Error    error
}

func (mc *MissionControl) GetScraper(id string) *Scraper {
	return &Scraper{
		mc:   mc,
		Id:   id,
		Name: "",
	}
}

type ScrapeResult struct {
	Errors  []string       `json:"errors"`
	Summary map[string]any `json:"scrape_summary"`
}

func (s *Scraper) Run() (*ScrapeResult, error) {
	r, err := s.mc.ConfigDB.R(context.Background()).Post("/run/"+s.Id, map[string]string{"scraper": s.Name})
	if err != nil {
		return nil, err
	}
	result := &ScrapeResult{}
	body, err := r.AsString()
	if err != nil {
		return nil, err
	}
	return result, json.Unmarshal([]byte(body), result)
}

type ResourceSelector struct {
	ID            string            `json:"id,omitempty"`
	Name          string            `json:"name,omitempty"`
	Namespace     string            `json:"namespace,omitempty"`
	Types         []string          `json:"types,omitempty"`
	Statuses      []string          `json:"statuses,omitempty"`
	Labels        map[string]string `json:"labels,omitempty"`
	FieldSelector string            `json:"field_selector,omitempty"`
	Search        string            `json:"search,omitempty"`
}

type SearchResourcesRequest struct {
	Limit   int                `json:"limit,omitempty"`
	Configs []ResourceSelector `json:"configs,omitempty"`
}

type SelectedResource struct {
	ID         string            `json:"id,omitempty"`
	Name       string            `json:"name,omitempty"`
	Namespace  string            `json:"namespace,omitempty"`
	Type       string            `json:"type,omitempty"`
	Tags       map[string]string `json:"tags,omitempty"`
	Config     string            `json:"config,omitempty"`
	Labels     map[string]string `json:"labels,omitempty"`
	Properties map[string]string `json:"properties,omitempty"`
}

type SearchResourcesResponse struct {
	Configs []SelectedResource `json:"configs"`
}

func (mc *MissionControl) QueryCatalog(selector ResourceSelector) ([]SelectedResource, error) {
	req := SearchResourcesRequest{
		Configs: []ResourceSelector{selector},
	}

	r, err := mc.HTTP.R(context.TODO()).Post("/resources/search", req)
	if err != nil {
		return nil, err
	}

	if !r.IsOK() {
		body, _ := r.AsString()
		return nil, fmt.Errorf("query catalog failed: %s", body)
	}

	var response SearchResourcesResponse
	if err := r.Into(&response); err != nil {
		return nil, err
	}

	return response.Configs, nil
}

func (mc *MissionControl) SearchCatalog(search string) ([]SelectedResource, error) {
	return mc.QueryCatalog(ResourceSelector{Search: search})
}
