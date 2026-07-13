package appstream

import (
	"encoding/xml"
	"fmt"
	"io"
	"strings"
)

// Catalog represents the root of the appstream.xml file
type Catalog struct {
	XMLName    xml.Name    `xml:"components"`
	Version    string      `xml:"version,attr"`
	Origin     string      `xml:"origin,attr"`
	Components []Component `xml:"component"`
}

// Component represents an individual application
type Component struct {
	Type           string          `xml:"type,attr"`
	ID             string          `xml:"id"`
	Name           LocalizedString `xml:"name"`
	Summary        LocalizedString `xml:"summary"`
	Developer      string          `xml:"developer_name"`
	Description    Description     `xml:"description"`
	URLs           []URL           `xml:"url"`
	Icons          []Icon          `xml:"icon"`
	Categories     Categories      `xml:"categories"`
	Releases       Releases        `xml:"releases"`
	Screenshots    []Screenshot    `xml:"screenshots>screenshot"`
	ProjectLicense string          `xml:"project_license"`
	ContentRating  *ContentRating  `xml:"content_rating"`
}

type LocalizedString struct {
	Value string
}

func (s *LocalizedString) UnmarshalXML(dec *xml.Decoder, start xml.StartElement) error {
	isEnglish := true
	for _, attr := range start.Attr {
		if attr.Name.Local == "lang" {
			val := strings.ToLower(attr.Value)
			if val != "" && val != "en" && !strings.HasPrefix(val, "en-") {
				isEnglish = false
			}
		}
	}

	var content string
	if err := dec.DecodeElement(&content, &start); err != nil {
		return err
	}

	// We overwrite only if the string is English or we don't have a value yet
	if isEnglish || s.Value == "" {
		s.Value = content
	}
	return nil
}

type Description struct {
	Raw string `xml:",innerxml"`
}

func (d *Description) UnmarshalXML(dec *xml.Decoder, start xml.StartElement) error {
	isEnglish := true
	for _, attr := range start.Attr {
		if attr.Name.Local == "lang" {
			val := strings.ToLower(attr.Value)
			if val != "" && val != "en" && !strings.HasPrefix(val, "en-") {
				isEnglish = false
			}
		}
	}

	// If this tag is not English, and we already have a description, skip it
	if !isEnglish && d.Raw != "" {
		return dec.Skip()
	}

	var buf strings.Builder
	depth := 0

	for {
		t, err := dec.Token()
		if err != nil {
			if err == io.EOF {
				break
			}
			return err
		}

		switch tok := t.(type) {
		case xml.StartElement:
			if depth > 0 {
				depth++
				continue
			}

			isLangAttr := false
			isEnglishOrEmpty := true
			for _, attr := range tok.Attr {
				if attr.Name.Local == "lang" {
					isLangAttr = true
					val := strings.ToLower(attr.Value)
					if val != "" && val != "en" && !strings.HasPrefix(val, "en-") {
						isEnglishOrEmpty = false
					}
				}
			}

			if isLangAttr && !isEnglishOrEmpty {
				depth = 1
				continue
			}

			buf.WriteString("<" + tok.Name.Local)
			for _, attr := range tok.Attr {
				if attr.Name.Local != "lang" {
					buf.WriteString(fmt.Sprintf(" %s=\"%s\"", attr.Name.Local, attr.Value))
				}
			}
			buf.WriteString(">")

		case xml.EndElement:
			if depth > 0 {
				depth--
				continue
			}

			if tok.Name == start.Name {
				d.Raw = buf.String()
				return nil
			}

			buf.WriteString("</" + tok.Name.Local + ">")

		case xml.CharData:
			if depth > 0 {
				continue
			}
			buf.Write(tok)
		}
	}

	d.Raw = buf.String()
	return nil
}

type URL struct {
	Type  string `xml:"type,attr"`
	Value string `xml:",chardata"`
}

type Screenshot struct {
	Type    string  `xml:"type,attr"`
	Caption string  `xml:"caption"`
	Images  []Image `xml:"image"`
}

type Image struct {
	Type   string `xml:"type,attr"`
	Width  int    `xml:"width,attr"`
	Height int    `xml:"height,attr"`
	Value  string `xml:",chardata"`
}

type Icon struct {
	Type   string `xml:"type,attr"` // e.g., "cached", "remote"
	Width  int    `xml:"width,attr"`
	Height int    `xml:"height,attr"`
	Value  string `xml:",chardata"`
}

type Categories struct {
	List []string `xml:"category"`
}

type Releases struct {
	List []Release `xml:"release"`
}

type Release struct {
	Version string `xml:"version,attr"`
	Date    string `xml:"date,attr"`
}

type ContentRating struct {
	Type       string             `xml:"type,attr"`
	Attributes []ContentAttribute `xml:"content_attribute"`
}

type ContentAttribute struct {
	ID    string `xml:"id,attr"`
	Value string `xml:",chardata"`
}

func (cr *ContentRating) GetAgeRating() string {
	if cr == nil || len(cr.Attributes) == 0 {
		return "Everyone"
	}

	maxLevel := "none"
	for _, attr := range cr.Attributes {
		val := strings.ToLower(attr.Value)
		if val == "intense" {
			maxLevel = "intense"
			break
		} else if val == "moderate" {
			maxLevel = "moderate"
		} else if val == "mild" && maxLevel == "none" {
			maxLevel = "mild"
		}
	}

	switch maxLevel {
	case "intense":
		return "18+"
	case "moderate":
		return "12+"
	case "mild":
		return "7+"
	default:
		return "Everyone"
	}
}
