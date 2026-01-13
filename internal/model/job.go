package model

// JobQueryRequest 岗位查询请求
type JobQueryRequest struct {
	Current             int    `json:"current" form:"current"`                                   // 当前页码
	PageSize            int    `json:"pageSize" form:"pageSize"`                                 // 每页数量
	JobTitle            string `json:"jobTitle,omitempty" form:"jobTitle"`                       // 岗位名称
	Latitude            string `json:"latitude,omitempty" form:"latitude"`                       // 纬度
	Longitude           string `json:"longitude,omitempty" form:"longitude"`                     // 经度
	Radius              string `json:"radius,omitempty" form:"radius"`                           // 搜索半径（km）
	Order               string `json:"order,omitempty" form:"order"`                             // 排序方式: 0-推荐, 1-最热, 2-最新
	MinSalary           string `json:"minSalary,omitempty" form:"minSalary"`                     // 最低薪资
	MaxSalary           string `json:"maxSalary,omitempty" form:"maxSalary"`                     // 最高薪资
	Experience          string `json:"experience,omitempty" form:"experience"`                   // 经验要求代码
	Education           string `json:"education,omitempty" form:"education"`                     // 学历要求代码
	CompanyNature       string `json:"companyNature,omitempty" form:"companyNature"`             // 企业类型代码
	JobLocationAreaCode string `json:"jobLocationAreaCode,omitempty" form:"jobLocationAreaCode"` // 区域代码
}

// JobAPIResponse 岗位API响应
type JobAPIResponse struct {
	Code int          `json:"code"`
	Msg  string       `json:"msg"`
	Rows []JobListing `json:"rows"`
	Data interface{}  `json:"data,omitempty"`
}

// JobListing 岗位信息
type JobListing struct {
	JobTitle            string `json:"jobTitle"`            // 职位名称
	CompanyName         string `json:"companyName"`         // 公司名称
	MinSalary           int    `json:"minSalary"`           // 最低薪资
	MaxSalary           int    `json:"maxSalary"`           // 最高薪资
	Education           string `json:"education"`           // 学历要求代码
	Experience          string `json:"experience"`          // 经验要求代码
	AppJobURL           string `json:"appJobUrl"`           // 职位链接
	JobLocationAreaCode int    `json:"jobLocationAreaCode"` // 工作地点代码
}

// FormattedJob 格式化后的岗位信息
type FormattedJob struct {
	JobTitle    string      `json:"jobTitle"`       // 职位名称
	CompanyName string      `json:"companyName"`    // 公司名称
	Salary      string      `json:"salary"`         // 薪资范围
	Location    string      `json:"location"`       // 工作地点
	Education   string      `json:"education"`      // 学历要求
	Experience  string      `json:"experience"`     // 经验要求
	AppJobURL   string      `json:"appJobUrl"`      // 职位链接
	Data        interface{} `json:"data,omitempty"` // 额外数据（最后一条时包含）
}

// JobResponse 岗位查询结果
type JobResponse struct {
	JobListings []FormattedJob `json:"jobListings"`
	Data        interface{}    `json:"data,omitempty"`
}

// 学历代码映射
var EducationMap = map[string]string{
	"-1": "学历不限",
	"0":  "初中及以下",
	"1":  "中专/中技",
	"2":  "高中",
	"3":  "大专",
	"4":  "本科",
	"5":  "硕士",
	"6":  "博士",
	"7":  "MBA/EMBA",
	"8":  "留学-学士",
	"9":  "留学-硕士",
	"10": "留学-博士",
}

// 经验代码映射
var ExperienceMap = map[string]string{
	"0": "经验不限",
	"1": "实习生",
	"2": "应届毕业生",
	"3": "1年以下",
	"4": "1-3年",
	"5": "3-5年",
	"6": "5-10年",
	"7": "10年以上",
}

// 企业类型代码映射
var CompanyNatureMap = map[string]string{
	"1": "私营企业",
	"2": "股份制企业",
	"3": "国有企业",
	"4": "外商及港澳台投资企业",
	"5": "医院",
}

// AmapPlaceResponse 高德地图地点查询响应
type AmapPlaceResponse struct {
	Status string      `json:"status"`
	Info   string      `json:"info"`
	Pois   []AmapPlace `json:"pois"`
}

// AmapPlace 地点信息
type AmapPlace struct {
	Name     string `json:"name"`
	Location string `json:"location"` // "经度,纬度"
	Address  string `json:"address"`
}
