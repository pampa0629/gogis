package mapping

import (
	"encoding/json"
	"gogis/base"
	"gogis/data"
	"gogis/draw"
)

// 图层类
type Layer struct {
	Name   string          `json:"LayerName"` // 图层名
	feaset data.Featureset // 数据来源
	Params data.ConnParams `json:"ConnParams"` // 存储和打开地图文档时用的数据连接信息
	Type   ThemeType       `json:"ThemeType"`
	theme  Theme           // 专题风格
	Object interface{}     `json:"Theme"` // 好一招狸猫换太子
}

func NewLayer(feaset data.Featureset, theme Theme) *Layer {
	layer := new(Layer)
	layer.Setting(feaset)
	// 默认图层名 等于 数据集名
	layer.Name = layer.Params["name"].(string)
	if theme == nil {
		layer.theme = new(SimpleTheme)
		layer.theme.MakeDefault(feaset)
	} else {
		layer.theme = theme
		layer.Name += "_" + string(theme.GetType())
	}
	layer.Type = layer.theme.GetType()

	return layer
}

func (this *Layer) UnmarshalJSON(data []byte) error {
	type cloneType Layer
	rawMsg := json.RawMessage{}
	this.Object = &rawMsg
	json.Unmarshal(data, (*cloneType)(this))

	this.theme = NewTheme(this.Type)
	json.Unmarshal(rawMsg, this.theme)
	return nil
}

// new出来的时候，做统一设置
func (this *Layer) Setting(feaset data.Featureset) bool {
	this.feaset = feaset
	store := feaset.GetStore()
	if store != nil {
		this.Params = store.GetConnParams()
		this.Params["name"] = feaset.GetName()
		return true
	}
	return false
}

// 地图 Save时，内部存储调整为相对路径
func (this *Layer) WhenSaving(mappath string) {
	// 拷贝后使用，避免 map save之后，filename变为相对路径，再打开数据就不好使了
	// newParams := base.DeepCopy(this.Params).(data.ConnParams)
	filename, ok := this.Params["filename"]
	if ok {
		storename := filename.(string)
		if len(storename) > 0 {
			this.Params["filename"] = base.GetRelativePath(mappath, storename)
		}
	}
	// 保证专题图类型的存储和读取
	this.Object = this.theme
}

// 地图Open时调用，加载实际数据，准备绘制
func (this *Layer) WhenOpenning(mappath string) {
	store := data.NewDatastore(data.StoreType(this.Params["type"].(string)))
	if store != nil {
		// 如果有文件路径，则需要恢复为绝对路径
		filename, ok := this.Params["filename"]
		if ok {
			storename := filename.(string)
			if len(storename) > 0 {
				this.Params["filename"] = base.GetAbsolutePath(mappath, storename)
			}
		}
		ok, _ = store.Open(this.Params)
		if ok {
			this.feaset, _ = store.GetFeasetByName(this.Params["name"].(string))
			this.feaset.Open()
		}
	}
	if this.theme != nil {
		this.theme.WhenOpenning()
	}
}

func (this *Layer) Draw(canvas *draw.Canvas) (objCount int64) {
	// tr := base.NewTimeRecorder()
	feait := this.feaset.QueryByBounds(canvas.Params.GetBounds())
	if this.theme != nil {
		objCount = this.theme.Draw(canvas, feait)
	}
	feait.Close()
	return
}
