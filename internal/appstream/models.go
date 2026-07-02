package appstream

import "encoding/xml"

// Catalog represents the root of the appstream.xml file
type Catalog struct {
	XMLName    xml.Name    `xml:"components"`
	Version    string      `xml:"version,attr"`
	Origin     string      `xml:"origin,attr"`
	Components []Component `xml:"component"`
}

// Component represents an individual application
type Component struct {
	Type        string     `xml:"type,attr"`
	ID          string     `xml:"id"`
	Name        string     `xml:"name"`
	Summary     string     `xml:"summary"`
	Description string     `xml:"description"`
	ProjectURL  string     `xml:"url"`
	Icons       []Icon     `xml:"icon"`
	Categories  Categories `xml:"categories"`
	Releases    Releases   `xml:"releases"`
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
