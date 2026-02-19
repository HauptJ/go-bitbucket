package bitbucket

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"strconv"
	"time"

	"github.com/mitchellh/mapstructure"
)

type PullRequests struct {
	c                     *Client
	ID                    int
	Title                 string
	Author                User
	Description           string
	CloseSourceBranch     bool
	SourceBranch          RepositoryBranch
	SourceRepository      Repository
	SourceCommit          string
	DestinationBranch     RepositoryBranch
	DestinationRepository Repository
	DestinationCommit     string
	Reviewers             []User
	State                 string
	Draft                 bool
	Commit                string
	Comments              []PullRequestsComments
	CreatedOnTime         *time.Time `mapstructure:"created_on"`
	UpdatedOnTime         *time.Time `mapstructure:"updated_on"`
}

type PullRequestsList struct {
	Page     int
	Pagelen  int
	MaxDepth int
	Size     int
	Next     string
	Items    []PullRequests
}

type PullRequestsComments struct {
	Owner         string
	RepoSlug      string
	PullRequestID int64
	Content       PullRequestCommentsContent
	CommentId     int64 `mapstructure:"id"`
	Parent        *int
}

type PullRequestCommentsContent struct {
	Raw    string "json:\"raw\",omitempty"
	Markup string "json:\"markup\",omitempty"
	Html   string "json:\"html\",omitempty"
}

func (c PullRequestCommentsContent) ContentString() string {
	return fmt.Sprintf(`"content": {"raw": "%s", "markup": "%s", "html": "%s"}`, c.Raw, c.Markup, c.Html)
}

type PullRequestsCommentsList struct {
	Page     int
	Pagelen  int
	MaxDepth int
	Size     int
	Next     string
	Items    []PullRequestsComments
}

func (p *PullRequests) Create(po *PullRequestOptions) (interface{}, error) {
	data, err := p.buildPullRequestBody(po)
	if err != nil {
		return nil, err
	}
	urlStr := p.c.requestUrl("/repositories/%s/%s/pullrequests/", po.Owner, po.RepoSlug)
	return p.c.executeWithContext("POST", urlStr, data, po.ctx)
}

func (p *PullRequests) Update(po *PullRequestOptions) (interface{}, error) {
	data, err := p.buildPullRequestBody(po)
	if err != nil {
		return nil, err
	}
	urlStr := p.c.GetApiBaseURL() + "/repositories/" + po.Owner + "/" + po.RepoSlug + "/pullrequests/" + po.ID
	return p.c.execute("PUT", urlStr, data)
}

func (p *PullRequests) GetByCommit(po *PullRequestOptions) (interface{}, error) {
	urlStr := p.c.GetApiBaseURL() + "/repositories/" + po.Owner + "/" + po.RepoSlug + "/commit/" + po.Commit + "/pullrequests/"
	return p.c.executePaginated("GET", urlStr, "", nil)
}

func (p *PullRequests) GetCommits(po *PullRequestOptions) (interface{}, error) {
	urlStr := p.c.GetApiBaseURL() + "/repositories/" + po.Owner + "/" + po.RepoSlug + "/pullrequests/" + po.ID + "/commits/"
	return p.c.executePaginated("GET", urlStr, "", nil)
}

func (p *PullRequests) List(po *PullRequestsOptions) (interface{}, error) {
	urlStr := p.c.GetApiBaseURL() + "/repositories/" + url.PathEscape(po.Owner) + "/" + url.PathEscape(po.RepoSlug) + "/pullrequests/"

	if po.States != nil && len(po.States) != 0 {
		parsed, err := url.Parse(urlStr)
		if err != nil {
			return nil, err
		}
		query := parsed.Query()
		for _, state := range po.States {
			query.Set("state", state)
		}
		parsed.RawQuery = query.Encode()
		urlStr = parsed.String()
	}

	if po.Query != "" {
		parsed, err := url.Parse(urlStr)
		if err != nil {
			return nil, err
		}
		query := parsed.Query()
		query.Set("q", po.Query)
		parsed.RawQuery = query.Encode()
		urlStr = parsed.String()
	}

	if po.Sort != "" {
		parsed, err := url.Parse(urlStr)
		if err != nil {
			return nil, err
		}
		query := parsed.Query()
		query.Set("sort", po.Sort)
		parsed.RawQuery = query.Encode()
		urlStr = parsed.String()
	}
	return p.c.executePaginated("GET", urlStr, "", nil)
}

// Append comments to each pull request object in the list with pass by reference
func (p *PullRequests) appendComments(po *PullRequestsOptions, prList **PullRequestsList) error {

	for _, pr := range (*prList).Items {
		prCommentOpts := &PullRequestCommentOptions{
			Owner:         po.Owner,
			RepoSlug:      po.RepoSlug,
			PullRequestID: strconv.Itoa(pr.ID),
		}
		prCommentsList, err := p.ListCommentsObjs(prCommentOpts)
		if err != nil {
			return err
		}
		pr.Comments = append(pr.Comments, prCommentsList.Items...)
	}
	return nil
}

func (p *PullRequests) ListObjs(po *PullRequestsOptions) (*PullRequestsList, error) {
	res, err := p.List(po)
	if err != nil {
		return nil, err
	}
	pullRequestListObjs, err := decodePullRequestsList(res)
	if err != nil {
		return nil, err
	}

	err = p.appendComments(po, &pullRequestListObjs)
	if err != nil {
		return nil, err
	}

	return pullRequestListObjs, nil
}

/*
Redirect to List() which is the function name declared in bitbucket.go
by Yoshimatsu on 6/10/19. This is to prevent breakage for anyone using
the missnamed Gets() function call.
*/
func (p *PullRequests) Gets(po *PullRequestsOptions) (interface{}, error) {
	return p.List(po)
}

func (p *PullRequests) Get(po *PullRequestOptions) (interface{}, error) {
	urlStr := p.c.GetApiBaseURL() + "/repositories/" + po.Owner + "/" + po.RepoSlug + "/pullrequests/" + po.ID
	return p.c.execute("GET", urlStr, "")
}

func (p *PullRequests) Activities(po *PullRequestOptions) (interface{}, error) {
	urlStr := p.c.GetApiBaseURL() + "/repositories/" + po.Owner + "/" + po.RepoSlug + "/pullrequests/activity"
	return p.c.executePaginated("GET", urlStr, "", nil)
}

func (p *PullRequests) Activity(po *PullRequestOptions) (interface{}, error) {
	urlStr := p.c.GetApiBaseURL() + "/repositories/" + po.Owner + "/" + po.RepoSlug + "/pullrequests/" + po.ID + "/activity"
	return p.c.execute("GET", urlStr, "")
}

func (p *PullRequests) Commits(po *PullRequestOptions) (interface{}, error) {
	urlStr := p.c.GetApiBaseURL() + "/repositories/" + po.Owner + "/" + po.RepoSlug + "/pullrequests/" + po.ID + "/commits"
	return p.c.executePaginated("GET", urlStr, "", nil)
}

func (p *PullRequests) Patch(po *PullRequestOptions) (interface{}, error) {
	urlStr := p.c.GetApiBaseURL() + "/repositories/" + po.Owner + "/" + po.RepoSlug + "/pullrequests/" + po.ID + "/patch"
	return p.c.executeRaw("GET", urlStr, "")
}

func (p *PullRequests) Diff(po *PullRequestOptions) (interface{}, error) {
	urlStr := p.c.GetApiBaseURL() + "/repositories/" + po.Owner + "/" + po.RepoSlug + "/pullrequests/" + po.ID + "/diff"
	return p.c.executeRaw("GET", urlStr, "")
}

func (p *PullRequests) Merge(po *PullRequestOptions) (interface{}, error) {
	data, err := p.buildPullRequestBody(po)
	if err != nil {
		return nil, err
	}
	urlStr := p.c.GetApiBaseURL() + "/repositories/" + po.Owner + "/" + po.RepoSlug + "/pullrequests/" + po.ID + "/merge"
	return p.c.executeWithContext("POST", urlStr, data, po.ctx)
}

func (p *PullRequests) Decline(po *PullRequestOptions) (interface{}, error) {
	data, err := p.buildPullRequestBody(po)
	if err != nil {
		return nil, err
	}
	urlStr := p.c.GetApiBaseURL() + "/repositories/" + po.Owner + "/" + po.RepoSlug + "/pullrequests/" + po.ID + "/decline"
	return p.c.executeWithContext("POST", urlStr, data, po.ctx)
}

func (p *PullRequests) Approve(po *PullRequestOptions) (interface{}, error) {
	urlStr := p.c.GetApiBaseURL() + "/repositories/" + po.Owner + "/" + po.RepoSlug + "/pullrequests/" + po.ID + "/approve"
	return p.c.executeWithContext("POST", urlStr, "", po.ctx)
}

func (p *PullRequests) UnApprove(po *PullRequestOptions) (interface{}, error) {
	urlStr := p.c.GetApiBaseURL() + "/repositories/" + po.Owner + "/" + po.RepoSlug + "/pullrequests/" + po.ID + "/approve"
	return p.c.execute("DELETE", urlStr, "")
}

func (p *PullRequests) RequestChanges(po *PullRequestOptions) (interface{}, error) {
	urlStr := p.c.GetApiBaseURL() + "/repositories/" + po.Owner + "/" + po.RepoSlug + "/pullrequests/" + po.ID + "/request-changes"
	return p.c.executeWithContext("POST", urlStr, "", po.ctx)
}

func (p *PullRequests) UnRequestChanges(po *PullRequestOptions) (interface{}, error) {
	urlStr := p.c.GetApiBaseURL() + "/repositories/" + po.Owner + "/" + po.RepoSlug + "/pullrequests/" + po.ID + "/request-changes"
	return p.c.execute("DELETE", urlStr, "")
}

func (p *PullRequests) AddComment(co *PullRequestCommentOptions) (interface{}, error) {
	data, err := p.buildPullRequestCommentBody(co)
	if err != nil {
		return nil, err
	}

	urlStr := p.c.requestUrl("/repositories/%s/%s/pullrequests/%s/comments", co.Owner, co.RepoSlug, co.PullRequestID)
	return p.c.executeWithContext("POST", urlStr, data, co.ctx)
}

func (p *PullRequests) AddCommentObj(opt PullRequestCommentOptions) (*PullRequestsComments, error) {
	res, err := p.AddComment(&opt)
	if err != nil {
		return nil, err
	}
	return decodePullRequestsComments(res)
}

func (p *PullRequests) UpdateComment(co *PullRequestCommentOptions) (interface{}, error) {
	data, err := p.buildPullRequestCommentBody(co)
	if err != nil {
		return nil, err
	}

	urlStr := p.c.requestUrl("/repositories/%s/%s/pullrequests/%s/comments/%s", co.Owner, co.RepoSlug, co.PullRequestID, co.CommentId)
	return p.c.execute("PUT", urlStr, data)
}

func (p *PullRequests) DeleteComment(co *PullRequestCommentOptions) (interface{}, error) {
	urlStr := p.c.requestUrl("/repositories/%s/%s/pullrequests/%s/comments/%s", co.Owner, co.RepoSlug, co.PullRequestID, co.CommentId)
	return p.c.execute("DELETE", urlStr, "")
}

func (p *PullRequests) ListComments(po *PullRequestCommentOptions) (interface{}, error) {
	urlStr := p.c.GetApiBaseURL() + "/repositories/" + po.Owner + "/" + po.RepoSlug + "/pullrequests/" + po.PullRequestID + "/comments/"
	return p.c.executePaginated("GET", urlStr, "", nil)
}

func (p *PullRequests) ListCommentsObjs(po *PullRequestCommentOptions) (*PullRequestsCommentsList, error) {
	commentsList, err := p.ListComments(po)
	if err != nil {
		return nil, err
	}

	commentsListObjs, err := decodePullRequestsCommentsList(commentsList)
	if err != nil {
		return nil, err
	}

	return commentsListObjs, nil
}

/*
Redirect to ListComments() which is the function name declared in bitbucket.go for consisitancy
This is to prevent breakage for anyone using the missnamed GetComments() function call.
*/
func (p *PullRequests) GetComments(po *PullRequestCommentOptions) (interface{}, error) {
	return p.ListComments(po)
}

func (p *PullRequests) GetComment(po *PullRequestCommentOptions) (interface{}, error) {
	urlStr := p.c.GetApiBaseURL() + "/repositories/" + po.Owner + "/" + po.RepoSlug + "/pullrequests/" + po.PullRequestID + "/comments/" + po.CommentId
	return p.c.execute("GET", urlStr, "")
}

func (p *PullRequests) GetCommentObj(po *PullRequestCommentOptions) (*PullRequestsComments, error) {
	prComment, err := p.GetComment(po)
	if err != nil {
		return nil, err
	}
	return decodePullRequestsComments(prComment)
}

func (p *PullRequests) Statuses(po *PullRequestOptions) (interface{}, error) {
	urlStr := p.c.GetApiBaseURL() + "/repositories/" + po.Owner + "/" + po.RepoSlug + "/pullrequests/" + po.ID + "/statuses"
	if po.Query != "" {
		parsed, err := url.Parse(urlStr)
		if err != nil {
			return nil, err
		}
		query := parsed.Query()
		query.Set("q", po.Query)
		parsed.RawQuery = query.Encode()
		urlStr = parsed.String()
	}

	if po.Sort != "" {
		parsed, err := url.Parse(urlStr)
		if err != nil {
			return nil, err
		}
		query := parsed.Query()
		query.Set("sort", po.Sort)
		parsed.RawQuery = query.Encode()
		urlStr = parsed.String()
	}
	return p.c.executePaginated("GET", urlStr, "", nil)
}

func (p *PullRequests) buildPullRequestBody(po *PullRequestOptions) (string, error) {
	body := map[string]interface{}{}
	body["source"] = map[string]interface{}{}
	body["destination"] = map[string]interface{}{}
	body["reviewers"] = []map[string]string{}
	body["title"] = ""
	body["description"] = ""
	body["message"] = ""
	body["close_source_branch"] = false

	if n := len(po.Reviewers); n > 0 {
		body["reviewers"] = make([]map[string]string, n)
		for i, uuid := range po.Reviewers {
			body["reviewers"].([]map[string]string)[i] = map[string]string{"uuid": uuid}
		}
	}

	if po.SourceBranch != "" {
		body["source"].(map[string]interface{})["branch"] = map[string]string{"name": po.SourceBranch}
	}

	if po.SourceRepository != "" {
		body["source"].(map[string]interface{})["repository"] = map[string]interface{}{"full_name": po.SourceRepository}
	}

	if po.DestinationBranch != "" {
		body["destination"].(map[string]interface{})["branch"] = map[string]interface{}{"name": po.DestinationBranch}
	}

	if po.DestinationCommit != "" {
		body["destination"].(map[string]interface{})["commit"] = map[string]interface{}{"hash": po.DestinationCommit}
	}

	if po.Title != "" {
		body["title"] = po.Title
	}

	if po.Description != "" {
		body["description"] = po.Description
	}

	if po.Message != "" {
		body["message"] = po.Message
	}

	if po.CloseSourceBranch || !po.CloseSourceBranch {
		body["close_source_branch"] = po.CloseSourceBranch
	}

	if po.Draft {
		body["draft"] = true
	}

	data, err := json.Marshal(body)
	if err != nil {
		return "", err
	}

	return string(data), nil
}

func (p *PullRequests) buildPullRequestCommentBody(co *PullRequestCommentOptions) (string, error) {
	body := map[string]interface{}{}
	body["content"] = map[string]interface{}{
		"raw": co.Content,
	}

	if co.Parent != nil {
		body["parent"] = map[string]interface{}{
			"id": co.Parent,
		}
	}

	data, err := json.Marshal(body)
	if err != nil {
		return "", err
	}

	return string(data), nil
}

func decodePullRequestsList(prResponse interface{}) (*PullRequestsList, error) {
	prResponseMap, ok := prResponse.(map[string]interface{})
	if !ok {
		return nil, errors.New("Not a valid format")
	}

	prArray := prResponseMap["values"].([]interface{})
	var prs []PullRequests
	for _, prEntry := range prArray {
		pr, err := decodePullRequests(prEntry)
		if err == nil {
			prs = append(prs, *pr)
		}
	}

	page, ok := prResponseMap["page"].(float64)
	if !ok {
		page = 0
	}

	pagelen, ok := prResponseMap["pagelen"].(float64)
	if !ok {
		pagelen = 0
	}
	size, ok := prResponseMap["size"].(float64)
	if !ok {
		size = 0
	}

	pullRequestList := PullRequestsList{
		Page:    int(page),
		Pagelen: int(pagelen),
		Size:    int(size),
		Items:   prs,
	}
	return &pullRequestList, nil
}

func decodePullRequests(pullRequestResp interface{}) (*PullRequests, error) {
	prMap := pullRequestResp.(map[string]interface{})

	if prMap["type"] == "error" {
		return nil, DecodeError(prMap)
	}

	var pullRequests = new(PullRequests)
	decoder, err := mapstructure.NewDecoder(&mapstructure.DecoderConfig{
		Metadata:   nil,
		Result:     pullRequests,
		DecodeHook: stringToTimeHookFunc,
	})
	if err != nil {
		return nil, err
	}
	err = decoder.Decode(prMap)
	if err != nil {
		return nil, err
	}

	return pullRequests, nil
}

func decodePullRequestsCommentsList(prCommentsResponse interface{}) (*PullRequestsCommentsList, error) {
	prCommentsResponseMap, ok := prCommentsResponse.(map[string]interface{})
	if !ok {
		return nil, errors.New("Not a valid format")
	}

	prCommentsArray := prCommentsResponseMap["values"].([]interface{})
	var prCommentsList []PullRequestsComments
	for _, prEntry := range prCommentsArray {
		pr, err := decodePullRequestsComments(prEntry)
		if err == nil {
			prCommentsList = append(prCommentsList, *pr)
		}
	}

	page, ok := prCommentsResponseMap["page"].(float64)
	if !ok {
		page = 0
	}

	pagelen, ok := prCommentsResponseMap["pagelen"].(float64)
	if !ok {
		pagelen = 0
	}
	size, ok := prCommentsResponseMap["size"].(float64)
	if !ok {
		size = 0
	}

	pullRequestsCommentsList := PullRequestsCommentsList{
		Page:    int(page),
		Pagelen: int(pagelen),
		Size:    int(size),
		Items:   prCommentsList,
	}
	return &pullRequestsCommentsList, nil
}

func decodePullRequestsComments(pullRequestResp interface{}) (*PullRequestsComments, error) {
	prCommentMap := pullRequestResp.(map[string]interface{})

	if prCommentMap["type"] == "error" {
		return nil, DecodeError(prCommentMap)
	}

	var pullRequestsComments = new(PullRequestsComments)
	decoder, err := mapstructure.NewDecoder(&mapstructure.DecoderConfig{
		Metadata:   nil,
		Result:     pullRequestsComments,
		DecodeHook: stringToTimeHookFunc,
	})
	if err != nil {
		return nil, err
	}
	err = decoder.Decode(prCommentMap)
	if err != nil {
		return nil, err
	}

	return pullRequestsComments, nil
}
