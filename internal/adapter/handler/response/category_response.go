package response

type SuccessCategoryResponse struct {
	ID            int64  `json:"id"`
	Title         string `json:"title"`
	Slug          string `json:"string"`
	CreatedByName string `json:"created_by_name"`
}