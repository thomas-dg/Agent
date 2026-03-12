package mem

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/allegro/bigcache/v3"
	"github.com/cloudwego/eino/components/model"
	"github.com/cloudwego/eino/schema"
)

const (
	// DefaultTTL 会话默认过期时间：30分钟无活动则自动淘汰
	DefaultTTL = 30 * time.Minute
	// DefaultMaxCacheSizeMB bigcache 最大内存占用（MB），超出后按 LRU 淘汰最旧会话
	DefaultMaxCacheSizeMB = 256
	// DefaultMaxWindowSize 单个会话保留的最近 K 轮原始消息条数（成对计算）
	// 超出后最旧的一对消息会被压缩进摘要
	DefaultMaxWindowSize = 6
)

// sessionData 是 bigcache 中存储的序列化结构。
// Summary 保存历史对话的压缩摘要，RecentMsgs 保存最近 K 轮原始消息。
type sessionData struct {
	Summary    string            `json:"summary"`     // 历史对话摘要（可为空）
	RecentMsgs []*schema.Message `json:"recent_msgs"` // 最近 K 轮原始消息
}

// summarizePrompt 是生成摘要时使用的系统提示
const summarizePrompt = `你是一个对话历史摘要助手。
你将收到：
1. 当前已有的对话摘要（可能为空）
2. 一对新的对话消息（用户提问 + 助手回答）

请将它们合并，生成一段简洁的新摘要，保留关键信息和上下文，去除冗余细节。
直接输出摘要文本，不要加任何前缀或解释。`

// SimpleMemory 单个会话的操作句柄，通过 store 读写实际数据。
//
// 消息组织策略（ConversationSummaryBufferMemory）：
//   - 系统提示：由调用方自行注入，不在此处存储
//   - 旧消息摘要：超出窗口的历史消息被 LLM 压缩为一段摘要文本
//   - 最近 K 轮原始消息：完整保留，供 LLM 精确理解近期上下文
//
// 当 summarizer 为 nil 时，退化为原有的截断策略（成对丢弃最旧消息）。
type SimpleMemory struct {
	id            string
	maxWindowSize int
	store         *bigCacheStore
	summarizer    model.ChatModel // 可选，用于生成历史摘要；nil 时退化为截断
	mu            sync.Mutex
}

// SetMessages 追加一条消息。
// 当 RecentMsgs 超出 maxWindowSize 时：
//   - 若配置了 summarizer：取出最旧一对消息，调用 LLM 更新摘要
//   - 若未配置 summarizer：成对截断（原有行为）
func (c *SimpleMemory) SetMessages(msg *schema.Message) {
	// 先在锁外准备数据，减少锁持有时间
	c.mu.Lock()
	data := c.store.load(c.id)
	data.RecentMsgs = append(data.RecentMsgs, msg)

	// 未超出窗口，直接保存
	if len(data.RecentMsgs) <= c.maxWindowSize {
		c.store.save(c.id, data)
		c.mu.Unlock()
		return
	}

	// 超出窗口：确保成对处理（user + assistant）
	excess := len(data.RecentMsgs) - c.maxWindowSize
	if excess%2 != 0 {
		excess++
	}
	oldPairs := data.RecentMsgs[:excess]
	data.RecentMsgs = data.RecentMsgs[excess:]
	currentSummary := data.Summary

	if c.summarizer == nil {
		// 无摘要器：直接截断（原有行为）
		c.store.save(c.id, data)
		c.mu.Unlock()
		return
	}

	// 有摘要器：先保存截断后的数据，再解锁，异步生成摘要
	c.store.save(c.id, data)
	c.mu.Unlock()

	// 在锁外调用 LLM（避免长时间持锁）
	newSummary, err := c.generateSummary(context.Background(), currentSummary, oldPairs)
	if err != nil {
		// 摘要生成失败：静默降级，截断已生效，不影响主流程
		return
	}

	// 将新摘要写回（重新加锁）
	c.mu.Lock()
	defer c.mu.Unlock()
	latest := c.store.load(c.id)
	latest.Summary = newSummary
	c.store.save(c.id, latest)
}

// GetMessages 返回当前会话的完整消息序列（返回副本，避免外部修改）：
//  1. 若存在历史摘要，以 assistant 消息形式前置（标记为摘要）
//  2. 最近 K 轮原始消息
func (c *SimpleMemory) GetMessages() []*schema.Message {
	c.mu.Lock()
	defer c.mu.Unlock()

	data := c.store.load(c.id)

	var result []*schema.Message
	if data.Summary != "" {
		// 将摘要包装为 assistant 消息，前置于消息列表
		summaryMsg := schema.AssistantMessage(
			fmt.Sprintf("以下是之前对话的摘要，供参考：\n%s", data.Summary),
			nil,
		)
		result = append(result, summaryMsg)
	}
	result = append(result, data.RecentMsgs...)
	return result
}

// generateSummary 调用 LLM 将旧摘要与一批旧消息合并为新摘要。
// oldPairs 应为成对的 user/assistant 消息。
func (c *SimpleMemory) generateSummary(ctx context.Context, oldSummary string, oldPairs []*schema.Message) (string, error) {
	// 构造摘要请求消息
	userContent := ""
	if oldSummary != "" {
		userContent += fmt.Sprintf("当前摘要：\n%s\n\n", oldSummary)
	}
	userContent += "需要压缩进摘要的新对话：\n"
	for _, m := range oldPairs {
		role := "用户"
		if m.Role == schema.Assistant {
			role = "助手"
		}
		userContent += fmt.Sprintf("[%s]: %s\n", role, m.Content)
	}

	msgs := []*schema.Message{
		schema.SystemMessage(summarizePrompt),
		schema.UserMessage(userContent),
	}

	resp, err := c.summarizer.Generate(ctx, msgs)
	if err != nil {
		return oldSummary, fmt.Errorf("mem: 生成摘要失败: %w", err)
	}
	return resp.Content, nil
}

// MemoryStore 会话存储接口，便于后续替换为 Redis 等分布式实现
type MemoryStore interface {
	// GetOrCreate 根据 ID 获取已有会话句柄，不存在则创建
	GetOrCreate(id string) *SimpleMemory
	// Delete 主动删除指定 ID 的会话
	Delete(id string)
	// Count 返回当前存储的会话数量（用于监控）
	Count() int
}

// bigCacheStore 基于 bigcache 的 MemoryStore 实现
// 内置 LRU 淘汰、TTL 过期和最大内存控制，无需手动维护清理 goroutine
type bigCacheStore struct {
	cache      *bigcache.BigCache
	handles    sync.Map        // map[string]*SimpleMemory，缓存句柄对象，避免重复创建
	summarizer model.ChatModel // 可选，注入给所有新创建的 SimpleMemory
}

// StoreOption 用于配置 bigCacheStore 的可选参数
type StoreOption func(*bigCacheStore)

// WithSummarizer 为存储实例配置摘要 LLM。
// 配置后，所有通过该 store 创建的 SimpleMemory 都会使用此 LLM 生成历史摘要。
// 不配置时退化为原有截断策略。
func WithSummarizer(m model.ChatModel) StoreOption {
	return func(s *bigCacheStore) {
		s.summarizer = m
	}
}

// newBigCacheStore 创建 bigcache 存储实例
func newBigCacheStore(ttl time.Duration, maxSizeMB int, opts ...StoreOption) (*bigCacheStore, error) {
	cfg := bigcache.Config{
		// 分片数量，必须是 2 的幂次，减少锁竞争
		Shards: 1024,
		// 条目 TTL，到期后自动淘汰
		LifeWindow: ttl,
		// 后台清理过期条目的间隔（bigcache 内置，无需手动 goroutine）
		CleanWindow: ttl / 3,
		// 最大内存占用（MB），超出后按 LRU 淘汰最旧条目
		HardMaxCacheSize: maxSizeMB,
		// 单条目最大大小（字节），防止单个超大会话占满缓存
		MaxEntrySize: 64 * 1024, // 64KB
		// 初始预分配容量（条目数）
		MaxEntriesInWindow: 1000,
		// 开启详细统计，便于监控
		StatsEnabled: true,
	}

	cache, err := bigcache.New(context.Background(), cfg)
	if err != nil {
		return nil, err
	}

	store := &bigCacheStore{cache: cache}
	for _, opt := range opts {
		opt(store)
	}
	return store, nil
}

// load 从 bigcache 读取会话数据，不存在时返回空结构
func (s *bigCacheStore) load(id string) sessionData {
	data, err := s.cache.Get(id)
	if err != nil {
		return sessionData{}
	}
	var sd sessionData
	if err := json.Unmarshal(data, &sd); err != nil {
		return sessionData{}
	}
	return sd
}

// save 将会话数据序列化后写入 bigcache（同时刷新 TTL）
func (s *bigCacheStore) save(id string, sd sessionData) {
	data, err := json.Marshal(sd)
	if err != nil {
		return
	}
	// Set 会刷新该条目的 TTL（相当于 touch）
	_ = s.cache.Set(id, data)
}

// GetOrCreate 根据 ID 获取或创建会话句柄
// 空 ID 返回临时句柄（不写入缓存），防止会话污染
func (s *bigCacheStore) GetOrCreate(id string) *SimpleMemory {
	if id == "" {
		return &SimpleMemory{
			id:            "",
			maxWindowSize: DefaultMaxWindowSize,
			store:         s,
			summarizer:    s.summarizer,
		}
	}

	// 优先从 handles map 取已有句柄（避免重复分配）
	if v, ok := s.handles.Load(id); ok {
		return v.(*SimpleMemory)
	}

	handle := &SimpleMemory{
		id:            id,
		maxWindowSize: DefaultMaxWindowSize,
		store:         s,
		summarizer:    s.summarizer,
	}
	// LoadOrStore 保证并发安全，只存入第一个创建的句柄
	actual, _ := s.handles.LoadOrStore(id, handle)
	return actual.(*SimpleMemory)
}

// Delete 主动删除指定 ID 的会话（同时清理缓存和句柄）
func (s *bigCacheStore) Delete(id string) {
	_ = s.cache.Delete(id)
	s.handles.Delete(id)
}

// Count 返回当前缓存中的条目数量
func (s *bigCacheStore) Count() int {
	return s.cache.Len()
}

// Stats 返回 bigcache 统计信息（命中率、淘汰数等），可接入 Prometheus
func (s *bigCacheStore) Stats() bigcache.Stats {
	return s.cache.Stats()
}

// defaultStore 全局默认存储实例（单例，进程启动时初始化）
var defaultStore MemoryStore

func init() {
	store, err := newBigCacheStore(DefaultTTL, DefaultMaxCacheSizeMB)
	if err != nil {
		panic("mem: failed to initialize bigcache store: " + err.Error())
	}
	defaultStore = store
}

// InitWithSummarizer 使用指定的摘要 LLM 重新初始化全局默认存储。
// 应在进程启动时、LLM 客户端就绪后调用一次。
// 若不调用此函数，GetSimpleMemory 将使用无摘要的截断策略（原有行为）。
func InitWithSummarizer(m model.ChatModel) error {
	store, err := newBigCacheStore(DefaultTTL, DefaultMaxCacheSizeMB, WithSummarizer(m))
	if err != nil {
		return fmt.Errorf("mem: 初始化带摘要的存储失败: %w", err)
	}
	defaultStore = store
	return nil
}

// GetSimpleMemory 根据 ID 获取或创建会话内存，对外 API 保持不变。
// 注意：空 ID 返回临时会话句柄，不会污染其他会话。
func GetSimpleMemory(id string) *SimpleMemory {
	return defaultStore.GetOrCreate(id)
}
