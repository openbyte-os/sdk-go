package app

import (
	"encoding/json"
	"errors"

	"github.com/openbyte-os/sdk-go/translation"
)

type Vendor struct {
	ID          string           `json:"id"`
	Name        translation.Text `json:"name"`
	Description translation.Text `json:"description"`
}

func VendorFromJson(jsonBytes []byte) (*Vendor, error) {
	def := &Vendor{}
	if err := json.Unmarshal(jsonBytes, def); err != nil {
		return nil, errors.New("unable to decode vendor definition json")
	}
	return def, nil
}
