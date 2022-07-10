package library

type GroupForm struct {
	GroupName string `form:"group_name" json:"group_name" binding:"required"`
	Ip        string `form:"ip" json:"ip" binding:"required"`
	Port      int    `form:"port" json:"port" binding:"required"`
	Addr      string `form:"addr" json:"addr" binding:"required"`
	ProbeType string `form:"probe_type" json:"probe_type" binding:"required"`
	ProbeUrl       string `form:"url" json:"probe_url"`
}
