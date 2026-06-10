package resp

type LoginResp struct {
	Token     string `json:"token"`
	TokenName string `json:"token_name"`
}
