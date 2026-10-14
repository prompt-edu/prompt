package callDTO

import "encoding/json"

type Content struct {
	Request      json.RawMessage `json:"request"`
	ResponseText string          `json:"responseText"`
}
