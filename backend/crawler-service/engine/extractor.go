package engine

import (
	"strings"

	"github.com/PuerkitoBio/goquery"
)

// Extractor extracts specific data from crawled pages
type Extractor struct{}

func NewExtractor() *Extractor {
	return &Extractor{}
}

// ExtractEmails extracts email addresses from HTML
func (e *Extractor) ExtractEmails(doc *goquery.Document) []string {
	var emails []string
	emailPattern := `[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}`

	doc.Find("body").Each(func(i int, s *goquery.Selection) {
		text := s.Text()
		// Simple regex-based extraction would go here
		// For now, we'll just check mailto links
	})

	doc.Find("a[href]").Each(func(i int, s *goquery.Selection) {
		href, _ := s.Attr("href")
		if strings.HasPrefix(href, "mailto:") {
			email := strings.TrimPrefix(href, "mailto:")
			if idx := strings.Index(email, "?"); idx > 0 {
				email = email[:idx]
			}
			emails = append(emails, email)
		}
	})

	// Also extract from text content
	content := doc.Text()
	// Would use regex here: regexp.FindAllString(emailPattern, content, -1)

	return emails
}

// ExtractForms extracts form information from HTML
func (e *Extractor) ExtractForms(doc *goquery.Document) []FormInfo {
	var forms []FormInfo

	doc.Find("form").Each(func(i int, s *goquery.Selection) {
		form := FormInfo{
			Method: "GET",
			Action: "",
		}

		if method, exists := s.Attr("method"); exists {
			form.Method = strings.ToUpper(method)
		}
		if action, exists := s.Attr("action"); exists {
			form.Action = action
		}

		s.Find("input").Each(func(j int, input *goquery.Selection) {
			name, _ := input.Attr("name")
			inputType, _ := input.Attr("type")
			form.Fields = append(form.Fields, FieldInfo{
				Name:     name,
				Type:     inputType,
				Required: input.AttrOr("required", "") != "",
			})
		})

		forms = append(forms, form)
	})

	return forms
}

// ExtractLinks extracts all links from HTML
func (e *Extractor) ExtractLinks(doc *goquery.Document) []string {
	var links []string

	doc.Find("a[href]").Each(func(i int, s *goquery.Selection) {
		if href, exists := s.Attr("href"); exists && href != "" {
			links = append(links, href)
		}
	})

	return links
}

// ExtractImages extracts image URLs from HTML
func (e *Extractor) ExtractImages(doc *goquery.Document) []ImageInfo {
	var images []ImageInfo

	doc.Find("img").Each(func(i int, s *goquery.Selection) {
		src, exists := s.Attr("src")
		if !exists || src == "" {
			return
		}

		img := ImageInfo{Src: src}

		if alt, exists := s.Attr("alt"); exists {
			img.Alt = alt
		}
		if width, exists := s.Attr("width"); exists {
			img.Width = width
		}
		if height, exists := s.Attr("height"); exists {
			img.Height = height
		}

		images = append(images, img)
	})

	return images
}

// ExtractMetadata extracts meta tags from HTML
func (e *Extractor) ExtractMetadata(doc *goquery.Document) map[string]string {
	metadata := make(map[string]string)

	doc.Find("meta").Each(func(i int, s *goquery.Selection) {
		name, _ := s.Attr("name")
		property, _ := s.Attr("property")
		content, _ := s.Attr("content")

		if name != "" {
			metadata[name] = content
		} else if property != "" {
			metadata[property] = content
		}
	})

	return metadata
}

// ExtractScripts extracts script source URLs
func (e *Extractor) ExtractScripts(doc *goquery.Document) []string {
	var scripts []string

	doc.Find("script[src]").Each(func(i int, s *goquery.Selection) {
		if src, exists := s.Attr("src"); exists {
			scripts = append(scripts, src)
		}
	})

	return scripts
}

// ExtractStylesheets extracts CSS stylesheet URLs
func (e *Extractor) ExtractStylesheets(doc *goquery.Document) []string {
	var styles []string

	doc.Find("link[rel='stylesheet']").Each(func(i int, s *goquery.Selection) {
		if href, exists := s.Attr("href"); exists {
			styles = append(styles, href)
		}
	})

	return styles
}

type FormInfo struct {
	Method string
	Action string
	Fields []FieldInfo
}

type FieldInfo struct {
	Name     string
	Type     string
	Required bool
}

type ImageInfo struct {
	Src    string
	Alt    string
	Width  string
	Height string
}
