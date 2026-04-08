package ai

import "context"

type DraftRequest struct {
	ScholarshipID string
	PromptContext string
}

type DraftResponse struct {
	Content string `json:"content"`
}

type Client interface {
	GenerateMotivationLetterDraft(ctx context.Context, input DraftRequest) (DraftResponse, error)
}

type NoopClient struct{}

func NewNoopClient() *NoopClient {
	return &NoopClient{}
}

func (c *NoopClient) GenerateMotivationLetterDraft(context.Context, DraftRequest) (DraftResponse, error) {
	return DraftResponse{
		Content: "motivation letter draft generation is not implemented yet",
	}, nil
}
