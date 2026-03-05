package engine

import (
	"strings"

	"github.com/PuerkitoBio/goquery"
)

// Parser extracts structured data from HTML documents
type Parser struct{}

// NewParser creates a new parser
func NewParser() *Parser {
	return &Parser{}
}

// ParseDocument parses an HTML document
func (p *Parser) ParseDocument(doc *goquery.Document) *ParsedData {
	data := &ParsedData{
		URL:     "",
		Title:   doc.Find("title").First().Text(),
		Links:   p.extractLinks(doc),
		Images:  p.extractImages(doc),
		Scripts: p.extractScripts(doc),
		Meta:    p.extractMeta(doc),
	}

	// Extract main content
	content := doc.Find("main, article, .content, #content").First()
	if content.Length() > 0 {
		data.Content = strings.TrimSpace(content.Text())
	} else {
		data.Content = strings.TrimSpace(doc.Find("body").Text())
	}

	return data
}

// ParsedData contains parsed HTML data
type ParsedData struct {
	URL     string
	Title   string
	Content string
	Links   []string
	Images  []string
	Scripts []string
	Meta    map[string]string
}

func (p *Parser) extractLinks(doc *goquery.Document) []string {
	var links []string
	doc.Find("a[href]").Each(func(i int, s *goquery.Selection) {
		if href, exists := s.Attr("href"); exists && isValidURL(href) {
			links = append(links, href)
		}
	})
	return links
}

func (p *Parser) extractImages(doc *goquery.Document) []string {
	var images []string
	doc.Find("img[src]").Each(func(i int, s *goquery.Selection) {
		if src, exists := s.Attr("src"); exists {
			images = append(images, src)
		}
	})
	return images
}

func (p *Parser) extractScripts(doc *goquery.Document) []string {
	var scripts []string
	doc.Find("script[src]").Each(func(i int, s *goquery.Selection) {
		if src, exists := s.Attr("src"); exists {
			scripts = append(scripts, src)
		}
	})
	return scripts
}

func (p *Parser) extractMeta(doc *goquery.Document) map[string]string {
	meta := make(map[string]string)
	doc.Find("meta").Each(func(i int, s *goquery.Selection) {
		name, _ := s.Attr("name")
		property, _ := s.Attr("property")
		content, _ := s.Attr("content")

		if name != "" {
			meta[name] = content
		} else if property != "" {
			meta[property] = content
		}
	})
	return meta
}

func isValidURL(url string) bool {
	if url == "" || strings.HasPrefix(url, "#") || strings.HasPrefix(url, "javascript:") {
		return false
	}
	return true
}

// ExtractText extracts clean text content
func (p *Parser) ExtractText(doc *goquery.Document) string {
	// Remove script and style elements
	doc.Find("script, style, nav, header, footer").Remove()

	// Get text from body
	return strings.TrimSpace(doc.Find("body").Text())
}

// ExtractHeadings extracts all headings
func (p *Parser) ExtractHeadings(doc *goquery.Document) map[string][]string {
	headings := make(map[string][]string)

	doc.Find("h1, h2, h3, h4, h5, h6").Each(func(i int, s *goquery.Selection) {
		tag := goquery.NodeName(s)
		text := strings.TrimSpace(s.Text())
		if text != "" {
			headings[tag] = append(headings[tag], text)
		}
	})

	return headings
}
