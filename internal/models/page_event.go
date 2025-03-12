package models

// Event properties for page interactions
type PageEvent struct {
	Page *PageEventPage `json:"page,omitempty"`

	Utm *PageEventUtm `json:"utm,omitempty"`

	Engagement *PageEventEngagement `json:"engagement,omitempty"`
}

type PageEventPage struct {
	// Full URL of the page
	Url string `json:"url,omitempty"`
	// Page path without domain
	Path string `json:"path,omitempty"`
	// URL of the previous page
	Referrer string `json:"referrer,omitempty"`
	// Title of the page
	Title string `json:"title,omitempty"`
	// Query parameters from the URL
	Search string `json:"search,omitempty"`
	// Logical category of the page
	PageCategory string `json:"page_category,omitempty"`
	// Type of page (e.g., landing_page, blog)
	PageType string `json:"page_type,omitempty"`
	// Type of content (e.g., article, video)
	ContentType string `json:"content_type,omitempty"`
	// Percentage of page scrolled
	ScrollDepth string `json:"scroll_depth,omitempty"`
	// Time spent on the page in seconds
	TimeOnPage int32 `json:"time_on_page,omitempty"`
	// Identifier for the previous page
	PreviousPage string `json:"previous_page,omitempty"`
}

type PageEventUtm struct {
	// Traffic source (e.g., google, facebook)
	Source string `json:"source,omitempty"`
	// Marketing medium (e.g., email, social)
	Medium string `json:"medium,omitempty"`
	// Campaign name
	Campaign string `json:"campaign,omitempty"`
}

type PageEventEngagement struct {
	// Custom engagement score
	EngagementScore float64 `json:"engagement_score,omitempty"`
	// List of interactive elements clicked
	InteractiveElements []string `json:"interactive_elements,omitempty"`
}
