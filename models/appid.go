package models

import (
	"time"

	"gorm.io/gorm"
)

type AppInfo struct {
	Id uint64 `json:"id" gorm:"primaryKey;autoIncrement;type:bigint unsigned"`

	Name string `json:"name" gorm:"column:name;comment:商户名称;"`

	AppId           string `json:"appId" gorm:"column:app_id;comment:应用ID;uniqueIndex"`
	CallBackUrl     string `json:"callBackUrl" gorm:"comment:api回调接口,需要商户提供;"`
	Currency        string `json:"currency" gorm:"column:currency;comment:货币类型;"`
	AccessKeyId     string `json:"accessKey" gorm:"column:access_key;comment:访问密钥Id;index"`
	AccessKeySecret string `json:"accessKeySecret" gorm:"column:access_secret;comment:访问密钥;"`
	Country         string `json:"country" gorm:"column:country;comment:国家如中国cn,美国us;"`
	TimeZone        string `json:"timeZone" gorm:"column:time_zone;comment:时区;default:'Asia/Kolkata'"`

	Rtp string `json:"rtp" gorm:"column:rtp;comment:默认rtp;default:95"`

	State            uint8   `json:"state" gorm:"column:state;comment:状态,0正常,1禁用;default:0"`
	Rate             float64 `json:"rate"  gorm:"column:rate;comment:费率;"`
	Note             string  `json:"note" gorm:"column:note;comment:备注;"`
	TriggerWinIfZero uint8   `json:"triggerWinIfZero" gorm:"column:trigger_win_if_zero;comment:派奖为0是否回调：0否, 1是;default:1"`

	CreateTime        time.Time      `json:"createTime" gorm:"autoCreateTime;comment:创建时间;"`
	UpdateTime        time.Time      `json:"updateTime" gorm:"autoCreateTime;comment:创建时间;"`
	DeletedAt         gorm.DeletedAt `json:"-" gorm:"column:deleted_at;type:datetime(3);index"`
	ShardingState     uint8          `json:"shardingState" gorm:"column:sharding_state;comment:分表状态,0否,1是;default:0"`
	ShardingStartDate *time.Time     `json:"shardingStartDate" gorm:"column:sharding_start_date;comment:分表开始日期;default:null"`
}

func (AppInfo) TableName() string {
	return "app_info"
}
