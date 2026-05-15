package lob

// USVerification is the response shape from POST /v1/us_verifications. The
// `deliverability` field is the agent-friendly summary.
type USVerification struct {
	ID                     string       `json:"id"`
	Recipient              string       `json:"recipient,omitempty"`
	PrimaryLine            string       `json:"primary_line,omitempty"`
	SecondaryLine          string       `json:"secondary_line,omitempty"`
	Urbanization           string       `json:"urbanization,omitempty"`
	LastLine               string       `json:"last_line,omitempty"`
	Deliverability         string       `json:"deliverability"`
	Components             USComponents `json:"components"`
	DeliverabilityAnalysis USAnalysis   `json:"deliverability_analysis"`
	LobConfidenceScore     USConfidence `json:"lob_confidence_score"`
	Object                 string       `json:"object,omitempty"`
}

// USComponents are the parsed parts of a verified US address.
type USComponents struct {
	PrimaryNumber            string  `json:"primary_number,omitempty"`
	StreetPredirection       string  `json:"street_predirection,omitempty"`
	StreetName               string  `json:"street_name,omitempty"`
	StreetSuffix             string  `json:"street_suffix,omitempty"`
	StreetPostdirection      string  `json:"street_postdirection,omitempty"`
	SecondaryDesignator      string  `json:"secondary_designator,omitempty"`
	SecondaryNumber          string  `json:"secondary_number,omitempty"`
	PMBDesignator            string  `json:"pmb_designator,omitempty"`
	PMBNumber                string  `json:"pmb_number,omitempty"`
	ExtraSecondaryDesignator string  `json:"extra_secondary_designator,omitempty"`
	ExtraSecondaryNumber     string  `json:"extra_secondary_number,omitempty"`
	City                     string  `json:"city,omitempty"`
	State                    string  `json:"state,omitempty"`
	Zip                      string  `json:"zip_code,omitempty"`
	ZipPlus4                 string  `json:"zip_code_plus_4,omitempty"`
	ZipType                  string  `json:"zip_code_type,omitempty"`
	DeliveryPointBarcode     string  `json:"delivery_point_barcode,omitempty"`
	AddressType              string  `json:"address_type,omitempty"`
	RecordType               string  `json:"record_type,omitempty"`
	DefaultBuildingAddress   bool    `json:"default_building_address,omitempty"`
	County                   string  `json:"county,omitempty"`
	CountyFIPS               string  `json:"county_fips,omitempty"`
	CarrierRoute             string  `json:"carrier_route,omitempty"`
	CarrierRouteType         string  `json:"carrier_route_type,omitempty"`
	Latitude                 float64 `json:"latitude,omitempty"`
	Longitude                float64 `json:"longitude,omitempty"`
}

// USAnalysis classifies why an address is or is not deliverable.
type USAnalysis struct {
	DPVConfirmation string   `json:"dpv_confirmation,omitempty"`
	DPVCMRA         string   `json:"dpv_cmra,omitempty"`
	DPVVacant       string   `json:"dpv_vacant,omitempty"`
	DPVActive       string   `json:"dpv_active,omitempty"`
	DPVFootnotes    []string `json:"dpv_footnotes,omitempty"`
	EWSMatch        bool     `json:"ews_match,omitempty"`
	LACSIndicator   string   `json:"lacs_indicator,omitempty"`
	LACSReturnCode  string   `json:"lacs_return_code,omitempty"`
	SuiteReturnCode string   `json:"suite_return_code,omitempty"`
}

// USConfidence is a 0-100 score with bucket label.
type USConfidence struct {
	Score float64 `json:"score"`
	Level string  `json:"level"`
}

// USVerificationCreate is the request body for POST /v1/us_verifications.
type USVerificationCreate struct {
	Recipient     string `json:"recipient,omitempty"`
	PrimaryLine   string `json:"primary_line,omitempty"`
	SecondaryLine string `json:"secondary_line,omitempty"`
	Urbanization  string `json:"urbanization,omitempty"`
	City          string `json:"city,omitempty"`
	State         string `json:"state,omitempty"`
	ZipCode       string `json:"zip_code,omitempty"`
	Address       string `json:"address,omitempty"` // single-line fallback
}

// IntlVerification is the response shape from POST /v1/intl_verifications.
type IntlVerification struct {
	ID             string         `json:"id"`
	Recipient      string         `json:"recipient,omitempty"`
	PrimaryLine    string         `json:"primary_line,omitempty"`
	SecondaryLine  string         `json:"secondary_line,omitempty"`
	LastLine       string         `json:"last_line,omitempty"`
	Country        string         `json:"country,omitempty"`
	Coverage       string         `json:"coverage,omitempty"`
	Deliverability string         `json:"deliverability"`
	Status         string         `json:"status,omitempty"`
	Components     IntlComponents `json:"components"`
	Object         string         `json:"object,omitempty"`
}

// IntlComponents are the parsed parts of a verified international address.
type IntlComponents struct {
	PrimaryNumber        string `json:"primary_number,omitempty"`
	StreetName           string `json:"street_name,omitempty"`
	City                 string `json:"city,omitempty"`
	State                string `json:"state,omitempty"`
	PostalCode           string `json:"postal_code,omitempty"`
	PostalCodeExtension  string `json:"postal_code_extension,omitempty"`
	DeliveryInstallation string `json:"delivery_installation,omitempty"`
	SubBuildingType      string `json:"sub_building_type,omitempty"`
	SubBuildingNumber    string `json:"sub_building_number,omitempty"`
	BuildingType         string `json:"building_type,omitempty"`
	BuildingNumber       string `json:"building_number,omitempty"`
	Country              string `json:"country,omitempty"`
}

// IntlVerificationCreate is the request body for POST /v1/intl_verifications.
type IntlVerificationCreate struct {
	Recipient     string `json:"recipient,omitempty"`
	PrimaryLine   string `json:"primary_line"`
	SecondaryLine string `json:"secondary_line,omitempty"`
	City          string `json:"city,omitempty"`
	State         string `json:"state,omitempty"`
	PostalCode    string `json:"postal_code,omitempty"`
	Country       string `json:"country"`
	Address       string `json:"address,omitempty"`
}

// USAutocompletion is the response from POST /v1/us_autocompletions.
type USAutocompletion struct {
	ID          string                  `json:"id"`
	Suggestions []USAutocompleteSuggest `json:"suggestions"`
	Object      string                  `json:"object,omitempty"`
}

// USAutocompleteSuggest is one suggestion in an autocompletion response.
type USAutocompleteSuggest struct {
	PrimaryLine string `json:"primary_line,omitempty"`
	City        string `json:"city,omitempty"`
	State       string `json:"state,omitempty"`
	ZipCode     string `json:"zip_code,omitempty"`
}

// USAutocompletionCreate is the request body for autocompletion lookups.
type USAutocompletionCreate struct {
	AddressPrefix string `json:"address_prefix"`
	City          string `json:"city,omitempty"`
	State         string `json:"state,omitempty"`
	ZipCode       string `json:"zip_code,omitempty"`
	GeoIPSort     bool   `json:"geo_ip_sort,omitempty"`
}

// ZipLookup is the response from GET /v1/us_zip_lookups/:zip_code.
type ZipLookup struct {
	ZipCode     string          `json:"zip_code"`
	ZipCodeType string          `json:"zip_code_type,omitempty"`
	Cities      []ZipLookupCity `json:"cities"`
	Object      string          `json:"object,omitempty"`
}

// ZipLookupCity is one city served by a ZIP code.
type ZipLookupCity struct {
	City          string `json:"city,omitempty"`
	State         string `json:"state,omitempty"`
	County        string `json:"county,omitempty"`
	CountyFIPS    string `json:"county_fips,omitempty"`
	PreferredCity string `json:"preferred_city,omitempty"`
	CityType      string `json:"city_type,omitempty"`
}
