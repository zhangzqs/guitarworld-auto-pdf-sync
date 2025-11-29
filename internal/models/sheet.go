package models

// APIResponse 表示 API 响应结构
type APIResponse struct {
	State int `json:"state"`
	Data  struct {
		List       []SheetMusicItem `json:"list"`
		Pagination struct {
			Total       int `json:"total"`
			PerPage     int `json:"per_page"`
			CurrentPage int `json:"current_page"`
			LastPage    int `json:"last_page"`
		} `json:"pagination"`
	} `json:"data"`
	Message string `json:"message"`
}

// SheetMusicItem 表示单个曲谱信息
type SheetMusicItem struct {
	ID          int    `json:"id"`
	Type        int    `json:"type"`
	Name        string `json:"name"`
	SubTitle    string `json:"sub_title"`
	CategoryID  int    `json:"category_id"`
	Singer      string `json:"singer"`
	Author      string `json:"author"`
	CreatorName string `json:"creator_name"`
	Qupu        struct {
		CategoryTxt string `json:"category_txt"`
	} `json:"qupu"`
}

// ProcessResult 表示曲谱处理结果
type ProcessResult struct {
	Sheet   SheetMusicItem
	Success bool
	Error   error
	Skipped bool
}
