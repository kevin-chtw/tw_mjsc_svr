package ai

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/kevin-chtw/tw_common/gamebase/mahjong"
	"github.com/topfreegames/pitaya/v3/pkg/logger"
)

// HTTPAIClient HTTP 客户端，负责与 Python AI 服务通信
type HTTPAIClient struct {
	baseURL string
	client  *http.Client
}

var httpAIClient *HTTPAIClient

// InitHTTPAIClient 初始化 HTTP AI 服务客户端
func InitHTTPAIClient(addr string) error {
	httpAIClient = &HTTPAIClient{
		baseURL: fmt.Sprintf("http://%s", addr),
		client: &http.Client{
			Timeout: 5 * time.Second,
		},
	}

	// 测试连接
	resp, err := httpAIClient.client.Get(httpAIClient.baseURL + "/health")
	if err != nil {
		return fmt.Errorf("failed to connect to AI service: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("AI service not healthy: status %d", resp.StatusCode)
	}

	logger.Log.Infof("✅ Connected to Python AI service at %s", addr)
	return nil
}

// GetHTTPAIClient 获取全局 HTTP AI 客户端
func GetHTTPAIClient() *HTTPAIClient {
	return httpAIClient
}

// StepTransition 单步转移
type StepTransition struct {
	State     []float32 `json:"state"`
	Operate   int       `json:"operate"` // 操作类型
	Tile      int       `json:"tile"`    // 牌值
	Reward    float32   `json:"reward"`
	NextState []float32 `json:"next_state,omitempty"`
	Done      bool      `json:"done"`
}

// Episode 整局轨迹
type Episode struct {
	Steps        []StepTransition `json:"steps"`
	ShapedReward float32          `json:"shaped_reward"`
	IsHu         bool             `json:"is_hu"`
	HuMulti      int64            `json:"hu_multi"`
	IsLiuju      bool             `json:"is_liuju"`
}

// CandidateAction 候选动作
type CandidateAction struct {
	Operate int `json:"operate"` // 操作类型
	Tile    int `json:"tile"`    // 牌值
}

// GetDecisionRequest 决策请求（包含候选动作列表）
type GetDecisionRequest struct {
	Obs        []float32         `json:"obs"`        // 观察向量
	Candidates []CandidateAction `json:"candidates"` // 可行的候选动作
}

// GetDecisionResponse 决策响应
type GetDecisionResponse struct {
	Operate int `json:"operate"` // 操作类型
	Tile    int `json:"tile"`    // 牌值
}

// GetDecision 发送观察向量和候选动作给Python，让AI决策
func (c *HTTPAIClient) GetDecision(state *GameState, obs []float32, candidates []*Decision) (*Decision, error) {
	start := time.Now()

	// 转换candidates为可序列化格式
	candActions := make([]CandidateAction, 0, len(candidates))
	for _, cand := range candidates {
		tileIndex := 0
		// PASS操作的tile统一为0，其他操作需要转换为index
		if cand.Operate != int(mahjong.OperatePass) && cand.Tile != 0 {
			tileIndex = mahjong.ToIndex(cand.Tile)
		} else if cand.Operate == int(mahjong.OperatePass) && cand.Tile != 0 {
			// 检测到PASS候选的tile不是0，记录警告
			logger.Log.Warnf("PASS candidate has non-zero tile: %d, normalizing to 0", cand.Tile)
		}
		candActions = append(candActions, CandidateAction{
			Operate: cand.Operate,
			Tile:    tileIndex,
		})
	}

	reqData := GetDecisionRequest{
		Obs:        obs,
		Candidates: candActions,
	}

	jsonData, err := json.Marshal(reqData)
	if err != nil {
		return nil, err
	}

	marshalTime := time.Since(start)

	resp, err := c.client.Post(c.baseURL+"/get_decision", "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	httpTime := time.Since(start) - marshalTime

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("AI service returned status %d: %s", resp.StatusCode, string(body))
	}

	var respData GetDecisionResponse
	if err := json.NewDecoder(resp.Body).Decode(&respData); err != nil {
		return nil, err
	}

	totalTime := time.Since(start)

	if totalTime > 500*time.Millisecond {
		logger.Log.Warnf("GetDecision slow: total=%v, http=%v, marshal=%v", totalTime, httpTime, marshalTime)
	}

	d := &Decision{
		Operate: respData.Operate,
		Tile:    mahjong.FromIndex(respData.Tile),
	}
	if d.Operate == int(mahjong.OperatePass) {
		d.Tile = 0
	}
	return d, nil
}

// ReportEpisode 向 Python 服务上报一局轨迹（异步）
func (c *HTTPAIClient) ReportEpisode(episode *Episode) {
	// 异步发送，不阻塞游戏
	go func() {
		jsonData, err := json.Marshal(episode)
		if err != nil {
			logger.Log.Warnf("Failed to marshal episode: %v", err)
			return
		}

		resp, err := c.client.Post(c.baseURL+"/report_episode", "application/json", bytes.NewBuffer(jsonData))
		if err != nil {
			logger.Log.Warnf("Failed to report episode: %v", err)
			return
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			body, _ := io.ReadAll(resp.Body)
			logger.Log.Warnf("AI service returned status %d: %s", resp.StatusCode, string(body))
		}
	}()
}

// Close 关闭连接（HTTP client 不需要特殊关闭）
func (c *HTTPAIClient) Close() error {
	return nil
}
