package pageutil

// PageReq 分页请求参数.
type PageReq struct {
	Page     int `json:"page" form:"page"`           // 页码，默认1
	PageSize int `json:"page_size" form:"page_size"` // 每页大小，默认10，最大50
}

func NormalizePage(page, pageSize int) (int, int) {
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 10
	}
	if pageSize > 50 {
		pageSize = 50
	}
	return page, pageSize
}
