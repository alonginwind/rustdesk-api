package admin

type OperationLogQuery struct {
	Username string `form:"username"`
	Resource string `form:"resource"`
	Op       string `form:"op"`
	PageQuery
}
