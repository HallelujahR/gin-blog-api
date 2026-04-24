package service

import (
	"api/internal/modules/content/models"
	"sort"
	"strings"
	"unicode"
)

const (
	maxSuggestedCategories = 6
	maxSuggestedTags       = 10
)

type ContentTaxonomySuggestion struct {
	ID              uint64   `json:"id"`
	Name            string   `json:"name"`
	Slug            string   `json:"slug"`
	Score           int      `json:"score"`
	MatchedKeywords []string `json:"matched_keywords,omitempty"`
	Description     string   `json:"description,omitempty"`
}

type ContentTaxonomySuggestions struct {
	Categories []ContentTaxonomySuggestion `json:"categories"`
	Tags       []ContentTaxonomySuggestion `json:"tags"`
}

type taxonomyKeyword struct {
	Value  string
	Weight int
}

func SuggestTaxonomy(title, excerpt, content string) (*ContentTaxonomySuggestions, error) {
	categories, err := ListCategories()
	if err != nil {
		return nil, err
	}

	tags, err := ListTags()
	if err != nil {
		return nil, err
	}

	text := normalizeMatchText(strings.Join([]string{title, excerpt, content}, "\n"))
	return &ContentTaxonomySuggestions{
		Categories: scoreCategories(text, categories, maxSuggestedCategories),
		Tags:       scoreTags(text, tags, maxSuggestedTags),
	}, nil
}

func scoreCategories(text string, categories []models.Category, limit int) []ContentTaxonomySuggestion {
	scored := make([]ContentTaxonomySuggestion, 0, len(categories))
	for _, category := range categories {
		score, matched := taxonomyScore(text, buildCategoryKeywords(category))
		if score <= 0 {
			continue
		}

		scored = append(scored, ContentTaxonomySuggestion{
			ID:              category.ID,
			Name:            category.Name,
			Slug:            category.Slug,
			Score:           score,
			MatchedKeywords: matched,
			Description:     category.Description,
		})
	}

	return sortAndTrimSuggestions(scored, limit)
}

func scoreTags(text string, tags []models.Tag, limit int) []ContentTaxonomySuggestion {
	scored := make([]ContentTaxonomySuggestion, 0, len(tags))
	for _, tag := range tags {
		score, matched := taxonomyScore(text, buildTagKeywords(tag))
		if score <= 0 {
			continue
		}

		scored = append(scored, ContentTaxonomySuggestion{
			ID:              tag.ID,
			Name:            tag.Name,
			Slug:            tag.Slug,
			Score:           score,
			MatchedKeywords: matched,
			Description:     tag.Description,
		})
	}

	return sortAndTrimSuggestions(scored, limit)
}

func sortAndTrimSuggestions(items []ContentTaxonomySuggestion, limit int) []ContentTaxonomySuggestion {
	sort.Slice(items, func(i, j int) bool {
		if items[i].Score == items[j].Score {
			return items[i].ID < items[j].ID
		}
		return items[i].Score > items[j].Score
	})

	if limit > 0 && len(items) > limit {
		items = items[:limit]
	}

	return items
}

func taxonomyScore(text string, keywords []taxonomyKeyword) (int, []string) {
	score := 0
	matched := make([]string, 0, 4)
	seen := make(map[string]struct{})

	for _, keyword := range keywords {
		if keyword.Value == "" || !strings.Contains(text, keyword.Value) {
			continue
		}

		score += keyword.Weight
		if _, ok := seen[keyword.Value]; !ok {
			seen[keyword.Value] = struct{}{}
			matched = append(matched, keyword.Value)
		}
	}

	return score, matched
}

func buildCategoryKeywords(category models.Category) []taxonomyKeyword {
	keywords := make([]taxonomyKeyword, 0, 12)
	seen := make(map[string]struct{})

	appendWeightedKeywords(&keywords, seen, []string{category.Name}, 8)
	appendWeightedKeywords(&keywords, seen, []string{category.Slug}, 7)
	appendWeightedKeywords(&keywords, seen, splitKeywords(category.Description), 3)
	appendWeightedKeywords(&keywords, seen, categorySynonyms(category), 5)

	return keywords
}

func buildTagKeywords(tag models.Tag) []taxonomyKeyword {
	keywords := make([]taxonomyKeyword, 0, 12)
	seen := make(map[string]struct{})

	appendWeightedKeywords(&keywords, seen, []string{tag.Name}, 8)
	appendWeightedKeywords(&keywords, seen, []string{tag.Slug}, 7)
	appendWeightedKeywords(&keywords, seen, splitKeywords(tag.Description), 3)
	appendWeightedKeywords(&keywords, seen, tagSynonyms(tag), 5)

	return keywords
}

func appendWeightedKeywords(target *[]taxonomyKeyword, seen map[string]struct{}, source []string, weight int) {
	for _, item := range source {
		normalized := normalizeMatchText(item)
		if !shouldKeepKeyword(normalized) {
			continue
		}
		if _, ok := seen[normalized]; ok {
			continue
		}
		seen[normalized] = struct{}{}
		*target = append(*target, taxonomyKeyword{
			Value:  normalized,
			Weight: weight,
		})
	}
}

func tagSynonyms(tag models.Tag) []string {
	lookup := normalizeMatchText(tag.Name)
	return taxonomySynonymMap[lookup]
}

func categorySynonyms(category models.Category) []string {
	lookup := normalizeMatchText(category.Name)
	if synonyms, ok := taxonomySynonymMap[lookup]; ok {
		return synonyms
	}

	slugLookup := normalizeMatchText(category.Slug)
	return taxonomySynonymMap[slugLookup]
}

func splitKeywords(text string) []string {
	return strings.FieldsFunc(text, func(r rune) bool {
		return !(unicode.IsLetter(r) || unicode.IsNumber(r) || unicode.Is(unicode.Han, r))
	})
}

func shouldKeepKeyword(keyword string) bool {
	if keyword == "" {
		return false
	}

	runeCount := len([]rune(keyword))
	if containsHan(keyword) {
		return runeCount >= 2
	}

	return runeCount >= 2
}

func containsHan(text string) bool {
	for _, r := range text {
		if unicode.Is(unicode.Han, r) {
			return true
		}
	}
	return false
}

func normalizeMatchText(text string) string {
	text = strings.ToLower(strings.TrimSpace(text))
	var builder strings.Builder
	builder.Grow(len(text))

	for _, r := range text {
		switch {
		case unicode.IsLetter(r), unicode.IsNumber(r), unicode.Is(unicode.Han, r):
			builder.WriteRune(r)
		case r == '+' || r == '#' || r == '/':
			builder.WriteRune(r)
		default:
			builder.WriteRune(' ')
		}
	}

	return strings.Join(strings.Fields(builder.String()), " ")
}

var taxonomySynonymMap = map[string][]string{
	"golang":         {"go", "go语言", "golang", "go lang"},
	"mysql":          {"mysql", "my sql", "innodb"},
	"redis":          {"redis", "缓存", "cache"},
	"docker":         {"docker", "容器", "容器化"},
	"k8s":            {"k8s", "kubernetes", "k8s集群"},
	"grpc":           {"grpc", "rpc", "protobuf", "proto"},
	"rabbitmq":       {"rabbitmq", "amqp", "消息队列", "mq"},
	"消息队列":           {"消息队列", "mq", "异步队列", "消息中间件", "rabbitmq", "nats"},
	"nats":           {"nats", "消息总线"},
	"prometheus":     {"prometheus", "metrics", "指标监控"},
	"nginx":          {"nginx", "反向代理", "网关"},
	"load balancing": {"负载均衡", "load balancing", "lb"},
	"ci/cd":          {"ci/cd", "cicd", "持续集成", "持续交付", "持续部署"},
	"jenkins":        {"jenkins", "pipeline"},
	"github actions": {"github actions", "githubactions", "actions workflow"},
	"sql":            {"sql", "查询语句"},
	"索引":             {"索引", "index", "索引优化"},
	"优化":             {"优化", "调优", "优化方案"},
	"事务":             {"事务", "transaction", "acid"},
	"连接池":            {"连接池", "pool", "db pool"},
	"go routines":    {"goroutine", "goroutines", "go routine", "go routines"},
	"channels":       {"channel", "channels", "通道"},
	"profiling":      {"profiling", "性能分析", "profile"},
	"pprof":          {"pprof", "性能剖析"},
	"benchmark":      {"benchmark", "基准测试", "压测"},
	"trace":          {"trace", "链路追踪", "追踪"},
	"jaeger":         {"jaeger", "tracing"},
	"opentelemetry":  {"opentelemetry", "otel", "open telemetry"},
	"缓存穿透":           {"缓存穿透", "cache penetration"},
	"缓存雪崩":           {"缓存雪崩", "cache avalanche"},
	"读写分离":           {"读写分离", "read write split"},
	"主从复制":           {"主从复制", "replication", "master slave"},
	"分库分表":           {"分库分表", "sharding", "database sharding"},
	"schema设计":       {"schema设计", "schema 设计", "表结构设计"},
	"架构设计":           {"架构设计", "architecture", "系统设计"},
	"微服务":            {"微服务", "microservice", "microservices"},
	"monolith":       {"monolith", "单体", "单体架构"},
	"api gateway":    {"api gateway", "gateway", "网关"},
	"jwt":            {"jwt", "token"},
	"oauth2":         {"oauth2", "oauth", "授权"},
	"安全":             {"安全", "security", "鉴权"},
	"日志":             {"日志", "logging", "log"},
	"监控":             {"监控", "monitoring", "可观测性"},
	"服务网格":           {"服务网格", "service mesh"},
	"istio":          {"istio", "service mesh"},
	"elastic":        {"elastic", "elasticsearch", "es"},
	"测试":             {"测试", "test", "testing"},
	"摄影":             {"摄影", "photo", "拍摄"},
	"城市摄影":           {"城市摄影", "street photography", "city photography"},
	"自然":             {"自然", "nature"},
	"随手拍":            {"随手拍", "snapshot", "daily shot"},
	"南京":             {"南京", "nanjing"},
	"后端开发":           {"后端", "backend", "服务端", "后端开发"},
	"go 语言":          {"go", "go语言", "golang"},
	"gin 框架":         {"gin", "gin框架"},
	"微服务架构":          {"微服务", "microservice", "服务治理"},
	"数据库设计":          {"数据库设计", "schema设计", "数据建模"},
	"mysql 优化":       {"mysql优化", "sql优化", "索引优化", "慢查询"},
	"分布式系统":          {"分布式", "distributed system", "分布式系统"},
	"缓存策略":           {"缓存", "cache", "缓存策略"},
	"性能调优":           {"性能调优", "性能优化", "调优", "profiling"},
	"高并发处理":          {"高并发", "并发", "吞吐", "goroutine"},
	"rpc 与 grpc":     {"rpc", "grpc", "protobuf"},
	"容器化与 docker":    {"docker", "容器化", "镜像", "container"},
	"kubernetes 实践":  {"k8s", "kubernetes", "容器编排"},
	"ci cd":          {"ci/cd", "cicd", "持续集成", "持续部署"},
	"日志与监控":          {"日志", "监控", "observability", "prometheus"},
	"可靠性工程":          {"可靠性", "高可用", "容灾", "稳定性"},
	"网络与安全":          {"网络", "安全", "鉴权", "tls"},
	"系统架构设计":         {"架构设计", "系统设计", "分层设计"},
	"工程化与部署":         {"部署", "工程化", "发布", "交付"},
	"代码质量与重构":        {"代码质量", "重构", "clean code"},
	"测试与调试":          {"测试", "调试", "benchmark", "pprof"},
	"api 设计":         {"api设计", "restful", "grpc", "接口设计"},
	"认证与授权":          {"认证", "授权", "jwt", "oauth2"},
	"数据建模":           {"数据建模", "schema", "领域模型"},
	"索引与查询优化":        {"索引", "查询优化", "sql优化", "explain"},
	"自动化运维":          {"自动化运维", "运维自动化", "脚本化"},
	"负载均衡":           {"负载均衡", "load balancing", "nginx"},
	"故障排查":           {"故障排查", "排障", "troubleshooting"},
	"架构演进":           {"架构演进", "演进", "升级改造"},
	"运维与 sre":        {"sre", "运维", "可靠性工程"},
	"系统观测":           {"观测", "监控", "trace", "metrics", "logging"},
	"缓存与一致性":         {"缓存一致性", "缓存", "一致性"},
	"连接池与事务":         {"连接池", "事务", "database pool"},
	"性能基准与压测":        {"benchmark", "压测", "基准测试"},
	"日志分析":           {"日志分析", "logging", "elastic"},
	"平台化建设":          {"平台化", "平台建设", "基础平台"},
	"服务拆分":           {"服务拆分", "拆分", "微服务化"},
	"demo":           {"demo", "示例", "样例"},
	"生活":             {"生活", "日常", "lifestyle"},
	"citywalk":       {"citywalk", "city walk", "城市漫步"},
}
