package media

import (
	"context"
	"testing"
)

// 模拟 Resolver
type autoFillMockResolver struct {
	data map[string]*ResourceInfo
}

func (m *autoFillMockResolver) Resolve(ctx context.Context, ids []string) (map[string]*ResourceInfo, error) {
	result := make(map[string]*ResourceInfo)
	for _, id := range ids {
		if info, ok := m.data[id]; ok {
			result[id] = info
		}
	}
	return result, nil
}

// ========== 源结构体（模拟 ent） ==========

type ProductLanguage struct {
	Name        string
	Cover       string   // 文件ID
	Gallery     []string // 文件ID列表
	Description string   // 富文本
}

type Product struct {
	ID        uint32
	Points    float64
	Status    int32
	Languages map[string]*ProductLanguage
}

// ========== 目标结构体（DTO）- 双字段模式 ==========

type ProductLangDTO struct {
	Name        string   `json:"name"`
	Cover       FileID   `json:"cover"`                  // ID 保持不变
	CoverURL    URL      `json:"cover_url" media:"Cover"` // URL 从 Cover 获取
	Gallery     FileIDs  `json:"gallery"`                 // IDs 保持不变
	GalleryURL  URLs     `json:"gallery_url" media:"Gallery"` // URLs 从 Gallery 获取
	Description RichText `json:"description"`             // 富文本
}

type ProductDTO struct {
	ID        uint32                     `json:"id"`
	Points    float64                    `json:"points"`
	Status    int32                      `json:"status"`
	Languages map[string]*ProductLangDTO `json:"languages"`
}

func TestAutoFill(t *testing.T) {
	// 模拟文件URL映射
	resolver := &autoFillMockResolver{
		data: map[string]*ResourceInfo{
			"cover_id":  {URL: "https://cdn.example.com/cover.jpg", Success: true},
			"gallery_1": {URL: "https://cdn.example.com/g1.jpg", Success: true},
			"gallery_2": {URL: "https://cdn.example.com/g2.jpg", Success: true},
			"rich_img":  {URL: "https://cdn.example.com/rich.jpg", Success: true},
			"cover_en":  {URL: "https://cdn.example.com/cover_en.jpg", Success: true},
			"video_id":  {URL: "https://cdn.example.com/video.mp4", Success: true},
		},
	}
	filler := NewFiller(resolver)

	// 源数据（模拟从数据库查询）
	products := []*Product{
		{
			ID:     1,
			Points: 99.9,
			Status: 1,
			Languages: map[string]*ProductLanguage{
				"zh": {
					Name:        "商品A",
					Cover:       "cover_id",
					Gallery:     []string{"gallery_1", "gallery_2"},
					Description: `<p>介绍</p><img data-helf="rich_img"><video data-helf="video_id"></video>`,
				},
				"en": {
					Name:        "Product A",
					Cover:       "cover_en",
					Gallery:     []string{"gallery_1"},
					Description: `<p>Description</p>`,
				},
			},
		},
	}

	// 执行 AutoFill
	var result []*ProductDTO
	err := AutoFill(context.Background(), filler, products, &result)
	if err != nil {
		t.Fatalf("AutoFill error: %v", err)
	}

	// 验证结果
	if len(result) != 1 {
		t.Fatalf("expected 1 result, got %d", len(result))
	}

	dto := result[0]

	// 验证基本字段
	if dto.ID != 1 {
		t.Errorf("ID: expected 1, got %d", dto.ID)
	}
	if dto.Points != 99.9 {
		t.Errorf("Points: expected 99.9, got %f", dto.Points)
	}
	if dto.Status != 1 {
		t.Errorf("Status: expected 1, got %d", dto.Status)
	}

	// 验证中文
	zh := dto.Languages["zh"]
	if zh == nil {
		t.Fatal("zh language is nil")
	}
	if zh.Name != "商品A" {
		t.Errorf("zh.Name: expected 商品A, got %s", zh.Name)
	}

	// 验证双字段模式 - ID保持不变
	if string(zh.Cover) != "cover_id" {
		t.Errorf("zh.Cover (ID): expected cover_id, got %s", zh.Cover)
	}
	// 验证双字段模式 - URL自动填充
	if string(zh.CoverURL) != "https://cdn.example.com/cover.jpg" {
		t.Errorf("zh.CoverURL: expected URL, got %s", zh.CoverURL)
	}

	// 验证多图 - IDs保持不变
	if len(zh.Gallery) != 2 {
		t.Errorf("zh.Gallery: expected 2 items, got %d", len(zh.Gallery))
	}
	if string(zh.Gallery[0]) != "gallery_1" {
		t.Errorf("zh.Gallery[0] (ID): expected gallery_1, got %s", zh.Gallery[0])
	}

	// 验证多图 - URLs自动填充
	if len(zh.GalleryURL) != 2 {
		t.Errorf("zh.GalleryURL: expected 2 items, got %d", len(zh.GalleryURL))
	}
	if zh.GalleryURL[0] != "https://cdn.example.com/g1.jpg" {
		t.Errorf("zh.GalleryURL[0]: expected URL, got %s", zh.GalleryURL[0])
	}
	if zh.GalleryURL[1] != "https://cdn.example.com/g2.jpg" {
		t.Errorf("zh.GalleryURL[1]: expected URL, got %s", zh.GalleryURL[1])
	}

	// 验证富文本替换
	expectedDesc := `<p>介绍</p><img src="https://cdn.example.com/rich.jpg"><video src="https://cdn.example.com/video.mp4"></video>`
	if string(zh.Description) != expectedDesc {
		t.Errorf("zh.Description:\nexpected: %s\ngot: %s", expectedDesc, zh.Description)
	}

	// 验证英文
	en := dto.Languages["en"]
	if en == nil {
		t.Fatal("en language is nil")
	}
	if string(en.Cover) != "cover_en" {
		t.Errorf("en.Cover (ID): expected cover_en, got %s", en.Cover)
	}
	if string(en.CoverURL) != "https://cdn.example.com/cover_en.jpg" {
		t.Errorf("en.CoverURL: expected URL, got %s", en.CoverURL)
	}

	t.Log("All tests passed!")
	t.Logf("zh.Cover (ID): %s", zh.Cover)
	t.Logf("zh.CoverURL: %s", zh.CoverURL)
	t.Logf("zh.Gallery (IDs): %v", zh.Gallery)
	t.Logf("zh.GalleryURL: %v", zh.GalleryURL)
	t.Logf("zh.Description: %s", zh.Description)
}

func TestAutoFillOne(t *testing.T) {
	resolver := &autoFillMockResolver{
		data: map[string]*ResourceInfo{
			"single_cover": {URL: "https://cdn.example.com/single.jpg", Success: true},
		},
	}
	filler := NewFiller(resolver)

	src := &Product{
		ID:     2,
		Points: 50.0,
		Languages: map[string]*ProductLanguage{
			"zh": {
				Name:  "单个商品",
				Cover: "single_cover",
			},
		},
	}

	var dst ProductDTO
	err := AutoFillOne(context.Background(), filler, src, &dst)
	if err != nil {
		t.Fatalf("AutoFillOne error: %v", err)
	}

	if dst.ID != 2 {
		t.Errorf("ID: expected 2, got %d", dst.ID)
	}
	if string(dst.Languages["zh"].Cover) != "single_cover" {
		t.Errorf("Cover (ID): expected single_cover, got %s", dst.Languages["zh"].Cover)
	}
	if string(dst.Languages["zh"].CoverURL) != "https://cdn.example.com/single.jpg" {
		t.Errorf("CoverURL: expected URL, got %s", dst.Languages["zh"].CoverURL)
	}

	t.Log("AutoFillOne test passed!")
}
