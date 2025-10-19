package controllers

import "time"

type UtilsController struct{}

func formatWIB(t *time.Time) string {
	if t == nil {
		return ""
	}
	loc, _ := time.LoadLocation("Asia/Jakarta")
	return t.In(loc).Format("2006-01-02 15:04:05")
}
