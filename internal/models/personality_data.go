package models

type PersonalityData struct {
	Interests []string `json:"interests,omitempty"`

	PreferredLanguage string `json:"preferred_language,omitempty"`

	CommunicationPreferences *PersonalityDataCommunicationPreferences `json:"communication_preferences,omitempty"`

	ShoppingPreferences *PersonalityDataShoppingPreferences `json:"shopping_preferences,omitempty"`
}

type PersonalityDataCommunicationPreferences struct {
	EmailNotifications bool `json:"email_notifications,omitempty"`

	SmsNotifications bool `json:"sms_notifications,omitempty"`

	PushNotifications bool `json:"push_notifications,omitempty"`
}

type PersonalityDataShoppingPreferences struct {
	FavoriteBrands []string `json:"favorite_brands,omitempty"`

	DiscountPreference string `json:"discount_preference,omitempty"`
}
