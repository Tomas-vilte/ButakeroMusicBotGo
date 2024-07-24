package youtube_provider

import (
	"context"
	"fmt"
	"google.golang.org/api/option"
	"google.golang.org/api/youtube/v3"
)

type (
	// SearchListCallWrapper es una interfaz que envuelve youtube.SearchListCall
	SearchListCallWrapper interface {
		Q(q string) SearchListCallWrapper
		MaxResults(maxResults int64) SearchListCallWrapper
		Type(typ string) SearchListCallWrapper
		Do() (*youtube.SearchListResponse, error)
	}

	// VideosListCallWrapper es una interfaz que envuelve youtube.VideosListCall
	VideosListCallWrapper interface {
		Id(id string) VideosListCallWrapper
		Do() (*youtube.VideoListResponse, error)
	}

	RealVideosListCallWrapper struct {
		Call *youtube.VideosListCall
	}

	RealSearchListCallWrapper struct {
		Call *youtube.SearchListCall
	}

	RealYouTubeClient struct {
		Service *youtube.Service
	}
)

func NewRealYouTubeClient(apiKey string) (*RealYouTubeClient, error) {
	ctx := context.Background()
	service, err := youtube.NewService(ctx, option.WithAPIKey(apiKey))
	if err != nil {
		return nil, fmt.Errorf("error al crear el servicio de YouTube: %w", err)
	}
	return &RealYouTubeClient{service}, nil
}

func (c *RealYouTubeClient) VideosListCall(ctx context.Context, part []string) VideosListCallWrapper {
	return &RealVideosListCallWrapper{Call: c.Service.Videos.List(part)}
}

func (c *RealYouTubeClient) SearchListCall(ctx context.Context, part []string) SearchListCallWrapper {
	return &RealSearchListCallWrapper{Call: c.Service.Search.List(part)}
}

func (r *RealSearchListCallWrapper) Q(q string) SearchListCallWrapper {
	r.Call.Q(q)
	return r
}

func (r *RealSearchListCallWrapper) MaxResults(maxResults int64) SearchListCallWrapper {
	r.Call.MaxResults(maxResults)
	return r
}

func (r *RealSearchListCallWrapper) Type(typ string) SearchListCallWrapper {
	r.Call.Type(typ)
	return r
}

func (r *RealSearchListCallWrapper) Do() (*youtube.SearchListResponse, error) {
	return r.Call.Do()
}

func (r *RealVideosListCallWrapper) Id(id string) VideosListCallWrapper {
	r.Call.Id(id)
	return r
}

func (r *RealVideosListCallWrapper) Do() (*youtube.VideoListResponse, error) {
	return r.Call.Do()
}
