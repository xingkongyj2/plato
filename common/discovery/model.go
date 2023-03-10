package discovery

import (
	"encoding/json"
)

type EndpointInfo struct {
	IP   string `json:"ip"`
	Port string `json:"port"`
	//这是一种扩展式的写法，业务方以后增加什么字段了，直接插入就好了
	MetaData map[string]interface{} `json:"meta"`
}

func UnMarshal(data []byte) (*EndpointInfo, error) {
	ed := &EndpointInfo{}
	err := json.Unmarshal(data, ed)
	if err != nil {
		return nil, err
	}
	return ed, nil
}
func (edi *EndpointInfo) Marshal() string {
	data, err := json.Marshal(edi)
	if err != nil {
		panic(err)
	}
	return string(data)
}
