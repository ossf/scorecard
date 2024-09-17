package raw

import (
	"encoding/xml"
)

type PropertyGroup struct {
	XMLName           xml.Name `xml:"PropertyGroup"`
	RestoreLockedMode bool     `xml:"RestoreLockedMode"`
}

type Project struct {
	XMLName        xml.Name        `xml:"Project"`
	PropertyGroups []PropertyGroup `xml:"PropertyGroup"`
}

func isRestoreLockedModeEnabled(content []byte) (error, bool) {
	var project Project

	err := xml.Unmarshal(content, &project)
	if err != nil {
		return errInvalidCsProjFile, false
	}

	for _, propertyGroup := range project.PropertyGroups {
		if propertyGroup.RestoreLockedMode {
			return nil, true
		}
	}

	return nil, false
}
