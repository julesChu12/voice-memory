package service

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

// MockLLMService 模拟 LLM 服务
type MockLLMService struct {
	MockSendMessage func(req ChatRequest) (*ChatResponse, error)
}

func (m *MockLLMService) SendMessage(req ChatRequest) (*ChatResponse, error) {
	if m.MockSendMessage != nil {
		return m.MockSendMessage(req)
	}
	return nil, fmt.Errorf("MockSendMessage not implemented")
}

func (m *MockLLMService) SendMessageStream(req ChatRequest, callback func(StreamChunk)) error {
	// Dummy implementation for interface satisfaction
	return nil
}

// TestOrganizeV2 tests the v2.0 knowledge organization prompt and parsing logic
func TestOrganizeV2(t *testing.T) {
	// 1. Prepare the mock response
	mockResponseJSON := `{
		"summary": "使用 Go gin 框架重构高 QPS 用户服务",
		"key_points": [
			"gin 框架性能优于 Python FastAPI",
			"当前用户服务 QPS 5000，需要重构"
		],
		"entities": {
			"people": [],
			"products": ["gin", "GORM", "FastAPI"],
			"companies": [],
			"locations": [],
			"concepts": ["QPS", "API服务", "重构"]
		},
		"category": "technology",
		"tags": ["Go", "gin", "性能优化"],
		"relations": [
			{
				"type": "contrasts_with",
				"target": "Python FastAPI",
				"context": "gin 性能优于 FastAPI"
			}
		],
		"observations": [
			{
				"category": "decision",
				"content": "使用 Go gin 框架重构用户服务",
				"context": "QPS 5000 性能瓶颈"
			}
		],
		"sentiment": "positive",
		"importance": "high",
		"action_items": [
			"搭建 gin + GORM 项目框架",
			"进行性能压测对比"
		]
	}`

	mockChatResponse := &ChatResponse{
		Type: "message",
		Content: []Content{
			{
				Type: "text",
				Text: mockResponseJSON,
			},
		},
	}

	// 2. Setup the mock service
	mockLLM := &MockLLMService{
		MockSendMessage: func(req ChatRequest) (*ChatResponse, error) {
			return mockChatResponse, nil
		},
	}

	// 3. Create the organizer with the mock service
	organizer := NewKnowledgeOrganizer(mockLLM)

	// 4. Call the Organize function
	result, err := organizer.Organize("some test content")

	// 5. Assert the results
	assert.NoError(t, err, "Organize function should not return an error")
	assert.NotNil(t, result, "Result should not be nil")

	assert.Equal(t, "使用 Go gin 框架重构高 QPS 用户服务", result.Summary)
	assert.Len(t, result.KeyPoints, 2)
	assert.Equal(t, "gin 框架性能优于 Python FastAPI", result.KeyPoints[0])

	// Assert Entities
	assert.Contains(t, result.Entities.Products, "gin")
	assert.Contains(t, result.Entities.Concepts, "QPS")
	assert.Empty(t, result.Entities.People)

	// Assert Category and Tags
	assert.Equal(t, "technology", result.Category)
	assert.Len(t, result.Tags, 3)
	assert.Contains(t, result.Tags, "Go")

	// Assert new V2 fields
	assert.Len(t, result.Relations, 1)
	assert.Equal(t, "contrasts_with", result.Relations[0].Type)
	assert.Equal(t, "Python FastAPI", result.Relations[0].Target)

	assert.Len(t, result.Observations, 1)
	assert.Equal(t, "decision", result.Observations[0].Category)
	assert.Equal(t, "使用 Go gin 框架重构用户服务", result.Observations[0].Content)

	assert.Len(t, result.ActionItems, 2)
	assert.Equal(t, "搭建 gin + GORM 项目框架", result.ActionItems[0])

	// Assert other fields
	assert.Equal(t, "positive", result.Sentiment)
	assert.Equal(t, "high", result.Importance)
}

// TestOrganizeJSONErrorHandling tests the fallback mechanism when JSON parsing fails
func TestOrganizeJSONErrorHandling(t *testing.T) {
	// 1. Prepare a malformed JSON response
	mockResponseJSON := `{"summary": "this is not valid json`

	mockChatResponse := &ChatResponse{
		Content: []Content{{Type: "text", Text: mockResponseJSON}},
	}

	// 2. Setup the mock service
	mockLLM := &MockLLMService{
		MockSendMessage: func(req ChatRequest) (*ChatResponse, error) {
			return mockChatResponse, nil
		},
	}

	// 3. Create the organizer
	organizer := NewKnowledgeOrganizer(mockLLM)

	// 4. Call the function
	testContent := "this is the original content"
	result, err := organizer.Organize(testContent)

	// 5. Assert the fallback behavior
	assert.NoError(t, err, "Error should be nil on JSON parsing failure as it has a fallback")
	assert.NotNil(t, result, "Result should not be nil")
	assert.Equal(t, testContent[:min(len(testContent), 30)], result.Summary, "Summary should be a snippet of original content")
	assert.Equal(t, "想法", result.Category, "Category should be the default")
	assert.Empty(t, result.KeyPoints)
	assert.Empty(t, result.Relations)
	assert.Empty(t, result.Observations)
}

// TestGenerateTitleFromSession_Mock tests title generation with a mock
func TestGenerateTitleFromSession_Mock(t *testing.T) {
	// 1. Mock response
	mockChatResponse := &ChatResponse{
		Type:    "message",
		Content: []Content{{Type: "text", Text: "Go框架并发编程"}},
	}

	// 2. Mock service
	mockLLM := &MockLLMService{
		MockSendMessage: func(req ChatRequest) (*ChatResponse, error) {
			// You could add assertions here about the request prompt if needed
			return mockChatResponse, nil
		},
	}

	// 3. Organizer
	organizer := NewKnowledgeOrganizer(mockLLM)

	// 4. Test session
	session := &Session{
		ID: "test_session_001",
		Messages: []Message{
			{Role: "user", Content: "我们来聊聊Go的并发"},
			{Role: "assistant", Content: "好呀，goroutine和channel是核心..."},
		},

	}

	// 5. Generate title
	title, err := organizer.GenerateTitleFromSession(session)

	// 6. Assert
	assert.NoError(t, err)
	assert.Equal(t, "Go框架并发编程", title)
}

// Just a dummy test to make sure file is not empty if all tests are skipped
func TestPlaceholder(t *testing.T) {
	assert.True(t, true)
}