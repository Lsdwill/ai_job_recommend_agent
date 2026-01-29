package config

import (
	"fmt"
	"os"
	"time"

	"gopkg.in/yaml.v3"
)

// Config 应用配置
type Config struct {
	Server      ServerConfig      `yaml:"server"`
	City        CityConfig        `yaml:"city"`
	LLM         LLMConfig         `yaml:"llm"`
	Amap        AmapConfig        `yaml:"amap"`
	JobAPI      JobAPIConfig      `yaml:"job_api"`
	OCR         OCRConfig         `yaml:"ocr"`
	Policy      PolicyConfig      `yaml:"policy"`
	Embedding   EmbeddingConfig   `yaml:"embedding"`
	Milvus      MilvusConfig      `yaml:"milvus"`
	Logging     LoggingConfig     `yaml:"logging"`
	Performance PerformanceConfig `yaml:"performance"`
}

// CityConfig 城市配置
type CityConfig struct {
	Name          string            `yaml:"name"`          // 城市名称，如：青岛
	SystemName    string            `yaml:"system_name"`   // 系统名称，如：青岛岗位匹配系统
	AreaCodes     map[string]string `yaml:"area_codes"`    // 区域代码映射，如：市南区:0, 市北区:1
	Landmarks     []string          `yaml:"landmarks"`     // 地标示例，如：五四广场、青岛啤酒博物馆
	Abbreviations map[string]string `yaml:"abbreviations"` // 简称映射，如：青啤:青岛啤酒
}

// ServerConfig 服务器配置
type ServerConfig struct {
	Port         int           `yaml:"port"`
	Host         string        `yaml:"host"`
	ReadTimeout  time.Duration `yaml:"read_timeout"`
	WriteTimeout time.Duration `yaml:"write_timeout"`
}

// LLMConfig LLM配置
type LLMConfig struct {
	BaseURL    string        `yaml:"base_url"`
	APIKey     string        `yaml:"api_key"`
	Model      string        `yaml:"model"`
	Timeout    time.Duration `yaml:"timeout"`
	MaxRetries int           `yaml:"max_retries"`
}

// AmapConfig 高德地图配置
type AmapConfig struct {
	APIKey  string        `yaml:"api_key"`
	BaseURL string        `yaml:"base_url"`
	Timeout time.Duration `yaml:"timeout"`
}

// JobAPIConfig 岗位API配置
type JobAPIConfig struct {
	BaseURL string        `yaml:"base_url"`
	Timeout time.Duration `yaml:"timeout"`
}

// OCRConfig OCR服务配置
type OCRConfig struct {
	BaseURL string        `yaml:"base_url"`
	Timeout time.Duration `yaml:"timeout"`
}

// PolicyConfig 政策API配置
type PolicyConfig struct {
	BaseURL string        `yaml:"base_url"`
	Timeout time.Duration `yaml:"timeout"`
}

// EmbeddingConfig Embedding配置
type EmbeddingConfig struct {
	BaseURL string        `yaml:"base_url"`
	Timeout time.Duration `yaml:"timeout"`
}

// MilvusConfig Milvus向量数据库配置
type MilvusConfig struct {
	Host           string        `yaml:"host"`
	Port           int           `yaml:"port"`
	CollectionName string        `yaml:"collection_name"`
	Dimension      int           `yaml:"dimension"`
	Timeout        time.Duration `yaml:"timeout"`
}

// LoggingConfig 日志配置
type LoggingConfig struct {
	Level  string `yaml:"level"`
	Format string `yaml:"format"`
}

// PerformanceConfig 性能配置
type PerformanceConfig struct {
	MaxGoroutines     int   `yaml:"max_goroutines"`
	GoroutinePoolSize int   `yaml:"goroutine_pool_size"`
	TaskQueueSize     int   `yaml:"task_queue_size"`
	EnablePprof       *bool `yaml:"enable_pprof"`
	EnableMetrics     *bool `yaml:"enable_metrics"`
	GCPercent         int   `yaml:"gc_percent"`
}

var globalConfig *Config

// Load 从文件加载配置
func Load(configPath string) (*Config, error) {
	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("读取配置文件失败: %w", err)
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("解析配置文件失败: %w", err)
	}

	// 环境变量覆盖（用于生产环境，避免密钥泄露）
	if v := os.Getenv("LLM_API_KEY"); v != "" {
		cfg.LLM.APIKey = v
	}
	if v := os.Getenv("LLM_BASE_URL"); v != "" {
		cfg.LLM.BaseURL = v
	}
	if v := os.Getenv("AMAP_API_KEY"); v != "" {
		cfg.Amap.APIKey = v
	}
	if v := os.Getenv("OCR_BASE_URL"); v != "" {
		cfg.OCR.BaseURL = v
	}
	if v := os.Getenv("EMBEDDING_BASE_URL"); v != "" {
		cfg.Embedding.BaseURL = v
	}
	if v := os.Getenv("MILVUS_HOST"); v != "" {
		cfg.Milvus.Host = v
	}
	if v := os.Getenv("MILVUS_PORT"); v != "" {
		fmt.Sscanf(v, "%d", &cfg.Milvus.Port)
	}
	if v := os.Getenv("SERVER_PORT"); v != "" {
		fmt.Sscanf(v, "%d", &cfg.Server.Port)
	}

	// 设置默认值
	if cfg.Server.Port == 0 {
		cfg.Server.Port = 8080
	}
	if cfg.Server.Host == "" {
		cfg.Server.Host = "0.0.0.0"
	}
	if cfg.LLM.MaxRetries == 0 {
		cfg.LLM.MaxRetries = 3
	}
	if cfg.Server.ReadTimeout == 0 {
		cfg.Server.ReadTimeout = 30 * time.Second
	}
	if cfg.Server.WriteTimeout == 0 {
		cfg.Server.WriteTimeout = 300 * time.Second
	}

	// 城市配置默认值
	if cfg.City.Name == "" {
		cfg.City.Name = "青岛"
	}
	if cfg.City.SystemName == "" {
		cfg.City.SystemName = cfg.City.Name + "岗位匹配系统"
	}
	if len(cfg.City.AreaCodes) == 0 {
		cfg.City.AreaCodes = map[string]string{
			"市南区": "0", "市北区": "1", "李沧区": "2", "崂山区": "3", "黄岛区": "4",
			"城阳区": "5", "即墨区": "6", "胶州市": "7", "平度市": "8", "莱西市": "9",
		}
	}
	if len(cfg.City.Landmarks) == 0 {
		cfg.City.Landmarks = []string{"五四广场", "青岛啤酒博物馆"}
	}
	if len(cfg.City.Abbreviations) == 0 {
		cfg.City.Abbreviations = map[string]string{"青啤": "青岛啤酒"}
	}

	// 性能配置默认值
	if cfg.Performance.MaxGoroutines == 0 {
		cfg.Performance.MaxGoroutines = 10000
	}
	if cfg.Performance.GoroutinePoolSize == 0 {
		cfg.Performance.GoroutinePoolSize = 5000
	}
	if cfg.Performance.TaskQueueSize == 0 {
		cfg.Performance.TaskQueueSize = 10000
	}
	if cfg.Performance.GCPercent == 0 {
		cfg.Performance.GCPercent = 100
	}
	// pprof和metrics默认启用（允许在配置文件中显式关闭）
	if cfg.Performance.EnablePprof == nil {
		v := true
		cfg.Performance.EnablePprof = &v
	}
	if cfg.Performance.EnableMetrics == nil {
		v := true
		cfg.Performance.EnableMetrics = &v
	}

	globalConfig = &cfg
	return &cfg, nil
}

// Get 获取全局配置
func Get() *Config {
	return globalConfig
}

// GetAreaCodesDescription 获取区域代码描述字符串
func (c *CityConfig) GetAreaCodesDescription() string {
	if len(c.AreaCodes) == 0 {
		return ""
	}
	// 按代码排序输出
	result := ""
	for i := 0; i <= 9; i++ {
		for name, code := range c.AreaCodes {
			if code == fmt.Sprintf("%d", i) {
				if result != "" {
					result += ", "
				}
				result += fmt.Sprintf("%s(%s)", name, code)
				break
			}
		}
	}
	return result
}

// GetLandmarksExample 获取地标示例字符串
func (c *CityConfig) GetLandmarksExample() string {
	if len(c.Landmarks) == 0 {
		return ""
	}
	result := ""
	for i, landmark := range c.Landmarks {
		if i > 0 {
			result += "、"
		}
		result += landmark
	}
	return result
}

// GetAbbreviationsDescription 获取简称映射描述
func (c *CityConfig) GetAbbreviationsDescription() string {
	if len(c.Abbreviations) == 0 {
		return ""
	}
	result := ""
	for abbr, full := range c.Abbreviations {
		if result != "" {
			result += "、"
		}
		result += fmt.Sprintf("\"%s\"指\"%s\"", abbr, full)
	}
	return result
}
