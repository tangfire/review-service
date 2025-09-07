package snowflake

import (
	"fmt"
	"github.com/bwmarrin/snowflake"
	"time"
)

// 全局雪花ID生成器
var node *snowflake.Node

// InitSnowflake 初始化雪花算法节点
// nodeNum 是节点编号 (0-1023)
func InitSnowflake(nodeNum int64) error {
	var err error
	// 设置起始时间（可选，建议设置为项目启动时间）
	epoch := time.Date(2025, 9, 6, 0, 0, 0, 0, time.UTC)
	snowflake.Epoch = epoch.UnixNano() / 1000000

	// 创建新节点
	node, err = snowflake.NewNode(nodeNum)
	if err != nil {
		return fmt.Errorf("failed to init snowflake: %v", err)
	}
	return nil
}

// GenerateID 生成唯一ID
func GenerateID() int64 {
	if node == nil {
		// 如果没有初始化，使用默认节点0（生产环境不建议这样做）
		_ = InitSnowflake(0)
	}
	return node.Generate().Int64()
}

func main() {
	// 初始化节点（生产环境应该从配置读取节点ID）
	err := InitSnowflake(1) // 使用节点1
	if err != nil {
		panic(err)
	}

	// 生成10个ID作为示例
	for i := 0; i < 10; i++ {
		id := GenerateID()
		fmt.Printf("Generated ID %d: %d\n", i+1, id)

		// 解析ID（可选）
		sf := snowflake.ParseInt64(id)
		fmt.Printf("  Parsed: Node=%d Step=%d Time=%d\n",
			sf.Node(), sf.Step(), sf.Time())
	}
}
