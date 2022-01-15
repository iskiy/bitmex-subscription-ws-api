package gate

type BitmexCommand struct {
	Op string 			`json:"op"`
	Args []interface{} `json:"args,omitempty"`
}