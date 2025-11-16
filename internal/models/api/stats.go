package api

type StatsResponse struct {
    TotalPR           int               `json:"total_pr"`
    OpenPR            int               `json:"open_pr"`
    ReviewAssignments map[string]int    `json:"review_assignments"` // user_id -> count
}