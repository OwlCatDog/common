package media

// FileID 单图类型
//
// 用于标记单个图片ID字段，AutoFill 会自动将其转换为URL
//
// 示例:
//
//	type Response struct {
//	    Cover image.FileID `json:"cover"` // file_id → url
//	}
type FileID string

// FileIDs 多图类型
//
// 用于标记多个图片ID字段，AutoFill 会自动将其转换为URL列表
//
// 示例:
//
//	type Response struct {
//	    Gallery image.FileIDs `json:"gallery"` // []file_id → []url
//	}
type FileIDs []string

// RichText 富文本类型
//
// 用于标记富文本字段，AutoFill 会自动解析其中所有 data-helf="file_id" 属性
// 并替换为 src="url"
//
// 支持任意标签：<img>, <video>, <audio> 等
//
// 示例:
//
//	type Response struct {
//	    Description image.RichText `json:"description"`
//	}
//
// 富文本内容示例:
//
//	<p>介绍</p><img data-helf="abc123"><video data-helf="def456"></video>
//
// 填充后:
//
//	<p>介绍</p><img src="https://cdn.example.com/abc123.jpg"><video src="https://cdn.example.com/def456.mp4"></video>
type RichText string
