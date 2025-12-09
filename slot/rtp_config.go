package slot

import "encoding/json"

type RateWeight struct {
	Rate      float64 `json:"rate"`
	Weighting int     `json:"weighting"`
}

type RateConfig struct {
	Rate    float64      `json:"rate"`
	Normal  []RateWeight `json:"normal"`
	Special []RateWeight `json:"special"`
}

type RtpConfig struct {
	Use  string                `json:"use"`
	Data map[string]RateConfig `json:"-"`
}

// 自定义UnmarshalJSON方法处理动态键
func (c *RtpConfig) UnmarshalJSON(data []byte) error {
	type Alias RtpConfig
	aux := &struct {
		Data map[string]RateConfig `json:"-"`
		*Alias
	}{
		Alias: (*Alias)(c),
	}

	if err := json.Unmarshal(data, &aux.Alias); err != nil {
		return err
	}

	// 处理动态键
	var raw map[string]json.RawMessage
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}

	c.Data = make(map[string]RateConfig)
	for key, value := range raw {
		if key == "use" {
			continue
		}
		var rc RateConfig
		if err := json.Unmarshal(value, &rc); err != nil {
			return err
		}
		c.Data[key] = rc
	}

	return nil
}
