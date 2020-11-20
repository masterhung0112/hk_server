package model

type SearchParams struct {
	Terms                  string
	ExcludedTerms          string
	IsHashtag              bool
	InChannels             []string
	ExcludedChannels       []string
	FromUsers              []string
	ExcludedUsers          []string
	AfterDate              string
	ExcludedAfterDate      string
	BeforeDate             string
	ExcludedBeforeDate     string
	OnDate                 string
	ExcludedDate           string
	OrTerms                bool
	IncludeDeletedChannels bool
	TimeZoneOffset         int
	// True if this search doesn't originate from a "current user".
	SearchWithoutUserId bool
}
