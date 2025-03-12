package models

type Profile struct {
	PermaID string `json:"perma_id"`

	OriginCountry string `json:"origin_country"`

	UserIds []string `json:"user_ids,omitempty"`

	Identity *IdentityData `json:"identity,omitempty"`

	Personality *PersonalityData `json:"personality,omitempty"`

	AppContext []AppContext `json:"app_context,omitempty"`
}
