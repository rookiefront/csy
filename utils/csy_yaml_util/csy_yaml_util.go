package csy_yaml_util

import (
	"os"
	"reflect"
	"strconv"
	"strings"

	"gopkg.in/yaml.v3"
)

func LoadByPath[T any](path string) (config T, err error) {
	file, err := os.ReadFile(path)
	if err == nil {
		return
	}
	err = yaml.Unmarshal(file, &config)
	return
}

// WriteFile 根据结构体实例生成带注释的 YAML 文件
func WriteFile(cfg interface{}, outputPath string) error {
	root := &yaml.Node{Kind: yaml.DocumentNode}
	content, err := buildNode(reflect.ValueOf(cfg))
	if err != nil {
		return err
	}
	root.Content = []*yaml.Node{content}
	out, err := yaml.Marshal(root)
	if err != nil {
		return err
	}
	return os.WriteFile(outputPath, out, 0644)
}

// Marshal 将结构体编码为 YAML 字节切片
// withComment: 是否包含结构体 tag "comment" 中的注释
func Marshal(v interface{}, withComment bool) ([]byte, error) {
	if !withComment {
		// 无注释，直接使用标准库
		return yaml.Marshal(v)
	}
	// 有注释，手动构建节点树
	root := &yaml.Node{Kind: yaml.DocumentNode}
	content, err := buildNode(reflect.ValueOf(v))
	if err != nil {
		return nil, err
	}
	root.Content = []*yaml.Node{content}
	return yaml.Marshal(root)
}

// Unmarshal 从 YAML 字节切片解码到结构体
func Unmarshal(data []byte, v interface{}) error {
	return yaml.Unmarshal(data, v)
}

// buildNode 递归构建 yaml.Node
func buildNode(v reflect.Value) (*yaml.Node, error) {
	// 处理指针
	if v.Kind() == reflect.Ptr {
		if v.IsNil() {
			return &yaml.Node{Kind: yaml.ScalarNode, Tag: "!!null", Value: "null"}, nil
		}
		v = v.Elem()
	}

	switch v.Kind() {
	case reflect.Struct:
		return buildStruct(v)
	case reflect.Slice, reflect.Array:
		return buildSlice(v)
	case reflect.Map:
		return buildMap(v)
	default:
		return buildScalar(v)
	}
}

// buildScalar 处理基本类型（string, int, bool, float 等）
func buildScalar(v reflect.Value) (*yaml.Node, error) {
	node := &yaml.Node{Kind: yaml.ScalarNode}
	switch v.Kind() {
	case reflect.String:
		node.Tag = "!!str"
		node.Value = v.String()
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		node.Tag = "!!int"
		node.Value = strconv.FormatInt(v.Int(), 10)
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		node.Tag = "!!int"
		node.Value = strconv.FormatUint(v.Uint(), 10)
	case reflect.Float32, reflect.Float64:
		node.Tag = "!!float"
		node.Value = strconv.FormatFloat(v.Float(), 'g', -1, 64)
	case reflect.Bool:
		node.Tag = "!!bool"
		node.Value = strconv.FormatBool(v.Bool())
	default:
		// 其他类型（如 interface{}）fallback 到标准 marshal
		data, err := yaml.Marshal(v.Interface())
		if err != nil {
			return nil, err
		}
		var tmp yaml.Node
		if err := yaml.Unmarshal(data, &tmp); err != nil {
			return nil, err
		}
		// 如果 tmp 是 DocumentNode，取其 Content 的第一个节点
		if tmp.Kind == yaml.DocumentNode && len(tmp.Content) > 0 {
			node = tmp.Content[0]
		} else {
			node = &tmp
		}
	}
	return node, nil
}

// buildStruct 处理结构体
func buildStruct(v reflect.Value) (*yaml.Node, error) {
	t := v.Type()
	node := &yaml.Node{Kind: yaml.MappingNode}

	for i := 0; i < v.NumField(); i++ {
		field := t.Field(i)
		if field.PkgPath != "" { // 忽略未导出字段
			continue
		}

		// 解析 yaml 标签
		yamlTag := field.Tag.Get("yaml")
		name := strings.ToLower(field.Name) // 默认小写
		if yamlTag != "" {
			parts := strings.Split(yamlTag, ",")
			if parts[0] != "" && parts[0] != "-" {
				name = parts[0]
			}
		}
		if name == "-" {
			continue
		}

		// 获取 comment 标签
		comment := strings.TrimSpace(field.Tag.Get("comment"))

		// 构建键节点（必须显式设置 Tag）
		keyNode := &yaml.Node{Kind: yaml.ScalarNode, Tag: "!!str", Value: name}
		if comment != "" {
			keyNode.HeadComment = comment
		}

		// 构建值节点
		valNode, err := buildNode(v.Field(i))
		if err != nil {
			return nil, err
		}

		node.Content = append(node.Content, keyNode, valNode)
	}
	return node, nil
}

// buildSlice 处理切片/数组
func buildSlice(v reflect.Value) (*yaml.Node, error) {
	node := &yaml.Node{Kind: yaml.SequenceNode}
	for i := 0; i < v.Len(); i++ {
		item, err := buildNode(v.Index(i))
		if err != nil {
			return nil, err
		}
		node.Content = append(node.Content, item)
	}
	return node, nil
}

// buildMap 处理 map（如果需要）
func buildMap(v reflect.Value) (*yaml.Node, error) {
	node := &yaml.Node{Kind: yaml.MappingNode}
	iter := v.MapRange()
	for iter.Next() {
		k := iter.Key()
		val := iter.Value()

		keyNode, err := buildNode(k)
		if err != nil {
			return nil, err
		}
		// 确保键是字符串节点
		if keyNode.Kind != yaml.ScalarNode || keyNode.Tag != "!!str" {
			// 若键不是字符串，转为字符串
			keyNode = &yaml.Node{Kind: yaml.ScalarNode, Tag: "!!str", Value: k.String()}
		}

		valNode, err := buildNode(val)
		if err != nil {
			return nil, err
		}
		node.Content = append(node.Content, keyNode, valNode)
	}
	return node, nil
}
