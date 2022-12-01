// Copyright 2021 OpenSSF Scorecard Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package clients

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"
)

const (
	inProgressResp = "in_progress"
	passingResp    = "passing"
	silverResp     = "silver"
	goldResp       = "gold"
)

var errUnsupportedBadge = errors.New("unsupported badge")

// BadgeResponse struct is used to read/write CII Best Practices badge data.
type BadgeResponse struct {
	BadgeLevel string `json:"badge_level"`
}

// getBadgeLevel parses a string badge value into BadgeLevel enum.
func (resp BadgeResponse) getBadgeLevel() (BadgeLevel, error) {
	if strings.Contains(resp.BadgeLevel, inProgressResp) {
		return InProgress, nil
	}
	if strings.Contains(resp.BadgeLevel, passingResp) {
		return Passing, nil
	}
	if strings.Contains(resp.BadgeLevel, silverResp) {
		return Silver, nil
	}
	if strings.Contains(resp.BadgeLevel, goldResp) {
		return Gold, nil
	}
	return Unknown, fmt.Errorf("%w: %s", errUnsupportedBadge, resp.BadgeLevel)
}

// AsJSON outputs BadgeResponse struct in JSON format.
func (resp BadgeResponse) AsJSON() ([]byte, error) {
	ret := []BadgeResponse{resp}
	jsonData, err := json.Marshal(ret)
	if err != nil {
		return nil, fmt.Errorf("error during json.Marshal: %w", err)
	}
	return jsonData, nil
}

// ParseBadgeResponseFromJSON parses input []byte value into []BadgeResponse.
func ParseBadgeResponseFromJSON(data []byte) ([]BadgeResponse, error) {
	parsedResponse := []BadgeResponse{}
	if err := json.Unmarshal(data, &parsedResponse); err != nil {
		return nil, fmt.Errorf("error during json.Unmarshal: %w", err)
	}
	return parsedResponse, nil
}
