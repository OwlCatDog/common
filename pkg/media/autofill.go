package media

import (
	"context"
	"reflect"
	"regexp"
	"sync"
)

// dataHelfRegex 匹配任意标签的 data-helf="xxx" 属性中的文件ID
// 支持 <img>, <video>, <audio> 等任意标签
var dataHelfRegex = regexp.MustCompile(`data-helf=["']([^"']+)["']`)

// ==================== 类型缓存 ====================

// typeInfo 缓存的类型信息
type typeInfo struct {
	fields []fieldInfo
}

// fieldInfo 字段信息
type fieldInfo struct {
	srcIndex  int    // 源字段索引
	dstIndex  int    // 目标字段索引
	name      string // 字段名
	fieldType fieldType
	// 嵌套类型信息（slice/struct/map）
	elemInfo *typeInfo
	srcElem  reflect.Type
	dstElem  reflect.Type
	keyType  reflect.Type // map的key类型
}

// fieldType 字段类型
type fieldType int

const (
	fieldTypeBasic    fieldType = iota // 基本类型，直接复制
	fieldTypeFileID                   // FileID 类型
	fieldTypeFileIDs                  // FileIDs 类型
	fieldTypeRichText                  // RichText 类型
	fieldTypeSlice                     // 切片类型，需要递归
	fieldTypeStruct                    // 结构体类型，需要递归
	fieldTypeMap                       // Map类型，需要递归（如多语言 map[string]*Lang）
)

// typeCache 类型信息缓存
var typeCache sync.Map // map[typePair]*typeInfo

// typePair 类型对
type typePair struct {
	src reflect.Type
	dst reflect.Type
}

// ==================== AutoFill 入口 ====================

// AutoFill 自动映射并填充图片URL
//
// 将源切片自动映射到目标切片，并填充所有图片URL
//
// 支持的图片字段类型:
//   - FileID: 单图，file_id → url
//   - FileIDs: 多图，[]file_id → []url
//   - RichText: 富文本，<img src="file_id"> → <img src="url">
//
// 参数:
//   - ctx: 上下文
//   - filler: 填充器
//   - src: 源数据切片（如 []*ent.Product）
//   - dst: 目标切片指针（如 *[]*ProductResponse）
//
// 示例:
//
//	var responses []*ProductResponse
//	image.AutoFill(ctx, filler, products, &responses)
func AutoFill[S, D any](ctx context.Context, filler *Filler, src []S, dst *[]D) error {
	if len(src) == 0 || dst == nil {
		return nil
	}

	// 1. 创建目标切片
	result := make([]D, len(src))

	// 2. 获取类型信息
	srcType := reflect.TypeOf(src).Elem()
	dstType := reflect.TypeOf(result).Elem()
	info := getTypeInfo(srcType, dstType)

	// 3. 收集所有图片ID
	collector := &idCollector{ids: make(map[string]struct{})}

	// 4. 映射并收集ID
	for i := range src {
		srcVal := reflect.ValueOf(&src[i]).Elem()
		dstVal := reflect.ValueOf(&result[i]).Elem()
		mapAndCollect(srcVal, dstVal, info, collector)
	}

	// 5. 批量获取URL
	if len(collector.ids) > 0 {
		ids := make([]string, 0, len(collector.ids))
		for id := range collector.ids {
			ids = append(ids, id)
		}

		resources, err := filler.resolver.Resolve(ctx, ids)
		if err != nil {
			return err
		}

		// 6. 填充URL
		for i := range result {
			dstVal := reflect.ValueOf(&result[i]).Elem()
			fillURLs(dstVal, info, resources)
		}
	}

	*dst = result
	return nil
}

// AutoFillOne 自动映射并填充单个对象
//
// 参数:
//   - ctx: 上下文
//   - filler: 填充器
//   - src: 源对象指针
//   - dst: 目标对象指针
//
// 示例:
//
//	var response ProductResponse
//	image.AutoFillOne(ctx, filler, product, &response)
func AutoFillOne[S, D any](ctx context.Context, filler *Filler, src *S, dst *D) error {
	if src == nil || dst == nil {
		return nil
	}

	srcSlice := []S{*src}
	var dstSlice []D

	if err := AutoFill(ctx, filler, srcSlice, &dstSlice); err != nil {
		return err
	}

	if len(dstSlice) > 0 {
		*dst = dstSlice[0]
	}
	return nil
}

// ==================== 内部实现 ====================

// idCollector ID收集器
type idCollector struct {
	ids map[string]struct{}
}

func (c *idCollector) add(id string) {
	if id != "" {
		c.ids[id] = struct{}{}
	}
}

func (c *idCollector) addAll(ids []string) {
	for _, id := range ids {
		c.add(id)
	}
}

// getTypeInfo 获取类型信息（带缓存）
func getTypeInfo(srcType, dstType reflect.Type) *typeInfo {
	// 解引用指针
	srcType = deref(srcType)
	dstType = deref(dstType)

	pair := typePair{src: srcType, dst: dstType}
	if cached, ok := typeCache.Load(pair); ok {
		return cached.(*typeInfo)
	}

	info := buildTypeInfo(srcType, dstType)
	typeCache.Store(pair, info)
	return info
}

// deref 解引用指针类型
func deref(t reflect.Type) reflect.Type {
	for t.Kind() == reflect.Ptr {
		t = t.Elem()
	}
	return t
}

// buildTypeInfo 构建类型信息
func buildTypeInfo(srcType, dstType reflect.Type) *typeInfo {
	if srcType.Kind() != reflect.Struct || dstType.Kind() != reflect.Struct {
		return &typeInfo{}
	}

	// 构建源字段索引映射
	srcFields := make(map[string]int)
	for i := 0; i < srcType.NumField(); i++ {
		f := srcType.Field(i)
		if f.IsExported() {
			srcFields[f.Name] = i
		}
	}

	var fields []fieldInfo
	for i := 0; i < dstType.NumField(); i++ {
		dstField := dstType.Field(i)
		if !dstField.IsExported() {
			continue
		}

		srcIdx, ok := srcFields[dstField.Name]
		if !ok {
			continue
		}

		srcField := srcType.Field(srcIdx)
		fi := fieldInfo{
			srcIndex: srcIdx,
			dstIndex: i,
			name:     dstField.Name,
		}

		// 判断字段类型
		dstFieldType := dstField.Type
		switch {
		case dstFieldType == reflect.TypeOf(FileID("")):
			fi.fieldType = fieldTypeFileID
		case dstFieldType == reflect.TypeOf(FileIDs{}):
			fi.fieldType = fieldTypeFileIDs
		case dstFieldType == reflect.TypeOf(RichText("")):
			fi.fieldType = fieldTypeRichText
		case dstFieldType.Kind() == reflect.Slice:
			fi.fieldType = fieldTypeSlice
			fi.srcElem = srcField.Type.Elem()
			fi.dstElem = dstFieldType.Elem()
			fi.elemInfo = getTypeInfo(fi.srcElem, fi.dstElem)
		case dstFieldType.Kind() == reflect.Map:
			fi.fieldType = fieldTypeMap
			fi.keyType = dstFieldType.Key()
			fi.srcElem = srcField.Type.Elem()
			fi.dstElem = dstFieldType.Elem()
			fi.elemInfo = getTypeInfo(fi.srcElem, fi.dstElem)
		case deref(dstFieldType).Kind() == reflect.Struct && !isBasicType(dstFieldType):
			fi.fieldType = fieldTypeStruct
			fi.srcElem = srcField.Type
			fi.dstElem = dstFieldType
			fi.elemInfo = getTypeInfo(fi.srcElem, fi.dstElem)
		default:
			fi.fieldType = fieldTypeBasic
		}

		fields = append(fields, fi)
	}

	return &typeInfo{fields: fields}
}

// isBasicType 判断是否为基础类型（不需要递归）
func isBasicType(t reflect.Type) bool {
	t = deref(t)
	switch t.Kind() {
	case reflect.Bool, reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64,
		reflect.Float32, reflect.Float64, reflect.String:
		return true
	}
	// time.Time 等也视为基础类型
	if t.PkgPath() == "time" && t.Name() == "Time" {
		return true
	}
	return false
}

// mapAndCollect 映射字段并收集ID
func mapAndCollect(srcVal, dstVal reflect.Value, info *typeInfo, collector *idCollector) {
	// 解引用指针
	srcVal = derefValue(srcVal)
	dstVal = derefValue(dstVal)

	if !srcVal.IsValid() || !dstVal.IsValid() {
		return
	}

	for _, fi := range info.fields {
		srcField := srcVal.Field(fi.srcIndex)
		dstField := dstVal.Field(fi.dstIndex)

		switch fi.fieldType {
		case fieldTypeBasic:
			if srcField.Type().AssignableTo(dstField.Type()) {
				dstField.Set(srcField)
			} else if srcField.Type().ConvertibleTo(dstField.Type()) {
				dstField.Set(srcField.Convert(dstField.Type()))
			}

		case fieldTypeFileID:
			// 复制值并收集ID
			id := getStringValue(srcField)
			dstField.SetString(id)
			collector.add(id)

		case fieldTypeFileIDs:
			// 复制值并收集ID
			ids := getStringSliceValue(srcField)
			if len(ids) > 0 {
				slice := reflect.MakeSlice(dstField.Type(), len(ids), len(ids))
				for i, id := range ids {
					slice.Index(i).SetString(id)
				}
				dstField.Set(slice)
				collector.addAll(ids)
			}

		case fieldTypeRichText:
			// 复制值并提取ID
			text := getStringValue(srcField)
			dstField.SetString(text)
			matches := dataHelfRegex.FindAllStringSubmatch(text, -1)
			for _, m := range matches {
				if len(m) > 1 {
					collector.add(m[1])
				}
			}

		case fieldTypeSlice:
			mapSliceAndCollect(srcField, dstField, fi, collector)

		case fieldTypeMap:
			mapMapAndCollect(srcField, dstField, fi, collector)

		case fieldTypeStruct:
			mapStructAndCollect(srcField, dstField, fi, collector)
		}
	}
}

// mapSliceAndCollect 映射切片并收集ID
func mapSliceAndCollect(srcField, dstField reflect.Value, fi fieldInfo, collector *idCollector) {
	srcField = derefValue(srcField)
	if !srcField.IsValid() || srcField.IsNil() || srcField.Len() == 0 {
		return
	}

	length := srcField.Len()
	slice := reflect.MakeSlice(dstField.Type(), length, length)

	for i := 0; i < length; i++ {
		srcElem := srcField.Index(i)
		dstElem := slice.Index(i)

		// 如果目标是指针类型，需要创建新实例
		if fi.dstElem.Kind() == reflect.Ptr {
			newElem := reflect.New(fi.dstElem.Elem())
			dstElem.Set(newElem)
			mapAndCollect(srcElem, newElem.Elem(), fi.elemInfo, collector)
		} else {
			mapAndCollect(srcElem, dstElem, fi.elemInfo, collector)
		}
	}

	dstField.Set(slice)
}

// mapStructAndCollect 映射结构体并收集ID
func mapStructAndCollect(srcField, dstField reflect.Value, fi fieldInfo, collector *idCollector) {
	srcField = derefValue(srcField)
	if !srcField.IsValid() {
		return
	}

	// 如果目标是指针类型，需要创建新实例
	if fi.dstElem.Kind() == reflect.Ptr {
		newElem := reflect.New(fi.dstElem.Elem())
		dstField.Set(newElem)
		mapAndCollect(srcField, newElem.Elem(), fi.elemInfo, collector)
	} else {
		mapAndCollect(srcField, dstField, fi.elemInfo, collector)
	}
}

// mapMapAndCollect 映射map并收集ID（如多语言 map[string]*Lang）
func mapMapAndCollect(srcField, dstField reflect.Value, fi fieldInfo, collector *idCollector) {
	srcField = derefValue(srcField)
	if !srcField.IsValid() || srcField.IsNil() || srcField.Len() == 0 {
		return
	}

	// 创建目标map
	dstMap := reflect.MakeMap(dstField.Type())

	for _, key := range srcField.MapKeys() {
		srcElem := srcField.MapIndex(key)

		// 如果目标value是指针类型，需要创建新实例
		if fi.dstElem.Kind() == reflect.Ptr {
			newElem := reflect.New(fi.dstElem.Elem())
			mapAndCollect(srcElem, newElem.Elem(), fi.elemInfo, collector)
			dstMap.SetMapIndex(key, newElem)
		} else {
			newElem := reflect.New(fi.dstElem).Elem()
			mapAndCollect(srcElem, newElem, fi.elemInfo, collector)
			dstMap.SetMapIndex(key, newElem)
		}
	}

	dstField.Set(dstMap)
}

// fillURLs 填充URL
func fillURLs(dstVal reflect.Value, info *typeInfo, resources map[string]*ResourceInfo) {
	dstVal = derefValue(dstVal)
	if !dstVal.IsValid() {
		return
	}

	for _, fi := range info.fields {
		dstField := dstVal.Field(fi.dstIndex)

		switch fi.fieldType {
		case fieldTypeFileID:
			id := dstField.String()
			if res, ok := resources[id]; ok && res.Success {
				dstField.SetString(res.URL)
			}

		case fieldTypeFileIDs:
			if dstField.Len() > 0 {
				for i := 0; i < dstField.Len(); i++ {
					id := dstField.Index(i).String()
					if res, ok := resources[id]; ok && res.Success {
						dstField.Index(i).SetString(res.URL)
					}
				}
			}

		case fieldTypeRichText:
			text := dstField.String()
			newText := dataHelfRegex.ReplaceAllStringFunc(text, func(match string) string {
				m := dataHelfRegex.FindStringSubmatch(match)
				if len(m) > 1 {
					if res, ok := resources[m[1]]; ok && res.Success {
						// 将 data-helf="file_id" 替换为 src="url"
						return `src="` + res.URL + `"`
					}
				}
				return match
			})
			dstField.SetString(newText)

		case fieldTypeSlice:
			fillSliceURLs(dstField, fi, resources)

		case fieldTypeMap:
			fillMapURLs(dstField, fi, resources)

		case fieldTypeStruct:
			fillStructURLs(dstField, fi, resources)
		}
	}
}

// fillSliceURLs 填充切片中的URL
func fillSliceURLs(dstField reflect.Value, fi fieldInfo, resources map[string]*ResourceInfo) {
	dstField = derefValue(dstField)
	if !dstField.IsValid() || dstField.IsNil() {
		return
	}

	for i := 0; i < dstField.Len(); i++ {
		elem := dstField.Index(i)
		fillURLs(elem, fi.elemInfo, resources)
	}
}

// fillStructURLs 填充结构体中的URL
func fillStructURLs(dstField reflect.Value, fi fieldInfo, resources map[string]*ResourceInfo) {
	dstField = derefValue(dstField)
	if !dstField.IsValid() {
		return
	}
	fillURLs(dstField, fi.elemInfo, resources)
}

// fillMapURLs 填充map中的URL
func fillMapURLs(dstField reflect.Value, fi fieldInfo, resources map[string]*ResourceInfo) {
	dstField = derefValue(dstField)
	if !dstField.IsValid() || dstField.IsNil() {
		return
	}

	for _, key := range dstField.MapKeys() {
		elem := dstField.MapIndex(key)
		// map中的元素是不可寻址的，需要复制后修改再设置回去
		if elem.Kind() == reflect.Ptr && !elem.IsNil() {
			fillURLs(elem.Elem(), fi.elemInfo, resources)
		}
	}
}

// derefValue 解引用Value
func derefValue(v reflect.Value) reflect.Value {
	for v.Kind() == reflect.Ptr {
		if v.IsNil() {
			return reflect.Value{}
		}
		v = v.Elem()
	}
	return v
}

// getStringValue 获取字符串值
func getStringValue(v reflect.Value) string {
	v = derefValue(v)
	if !v.IsValid() {
		return ""
	}
	if v.Kind() == reflect.String {
		return v.String()
	}
	return ""
}

// getStringSliceValue 获取字符串切片值
func getStringSliceValue(v reflect.Value) []string {
	v = derefValue(v)
	if !v.IsValid() || v.Kind() != reflect.Slice {
		return nil
	}

	result := make([]string, v.Len())
	for i := 0; i < v.Len(); i++ {
		result[i] = getStringValue(v.Index(i))
	}
	return result
}
