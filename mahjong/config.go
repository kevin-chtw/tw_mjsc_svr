package mahjong

import (
	"encoding/json"
	"strconv"
	"strings"
)

type ConfigIndex uint

type ConfigFDDescriptor struct {
	JsonInfo        map[int]string      // json_index到json键的映射
	IndexReflection map[int]ConfigIndex // json_index到ConfigIndex的映射
}

type FDConfigMap map[string]ConfigIndex

type Config struct {
	values []int
}

func NewConfig() *Config {
	return &Config{
		values: make([]int, 0),
	}
}

func (c *Config) InitDefault(defaultValues []int) {
	c.values = make([]int, len(defaultValues))
	copy(c.values, defaultValues)
}

func (c *Config) GetValue(index ConfigIndex) int {
	if int(index) < len(c.values) {
		return c.values[index]
	}
	return 0
}

func (c *Config) ToCommaString() string {
	strValues := make([]string, len(c.values))
	for i, v := range c.values {
		strValues[i] = strconv.Itoa(v)
	}
	return strings.Join(strValues, ",")
}

func (c *Config) LoadRuler(commaRuler string) {
	parts := strings.Split(commaRuler, ",")
	c.values = make([]int, len(parts))
	for i := range parts {
		// 这里需要添加字符串到int的转换逻辑
		c.values[i] = 0 // 临时值，实际应根据字符串解析
	}
}

func (c *Config) LoadJsonRule(jsonRuler string, formats ConfigFDDescriptor) {
	var jsonData interface{}
	if err := json.Unmarshal([]byte(jsonRuler), &jsonData); err != nil {
		return
	}
	c._InitJsonConfig(formats)
}

func (c *Config) LoadJsonRuleWithMap(jsonRuler string, configs FDConfigMap) {
	var jsonData interface{}
	if err := json.Unmarshal([]byte(jsonRuler), &jsonData); err != nil {
		return
	}
	// 实现基于configs的配置加载
}

func (c *Config) GetValues() []int {
	return c.values
}

func (c *Config) _InitJsonConfig(formats ConfigFDDescriptor) {
	// 实现基于formats的JSON配置初始化
}
