package app

import (
	"errors"
	"regexp"
	"strings"
)

var (
	// ErrInvalidGlobalAppID invalid Global App ID
	ErrInvalidGlobalAppID = errors.New("the Global App ID specified is invalid")
	// ErrInvalidID Error for an invalid ID
	ErrInvalidID = errors.New("the ID specified is invalid")
)

type GlobalAppID struct {
	VendorID string `json:"vendorID,omitempty"`
	AppID    string `json:"appID,omitempty"`
	tertiary string // Left over data from the split
}

func (g GlobalAppID) String() string            { return g.VendorID + "/" + g.AppID }
func (g *GlobalAppID) SetTertiary(value string) { g.tertiary = value }
func (g *GlobalAppID) ClearTertiary() string {
	value := g.tertiary
	g.tertiary = ""
	return value
}

func (g GlobalAppID) Tertiary() string { return g.tertiary }
func (g GlobalAppID) AsPath() string   { return g.String() + "/" + g.tertiary }

// Validate a Global App ID, strict mode will ensure tertiary data is empty
func (g GlobalAppID) Validate(strict bool) error {

	if err := ValidateID(g.VendorID); err != nil {
		return errors.New("Invalid Vendor ID " + g.VendorID)
	}

	if err := ValidateID(g.AppID); err != nil {
		return errors.New("Invalid App ID " + g.AppID)
	}
	if strict && g.tertiary != "" {
		return ErrInvalidGlobalAppID
	}
	return nil

}

func (g GlobalAppID) Matches(against GlobalAppID, strict bool) bool {
	return against.AppID == g.AppID && against.VendorID == g.VendorID && (!strict || g.tertiary == against.tertiary)
}

// NewID New Create a Global App ID from your vendor and application IDs
func NewID(vendorID string, applicationID string) GlobalAppID {
	resp := GlobalAppID{VendorID: vendorID, AppID: applicationID}
	return resp
}

// IDFromString FromString Take a string starting with a GlobalAppID and extract the vendor, app and tertiary
func IDFromString(input string) GlobalAppID {
	gid := GlobalAppID{}
	parts := strings.SplitN(input, "/", 3)
	if len(parts) > 1 {
		gid = NewID(parts[0], parts[1])
		if len(parts) > 2 {
			gid.tertiary = parts[2]
		}
	}
	return gid
}

var createIDRegex = regexp.MustCompile("[^a-zA-Z0-9]+")

// CreateID converts a string to a valid ID
func CreateID(input string) string {
	output := createIDRegex.ReplaceAllString(input, "-")
	output = strings.ToLower(output)
	output = strings.Trim(output, "-")
	return output
}

var validateIDRegex = regexp.MustCompile("^[a-z0-9][a-z0-9\\-]+[a-z0-9]$")

// ValidateID verifies an ID is a valid string
func ValidateID(input string) error {
	if !validateIDRegex.MatchString(input) {
		return ErrInvalidID
	}
	return nil
}

type ScopedKey struct {
	GlobalAppID
	Key string
}

func (k ScopedKey) String() string { return k.VendorID + "/" + k.AppID + "/" + k.Key }

func ScopedKeyFromString(input string) ScopedKey {
	i := IDFromString(input)
	return ScopedKey{Key: i.ClearTertiary(), GlobalAppID: i}
}

func NewScopedKey(key string, gaid *GlobalAppID) ScopedKey {
	if gaid == nil {
		return ScopedKey{Key: key}
	}
	return ScopedKey{Key: key, GlobalAppID: *gaid}
}
