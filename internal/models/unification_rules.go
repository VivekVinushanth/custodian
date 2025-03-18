package models

// Rule represents an attribute and its assigned priority
type Rule struct {
	Attribute     string `json:"attribute" bson:"attribute" binding:"required"`
	Priority      int    `json:"priority" bson:"priority" binding:"required"`             // 0 = highest priority
	MergeStrategy string `json:"merge_strategy" bson:"merge_strategy" binding:"required"` // Strategy for merging profiles
}

// UnificationRule represents rules for merging user profiles
type UnificationRule struct {
	RuleName  string `json:"rule_name" bson:"rule_name" binding:"required"`
	Rules     []Rule `json:"rules" bson:"rules" binding:"required"` // List of attributes with priority
	IsActive  bool   `json:"is_active" bson:"is_active"`
	CreatedAt int64  `json:"created_at" bson:"created_at"`
	UpdatedAt int64  `json:"updated_at" bson:"updated_at"`
}
