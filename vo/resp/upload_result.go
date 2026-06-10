package resp

type UploadResultResp struct {
	Success int `json:"success"`
	Skipped int `json:"skipped"`
	Failed  int `json:"failed"`
}
