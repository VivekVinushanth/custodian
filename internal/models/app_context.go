package models

// AppContext represents contextual data for an application
type AppContext struct {
	AppID            string                      `json:"app_id" bson:"app_id" binding:"required"`
	SubscriptionPlan string                      `json:"subscription_plan,omitempty" bson:"subscription_plan,omitempty"`
	AppPermissions   []string                    `json:"app_permissions,omitempty" bson:"app_permissions,omitempty"`
	FeatureFlags     *AppContextFeatureFlags     `json:"feature_flags,omitempty" bson:"feature_flags,omitempty"`
	LastActiveApp    string                      `json:"last_active_app,omitempty" bson:"last_active_app,omitempty"`
	UsageMetrics     *AppContextUsageMetrics     `json:"usage_metrics,omitempty" bson:"usage_metrics,omitempty"`
	Devices          []AppContextDevices         `json:"devices,omitempty" bson:"devices,omitempty"`
	RegionsAccessed  []AppContextRegionsAccessed `json:"regions_accessed,omitempty" bson:"regions_accessed,omitempty"`
}

// AppContextFeatureFlags represents feature toggles
type AppContextFeatureFlags struct {
	DarkMode     bool `json:"dark_mode,omitempty" bson:"dark_mode,omitempty"`
	Experimental bool `json:"experimental,omitempty" bson:"experimental,omitempty"`
}

// AppContextUsageMetrics stores app usage details
type AppContextUsageMetrics struct {
	SessionCount int `json:"session_count,omitempty" bson:"session_count,omitempty"`
	ActiveTime   int `json:"active_time,omitempty" bson:"active_time,omitempty"` // in minutes
}

// AppContextDevices represents user devices
type AppContextDevices struct {
	DeviceID       string `json:"device_id,omitempty" bson:"device_id,omitempty"`
	DeviceType     string `json:"device_type,omitempty" bson:"device_type,omitempty"`
	LastUsed       int    `json:"last_used,omitempty" bson:"last_used,omitempty"`
	Os             string `json:"os,omitempty" bson:"os,omitempty"`
	Browser        string `json:"browser,omitempty" bson:"browser,omitempty"`
	BrowserVersion string `json:"browser_version,omitempty" bson:"browser_version,omitempty"`
	Ip             string `json:"ip,omitempty" bson:"ip,omitempty"`
}

// AppContextRegionsAccessed stores region-based access
type AppContextRegionsAccessed struct {
	RegionName  string `json:"region_name,omitempty" bson:"region_name,omitempty"`
	AccessCount int    `json:"access_count,omitempty" bson:"access_count,omitempty"`
}
