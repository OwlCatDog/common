package media

// URL 单文件URL类型（双字段模式）
//
// 用于标记URL字段，AutoFill 会自动从对应的ID字段获取文件ID并填充URL
// 命名约定：XxxURL 字段会从 Xxx 字段获取文件ID
//
// 示例:
//
//	type Response struct {
//	    Cover    string    `json:"cover"`     // ID 保持不变
//	    CoverURL media.URL `json:"cover_url"` // 自动填充 URL
//	}
type URL string

// URLs 多文件URL类型（双字段模式）
//
// 用于标记URL列表字段，AutoFill 会自动从对应的IDs字段获取文件ID列表并填充URL列表
// 命名约定：XxxURL 字段会从 Xxx 字段获取文件ID列表
//
// 示例:
//
//	type Response struct {
//	    Gallery    []string   `json:"gallery"`     // IDs 保持不变
//	    GalleryURL media.URLs `json:"gallery_url"` // 自动填充 URLs
//	}
type URLs []string

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
//	    Description media.RichText `json:"description"`
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
