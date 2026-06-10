package snowflake

import (
	"fmt"

	"github.com/bwmarrin/snowflake"
)

var node *snowflake.Node

func init() {
	node, _ = snowflake.NewNode(1)
}

// Init 初始化雪花算法节点.
func Init(nodeID int64) error {
	var err error
	node, err = snowflake.NewNode(nodeID)
	if err != nil {
		return fmt.Errorf("failed to initialize snowflake node: %w", err)
	}
	return nil
}

// Generate 生成唯一 ID.
func GenerateID() int64 {
	if node == nil {
		panic("snowflake node is not initialized")
	}
	return node.Generate().Int64()
}

// GenerateString 生成唯一字符串 ID（Base62 编码）.
func GenerateUniqueString() string {
	id := GenerateID()
	return base62Encode(id)
}

// base62Encode 将数字转换为 Base62 编码.
func base62Encode(num int64) string {
	charset := "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz"
	base := int64(len(charset))
	var result []byte
	for num > 0 {
		result = append(result, charset[num%base])
		num /= base
	}
	// 反转结果
	for i, j := 0, len(result)-1; i < j; i, j = i+1, j-1 {
		result[i], result[j] = result[j], result[i]
	}
	return string(result)
}
