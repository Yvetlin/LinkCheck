package types

type LinkStatus string

const (
	StatusAvailable    LinkStatus = "available"
	StatusNotAvailable LinkStatus = "not available"
)

type LinkInfo struct {
	URL    string     `json:"url"`
	Status LinkStatus `json:"status"`
}

type LinksSet struct {
	LinksNum int                   `json:"links_num"`
	Links    map[string]LinkStatus `json:"links"`
}

// SubmitLinksRequest представляет запрос на проверку ссылок
type SubmitLinksRequest struct {
	Links []string `json:"links"`
}

// SubmitLinksResponse представляет ответ на запрос проверки ссылок
type SubmitLinksResponse struct {
	Links    map[string]string `json:"links"`
	LinksNum int               `json:"links_num"`
}

// GetReportRequest представляет запрос на получение PDF отчета
type GetReportRequest struct {
	LinksList []int `json:"links_list"`
}

// State представляет полное состояние приложения для сохранения в файл
type State struct {
	LinksSets    map[int]LinksSet `json:"links_sets"`
	NextLinksNum int              `json:"next_links_num"`
	PendingTasks []PendingTask    `json:"pending_tasks"`
}

// PendingTask представляет задачу проверки ссылки в очереди
type PendingTask struct {
	LinksNum int    `json:"links_num"`
	URL      string `json:"url"`
	Status   string `json:"status"`
}
