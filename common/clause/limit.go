package clause

type LimitClause struct {
	// 每页大小
	LimitN *int `json:"limit"`
	// 偏移量
	Offset int `json:"offset"`
}

// SetLimitN /** 设置每页大小 */
func (limitN LimitClause) SetLimitN(newLimit *int) {
	limitN.LimitN = newLimit
}
