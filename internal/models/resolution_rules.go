package models

// ResolutionRule represents rules for merging user profiles
type ResolutionRule struct {
	RuleId    string `json:"rule_id" bson:"rule_id" binding:"required"`
	RuleName  string `json:"rule_name" bson:"rule_name" binding:"required"`
	Attribute string `json:"attribute" bson:"attribute" binding:"required"`
	Priority  int    `json:"priority" bson:"priority" binding:"required"` // 0 = highest priority
	IsActive  bool   `json:"is_active" bson:"is_active" binding:"required"`
	CreatedAt int64  `json:"created_at" bson:"created_at"`
	UpdatedAt int64  `json:"updated_at" bson:"updated_at"`
}
