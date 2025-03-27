import ballerina/http;

service /customer\-data/'1\.0\.0 on new http:Listener(9090) {
    # Delete event schema
    #
    # + attribute - Attribute path to delete
    # + return - Attribute deleted 
    resource function delete schema/event/[string event_type](string attribute) returns http:Ok {
        return http:OK;
    }

    # Delete profile schema
    #
    # + attribute - Attribute path to delete
    # + return - Schema entry deleted 
    resource function delete schema/profile(string attribute) returns http:NoContent {
        return http:NO_CONTENT;
    }

    # Fetch events emitted by a user
    #
    # + perma_id - Unique identifier for the profile
    # + app_id - Unique identifier for the application
    # + searchString - Optional search filter
    # + offset - Number of records to skip for pagination
    # + 'limit - Maximum number of records to return
    # + return - returns can be any of following types 
    # http:Ok (Events retrieved successfully)
    # http:BadRequest (Bad input parameter)
    resource function get [string perma_id]/[string app_id]/events(string? searchString, int? offset, int:Signed32? 'limit) returns Event[]|http:BadRequest {
        return [];
    }

    # Fetch specific event emitted by a user
    #
    # + perma_id - Unique identifier for the profile
    # + app_id - Unique identifier for the application
    # + event_id - Unique identifier for the event
    # + searchString - Optional search filter
    # + offset - Number of records to skip for pagination
    # + 'limit - Maximum number of records to return
    # + return - returns can be any of following types 
    # http:Ok (Events retrieved successfully)
    # http:BadRequest (Bad input parameter)
    resource function get [string perma_id]/[string app_id]/events/[string event_id](string? searchString, int? offset, int:Signed32? 'limit) returns Event|http:BadRequest {
        return {};
    }

    # Fetch alias of a user
    #
    # + perma_id - Unique identifier for the user
    # + return - returns can be any of following types 
    # http:Ok (Fetch alias of a profile)
    # http:BadRequest (Bad input parameter)
    resource function get [string perma_id]/alias() returns string[]|http:BadRequest {
        return [];
    }

    # Fetch users bounded to a profile
    #
    # + perma_id - Unique identifier for the user
    # + return - returns can be any of following types 
    # http:Ok (Fetch alias of a profile)
    # http:BadRequest (Bad input parameter)
    resource function get [string perma_id]/bindUsers() returns string[]|http:BadRequest {
        return [];
    }

    # Fetch profileconsent
    #
    # + perma_id - Unique identifier for the user
    # + return - returns can be any of following types 
    # http:Ok (Fetch list of consented apps to collect data)
    # http:BadRequest (Bad input parameter)
    resource function get [string perma_id]/consent() returns Consent|http:BadRequest {
        return {};
    }

    # Fetch applications users has given consent to collect
    #
    # + perma_id - Unique identifier for the user
    # + return - returns can be any of following types 
    # http:Ok (Fetch list of consented apps to collect data)
    # http:BadRequest (Bad input parameter)
    resource function get [string perma_id]/consent/collect() returns string[]|http:BadRequest {
        return [];
    }

    # Fetch applications users has given consent to share
    #
    # + perma_id - Unique identifier for the user
    # + return - returns can be any of following types 
    # http:Ok (Fetch alias of a profile)
    # http:BadRequest (Bad input parameter)
    resource function get [string perma_id]/consent/share() returns string[]|http:BadRequest {
        return [];
    }

    # Fetch 360 profile of a user
    #
    # + perma_id - Unique identifier for the user
    # + includeAppContext - Whether to include application context data
    # + app_id - Application ID for fetching profile data with app context
    # + return - returns can be any of following types 
    # http:Ok (Personality data retrieved successfully)
    # http:BadRequest (Bad input parameter)
    resource function get [string perma_id]/profile(string? app_id, boolean includeAppContext = false) returns inline_response_200|http:BadRequest {
        return {};
    }

    # Fetch App context of the user
    #
    # + perma_id - Unique identifier for the user
    # + app_id - Unique identifier for the application
    # + return - returns can be any of following types 
    # http:Ok (Personality data retrieved successfully)
    # http:BadRequest (Bad input parameter)
    resource function get [string perma_id]/profile/[string app_id]/app_context() returns AppContext|http:BadRequest {
        return {
            app_id: "",
            subscription_plan: "",
            app_permissions: [],
            feature_flags: {
                beta_features_enabled: false,
                dark_mode: false
            },
            last_active_app: "",
            usage_metrics: {
                daily_active_time: 0,
                monthly_logins: 0
            },
            devices: [],
            regions_accessed: []
        };
    }

    # Fetch App context of the user
    #
    # + perma_id - Unique identifier for the user
    # + return - returns can be any of following types 
    # http:Ok (Personality data retrieved successfully)
    # http:BadRequest (Bad input parameter)
    resource function get [string perma_id]/profile/app_context() returns AppContext[]|http:BadRequest {
        return [];
    }

    # Fetch 360 profile of a user
    #
    # + perma_id - Unique identifier for the user
    # + return - returns can be any of following types 
    # http:Ok (Personality data retrieved successfully)
    # http:BadRequest (Bad input parameter)
    resource function get [string perma_id]/profile/personality() returns PersonalityData|http:BadRequest {
        return {
            interests: [],
            preferred_language: "",
            communication_preferences: {
                email_notifications: false,
                sms_notifications: false,
                push_notifications: false
            },
            shopping_preferences: {
                favorite_brands: [],
                discount_preference: ""
            }
        };
    }

    # Get event schema definition by type
    #
    # + return - Event schema retrieved 
    resource function get schema/event/[string event_type]() returns SchemaDefinition[] {
        return [];
    }

    # Get profile schema definition
    #
    # + return - Profile schema retrieved 
    resource function get schema/profile() returns SchemaDefinition[] {
        return [];
    }

    # update personality data for a user
    #
    # + perma_id - Unique identifier for the profile
    # + app_id - Unique identifier for the application
    # + return - returns can be any of following types 
    # http:Created (Personality data added successfully)
    # http:BadRequest (Invalid input data)
    resource function patch [string perma_id]/profile/[string app_id]/app_context(@http:Payload AppContext payload) returns http:Created|http:BadRequest {
        return http:CREATED;
    }

    # update personality data for a user
    #
    # + perma_id - Unique identifier for the profile
    # + return - returns can be any of following types 
    # http:Created (Personality data added successfully)
    # http:BadRequest (Invalid input data)
    resource function patch [string perma_id]/profile/personality(@http:Payload PersonalityData payload) returns http:Created|http:BadRequest {
        return http:CREATED;
    }

    # Patch event schema
    #
    # + return - Event schema patched 
    resource function patch schema/event/[string event_type](@http:Payload SchemaDefinition[] payload) returns http:Ok {
        return http:OK;
    }

    # Patch profile schema
    #
    # + return - returns can be any of following types 
    # http:Ok (Schema patched successfully)
    # http:BadRequest (Bad request)
    # http:InternalServerError (Server encountered error while responding to the request)
    resource function patch schema/profile(@http:Payload SchemaDefinition[] payload) returns http:Ok|http:BadRequest|http:InternalServerError {
        return http:OK;
    }

    # Push profile events
    #
    # + perma_id - Unique identifier for the profile
    # + app_id - Unique identifier for the application
    # + return - returns can be any of following types 
    # http:Created (Event added successfully)
    # http:BadRequest (Invalid event object)
    resource function post [string perma_id]/[string app_id]/event(@http:Payload Event payload) returns http:Created|http:BadRequest {
        return http:CREATED;
    }

    # Push user events
    #
    # + perma_id - Unique identifier for the user
    # + app_id - Unique identifier for the application
    # + return - returns can be any of following types 
    # http:Created (Event added successfully)
    # http:BadRequest (Invalid event object)
    resource function post [string perma_id]/[string app_id]/events(@http:Payload Event payload) returns http:Created|http:BadRequest {
        return http:CREATED;
    }

    # Merge different identities of a known user
    #
    # + perma_id - Unique identifier for the profile
    # + return - returns can be any of following types 
    # http:Created (Alias successfully updated)
    # http:BadRequest (Invalid alias request)
    resource function post [string perma_id]/alias(@http:Payload Alias payload) returns http:Created|http:BadRequest {
        return http:CREATED;
    }

    # bind user identities to a known profile
    #
    # + perma_id - Unique identifier for the profile
    # + return - returns can be any of following types 
    # http:Created (Alias successfully updated)
    # http:BadRequest (Invalid alias request)
    resource function post [string perma_id]/bindUsers(@http:Payload string[] payload) returns http:Created|http:BadRequest {
        return http:CREATED;
    }

    # provide consent to collect data
    #
    # + perma_id - Unique identifier for the profile
    # + return - returns can be any of following types 
    # http:Created (Alias successfully updated)
    # http:BadRequest (Invalid alias request)
    resource function post [string perma_id]/consent/collect(@http:Payload string[] payload) returns http:Created|http:BadRequest {
        return http:CREATED;
    }

    # provide consent to share profile
    #
    # + perma_id - Unique identifier for the profile
    # + return - returns can be any of following types 
    # http:Created (Alias successfully updated)
    # http:BadRequest (Invalid alias request)
    resource function post [string perma_id]/consent/share(@http:Payload string[] payload) returns http:Created|http:BadRequest {
        return http:CREATED;
    }

    # Create a user profile
    #
    # + return - returns can be any of following types 
    # http:Created (Alias successfully updated)
    # http:BadRequest (Invalid alias request)
    resource function post profile(@http:Payload ProfileData payload) returns http:Created|http:BadRequest {
        return http:CREATED;
    }

    # Add new event schema definition
    #
    # + return - Event schema added 
    resource function post schema/event/[string event_type](@http:Payload SchemaDefinition[] payload) returns http:Created {
        return http:CREATED;
    }

    # Add new profile schema definition
    #
    # + return - Schema added successfully 
    resource function post schema/profile(@http:Payload SchemaDefinition[] payload) returns http:Created {
        return http:CREATED;
    }

    # update personality data for a user
    #
    # + app_id - Unique identifier for the application
    # + return - returns can be any of following types 
    # http:Created (Personality data added successfully)
    # http:BadRequest (Invalid input data)
    resource function put [string perma_id]/profile/[string app_id]/app_context(@http:Payload AppContext payload) returns http:Created|http:BadRequest {
        return http:CREATED;
    }

    # update personality data for a user
    #
    # + perma_id - Unique identifier for the profile
    # + return - returns can be any of following types 
    # http:Created (Personality data added successfully)
    # http:BadRequest (Invalid input data)
    resource function put [string perma_id]/profile/personality(@http:Payload PersonalityData payload) returns http:Created|http:BadRequest {
        return http:CREATED;
    }

    # Replace event schema
    #
    # + return - Schema replaced 
    resource function put schema/event/[string event_type](@http:Payload SchemaDefinition[] payload) returns http:Ok {
        return http:OK;
    }

    # Replace entire profile schema
    #
    # + return - Schema replaced successfully 
    resource function put schema/profile(@http:Payload SchemaDefinition[] payload) returns http:Ok {
        return http:OK;
    }
}

public type PageEvent_engagement record {
    # Custom engagement score
    decimal engagement_score?;
    # List of interactive elements clicked
    string[] interactive_elements?;
};

public type SchemaDefinition record {
    # JSON path (e.g., identity.email)
    string attribute;
    "string"|"number"|"boolean"|"array"|"object" 'type;
    "unify"|"combine"|"ignore" merge_strategy = "unify";
    boolean masking = false;
    "hash"|"redacted"|"partial"|"none" masking_strategy = "none";
};

public type ProfileDataWithoutAppContext record {
    string originCountry;
    string[] user_ids?;
    IdentityData identity?;
    PersonalityData personality?;
};

public type PageEvent_page record {
    # Full URL of the page
    string url?;
    # Page path without domain
    string path?;
    # URL of the previous page
    string referrer?;
    # Title of the page
    string title?;
    # Query parameters from the URL
    string search?;
    # Logical category of the page
    string page_category?;
    # Type of page (e.g., landing_page, blog)
    string page_type?;
    # Type of content (e.g., article, video)
    string content_type?;
    # Percentage of page scrolled
    string scroll_depth?;
    # Time spent on the page in seconds
    int time_on_page?;
    # Identifier for the previous page
    string previous_page?;
};

public type AppContext_regions_accessed record {
    string country?;
    string city?;
    string timezone?;
    string first_accessed?;
    string last_accessed?;
};

public type AppContext record {
    string app_id?;
    string subscription_plan?;
    string[] app_permissions?;
    AppContext_feature_flags feature_flags?;
    string last_active_app?;
    AppContext_usage_metrics usage_metrics?;
    AppContext_devices[] devices?;
    AppContext_regions_accessed[] regions_accessed?;
};

# Event properties for tracking user interactions
public type TrackEvent record {
    # The specific action performed (click, scroll)
    string action?;
    # The type of object interacted with (button, product)
    string object_type?;
    # Unique identifier for the interacted object
    string object_id?;
    # Human-readable name of the object
    string object_name?;
    # A numeric value associated with the event
    decimal value?;
    # Additional label for categorization
    string label?;
    # Source of the interaction (website, mobile app)
    string 'source?;
    # URL where the event occurred
    string url?;
    # URL of the referring page
    string referrer?;
};

public type AppContext_feature_flags record {
    boolean beta_features_enabled?;
    boolean dark_mode?;
};

public type ProfileData record {
    string originCountry;
    string[] user_ids?;
    IdentityData identity?;
    PersonalityData personality?;
    AppContext[] app_context?;
};

public type PersonalityData_shopping_preferences record {
    string[] favorite_brands?;
    string discount_preference?;
};

public type AppContext_devices record {
    string device_id?;
    string device_type?;
    string os?;
    string browser?;
    string browser_version?;
    string ip?;
    string last_used?;
};

public type PageEvent_utm record {
    # Traffic source (e.g., google, facebook)
    string 'source?;
    # Marketing medium (e.g., email, social)
    string medium?;
    # Campaign name
    string campaign?;
};

# Event properties for user identity tracking
public type IdentifyEvent record {
    # Unique identifier for the user
    string user_id?;
    # Custom user attributes
    record {} traits?;
};

public type Consent record {
    string[] collect;
    string[] share;
};

public type Alias record {
    string previous_perma_id;
};

# Event properties for page interactions
public type PageEvent record {
    PageEvent_page page?;
    PageEvent_utm utm?;
    PageEvent_engagement engagement?;
};

public type PersonalityData_communication_preferences record {
    boolean email_notifications?;
    boolean sms_notifications?;
    boolean push_notifications?;
};

public type Event record {
    # Type of the event
    "Identify"|"Track"|"Page" event_type;
    # Name of the event
    string event_name;
    # Unique ID for the event
    string event_id;
    # Unique ID for the application
    string app_id;
    # Time at which the event occurred
    string event_timestamp;
    # Device and session information
    record {} context;
    # User's language and regional settings
    string locale?;
    PageEvent|TrackEvent|IdentifyEvent properties?;
};

public type IdentityData record {
    string username?;
    string email?;
    string[] phone_numbers?;
    string first_name?;
    string last_name?;
    string display_name?;
    string preferred_username?;
    string profile_url?;
    string picture?;
    string[] roles?;
    string[] groups?;
    "active"|"inactive"|"suspended" account_status?;
    string created_at?;
    string updated_at?;
    string idp_provider?;
    boolean mfa_enabled?;
    string last_login?;
    string locale?;
    string timezone?;
};

public type inline_response_200 ProfileDataWithoutAppContext|ProfileData;

public type AppContext_usage_metrics record {
    int daily_active_time?;
    int monthly_logins?;
};

public type PersonalityData record {
    string[] interests?;
    string preferred_language?;
    PersonalityData_communication_preferences communication_preferences?;
    PersonalityData_shopping_preferences shopping_preferences?;
};
