package data

import (
	"gogis/base"
	"strconv"
	"strings"
	"time"
)

type FieldType int32

const (
	TypeUnknown FieldType = 0
	TypeBool    FieldType = 1 // bool
	TypeInt     FieldType = 2 // int32
	TypeFloat   FieldType = 3 // float64
	TypeString  FieldType = 5 // string
	TypeTime    FieldType = 6 // time.Time
	TypeBlob    FieldType = 7 // []byte
)

// 字段描述信息
type FieldInfo struct {
	Name   string
	Type   FieldType
	Length int
}

// 字段比较
type FieldComp struct {
	Log   string // "and" "or" ""
	Name  string
	Op    string
	Value interface{}
	Type  FieldType
}

func (this *FieldComp) Match(atts Atts) bool {
	switch this.Type {
	case TypeBool:
		if !base.IsMatchBool(atts[this.Name].(bool), this.Op, this.Value.(bool)) {
			return false
		}
	case TypeInt:
		if !base.IsMatchInt(atts[this.Name].(int), this.Op, this.Value.(int)) {
			return false
		}
	case TypeFloat:
		if !base.IsMatchFloat(atts[this.Name].(float64), this.Op, this.Value.(float64)) {
			return false
		}
	case TypeString:
		if !base.IsMatchString(atts[this.Name].(string), this.Op, this.Value.(string)) {
			return false
		}
	case TypeTime:
		if !base.IsMatchTime(atts[this.Name].(time.Time), this.Op, this.Value.(time.Time)) {
			return false
		}
	case TypeBlob:
		// 暂不支持
	}
	// 每一条都符合，才能通过
	return true
}

func (this *FieldComp) Parse(where string, finfos []FieldInfo) {
	// 先确定 and or
	if posOr := strings.Index(where, "or"); posOr >= 0 {
		this.Log = "or"
		where = where[posOr+2:]
	} else if posAnd := strings.Index(where, "and"); posAnd >= 0 {
		this.Log = "and"
		where = where[posAnd+3:]
	}

	// 再找“比较”字符串
	var value string
	this.Name, this.Op, value = findCompOp(where)
	this.Type = GetFieldTypeByName(finfos, this.Name)
	this.Value = String2value(value, this.Type)
}

// 查找比较字符串，返回分开的三个字符串
func findCompOp(where string) (field, op, value string) {
	ops := []string{">=", "<=", "!=", "=", ">", "<"} // todo  <>
	for _, v := range ops {
		if pos := strings.Index(where, v); pos >= 0 {
			field = strings.TrimSpace(where[0:pos])
			// field = strings.ToUpper(strings.TrimSpace(where[0:pos]))
			op = v
			value = strings.TrimSpace(where[pos+len(v):])
			break
		}
	}
	return
}

type FieldComps struct {
	Log   string // "and" "or" ""
	Comps []interface{}
}

// 看这个fea是否满足要求
func (this *FieldComps) Match(atts Atts) (res bool) {
	res = true
	// log := this.Log
	for _, v := range this.Comps {
		if comp, ok := v.(FieldComp); ok {
			match := comp.Match(atts)
			res = LogicalJudge(res, comp.Log, match)
		} else if comps, ok := v.(FieldComps); ok {
			match := comps.Match(atts)
			res = LogicalJudge(res, comps.Log, match)
		}
	}
	return
}

// 判断 one op two 的逻辑结果
// op: and / or
func LogicalJudge(one bool, op string, two bool) bool {
	switch op {
	case "and", "":
		return one && two
	case "or":
		return one || two

	}
	return false
}

// 解释where子句，支持 and or 和 () 可嵌套
func (this *FieldComps) Parse(where string, finfos []FieldInfo) {
	this.Comps = this.Comps[:0]
	// 1) 去掉前后的空格
	where = strings.TrimSpace(where)
	// 2) 遍历字符串，处理(); 如果遇到 ( 则 找到配套的 );  即后面遇到的(都需要记录下来,直到遇到对应的) 再停下
	//   2.1) 去掉最外层的(), 构建一个 FieldComps
	//   2.2) 记录()外面的 and/or, 没有认为是 and
	clause := this.parseBrackets(where, finfos)
	// 3) 之后再处理剩余的部分,构建FieldComp
	this.parseClause(clause, finfos)
}

// 先处理括号，返回不带括号的where
func (this *FieldComps) parseBrackets(where string, finfos []FieldInfo) string {
	// 开始遍历
	for i := 0; i < len(where); i++ {
		if pos := strings.Index(where, "("); pos >= 0 {
			clause, end := findBrackets(where, pos)
			var comps FieldComps
			comps.Parse(clause, finfos)
			andOr, before := findLastAndOr(where[0:pos])
			comps.Log = andOr
			this.Comps = append(this.Comps, comps)
			where = before + " " + where[end:]
		} else { // 找不到( ，则可以跳出循环
			break
		}
	}
	return where
}

// 解析不带括号的子句，解析and or
func (this *FieldComps) parseClause(clause string, finfos []FieldInfo) {
	clause = strings.TrimSpace(clause)
	for i := 0; i < len(clause); i++ {
		_, one, remain := splitWhithAndOr(clause)
		var comp FieldComp
		// comp.Log = andOr
		comp.Parse(one, finfos)
		this.Comps = append(this.Comps, comp)

		clause = remain
	}
}

// 从where找到对应的()内容，start是(位置, 返回()内不含括号的字符串,以及 ) 的位置
func findBrackets(where string, start int) (clause string, end int) {
	indent := 0
	for i, v := range where {
		switch v {
		case '(':
			indent++
		case ')':
			indent--
			if indent == 0 {
				clause = where[start+1 : i]
				end = i + 1
				return
			}
		}
	}
	return
}

// 从后往前查找 and/or，找到就返回之，找不到返回""；同时返回去掉and/or的字符串
func findLastAndOr(where string) (string, string) {
	posAnd := strings.LastIndex(where, "and")
	posOr := strings.LastIndex(where, "or")
	if posAnd >= 0 || posOr >= 0 {
		var andOr string
		if posAnd > posOr {
			andOr = "and"
		} else {
			andOr = "or"
		}
		return andOr, where[0:base.IntMax(posAnd, posOr)]
	}
	return "", where
}

// 从前往后查找 and/or，并返回查找结果；同时返回前后两段子句
func splitWhithAndOr(where string) (string, string, string) {
	where = strings.TrimSpace(where) // 去掉后面的空格
	posAnd := strings.Index(where, "and")
	posOr := strings.Index(where, "or")
	// 0 开始的不要
	if posAnd > 0 || posOr > 0 {
		if posAnd <= 0 {
			posAnd = len(where) + 1
		}
		if posOr <= 0 {
			posOr = len(where) + 1
		}
		var andOr string
		if posAnd < posOr {
			andOr = "and"
		} else {
			andOr = "or"
		}
		pos := base.IntMin(posAnd, posOr)
		return andOr, where[:pos], where[pos:]
	}
	// 都没找到，则返回自身
	return "", where, ""
}

func GetFieldTypeByName(finfos []FieldInfo, name string) FieldType {
	for _, finfo := range finfos {
		if strings.ToUpper(finfo.Name) == strings.ToUpper(name) {
			return finfo.Type
		}
	}
	return TypeUnknown
}

const TIME_LAYOUT = "2006-01-02 15:04:05"

// 把字符串性质的值，根据字段类型，转化为特定数据类型
func String2value(str string, ftype FieldType) interface{} {
	switch ftype {
	case TypeBool:
		// case "1", "t", "T", "true", "TRUE", "True":
		// case "0", "f", "F", "false", "FALSE", "False":
		value, _ := strconv.ParseBool(str)
		return value
	case TypeInt:
		value, _ := strconv.Atoi(str)
		return value
	case TypeFloat:
		value, _ := strconv.ParseFloat(str, 64)
		return value
	case TypeString:
		return str
	case TypeTime:
		value, _ := time.Parse(TIME_LAYOUT, str)
		return value
	case TypeBlob:
		return []byte(str)
	}
	return nil
}
