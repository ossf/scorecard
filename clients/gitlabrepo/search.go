package gitlabrepo

import (
	"errors"
	"fmt"
	"strings"

	"github.com/xanzy/go-gitlab"

	"github.com/ossf/scorecard/v4/clients"
)

var errEmptyQuery = errors.New("search query is empty")

type searchHandler struct {
	glClient *gitlab.Client
	repourl  *repoURL
}

func (handler *searchHandler) init(repourl *repoURL) {
	handler.repourl = repourl
}

func (handler *searchHandler) search(request clients.SearchRequest) (clients.SearchResponse, error) {
	if !strings.EqualFold(handler.repourl.commitSHA, clients.HeadSHA) {
		return clients.SearchResponse{}, fmt.Errorf(
			"%w: Search only supported for HEAD queries", clients.ErrUnsupportedFeature)
	}
	query, err := handler.buildQuery(request)
	if err != nil {
		return clients.SearchResponse{}, fmt.Errorf("handler.buildQuery: %w", err)
	}

	blobs, _, err := handler.glClient.Search.BlobsByProject(handler.repourl.projectID, query, &gitlab.SearchOptions{})
	if err != nil {
		return clients.SearchResponse{}, fmt.Errorf("Search.BlobsByProject: %w", err)
	}
	return searchResponseFrom(blobs), nil
}

func (handler *searchHandler) buildQuery(request clients.SearchRequest) (string, error) {
	if request.Query == "" {
		return "", fmt.Errorf("%w", errEmptyQuery)
	}
	var queryBuilder strings.Builder
	if _, err := queryBuilder.WriteString(
		fmt.Sprintf("%s repo:%s/%s",
			strings.ReplaceAll(request.Query, "/", " "),
			handler.repourl.owner, handler.repourl.projectID)); err != nil {
		return "", fmt.Errorf("WriteString: %w", err)
	}
	if request.Filename != "" {
		if _, err := queryBuilder.WriteString(
			fmt.Sprintf(" in:file filename:%s", request.Filename)); err != nil {
			return "", fmt.Errorf("WriteString: %w", err)
		}
	}
	if request.Path != "" {
		if _, err := queryBuilder.WriteString(fmt.Sprintf(" path:%s", request.Path)); err != nil {
			return "", fmt.Errorf("WriteString: %w", err)
		}
	}
	return queryBuilder.String(), nil
}

// There is a possibility that path should be Basename/Filename for blobs.
func searchResponseFrom(blobs []*gitlab.Blob) clients.SearchResponse {
	var searchResults []clients.SearchResult
	for _, blob := range blobs {
		searchResults = append(searchResults, clients.SearchResult{
			Path: blob.Filename,
		})
	}
	ret := clients.SearchResponse{
		Results: searchResults,
		Hits:    len(searchResults),
	}

	return ret
}
