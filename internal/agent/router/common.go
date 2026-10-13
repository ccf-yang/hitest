package router

func commonGroups() []CommonRouter {
	return []CommonRouter{
		&DashboardRouter{},
		&HostRouter{},
		&ToolboxRouter{},
		&ProcessRouter{},
		&QuickCommandRouter{},
	}
}
