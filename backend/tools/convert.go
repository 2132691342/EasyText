package tools

import (
	"encoding/json"
	"encoding/xml"
	"fmt"
	"strings"

	"github.com/BurntSushi/toml"
	yaml "gopkg.in/yaml.v3"
)

// ConvertTool provides format conversion between JSON, YAML, TOML, and XML.
type ConvertTool struct{}

// NewConvertTool creates a new ConvertTool.
func NewConvertTool() *ConvertTool {
	return &ConvertTool{}
}

// Convert converts content from fromFmt to toFmt.
// Supported format values: "json", "yaml", "toml", "xml".
func (ct *ConvertTool) Convert(content, fromFmt, toFmt string) (string, error) {
	fromFmt = strings.ToLower(strings.TrimSpace(fromFmt))
	toFmt = strings.ToLower(strings.TrimSpace(toFmt))

	if fromFmt == toFmt {
		return content, nil
	}

	// Step 1: decode to intermediate representation
	var intermediate interface{}
	var err error

	switch fromFmt {
	case "json":
		intermediate, err = decodeJSON(content)
	case "yaml":
		intermediate, err = decodeYAML(content)
	case "toml":
		intermediate, err = decodeTOML(content)
	case "xml":
		intermediate, err = decodeXML(content)
	default:
		return "", fmt.Errorf("unsupported source format: %q", fromFmt)
	}
	if err != nil {
		return "", fmt.Errorf("parse %s: %w", fromFmt, err)
	}

	// Step 2: encode to target format
	switch toFmt {
	case "json":
		return encodeJSON(intermediate)
	case "yaml":
		return encodeYAML(intermediate)
	case "toml":
		return encodeTOML(intermediate)
	case "xml":
		return encodeXML(intermediate)
	default:
		return "", fmt.Errorf("unsupported target format: %q", toFmt)
	}
}

// --- convenience methods --------------------------------------------------

func (ct *ConvertTool) JSONToYAML(content string) (string, error) {
	return ct.Convert(content, "json", "yaml")
}
func (ct *ConvertTool) JSONToTOML(content string) (string, error) {
	return ct.Convert(content, "json", "toml")
}
func (ct *ConvertTool) JSONToXML(content string) (string, error) {
	return ct.Convert(content, "json", "xml")
}
func (ct *ConvertTool) YAMLToJSON(content string) (string, error) {
	return ct.Convert(content, "yaml", "json")
}
func (ct *ConvertTool) YAMLToTOML(content string) (string, error) {
	return ct.Convert(content, "yaml", "toml")
}
func (ct *ConvertTool) YAMLToXML(content string) (string, error) {
	return ct.Convert(content, "yaml", "xml")
}
func (ct *ConvertTool) TOMLToJSON(content string) (string, error) {
	return ct.Convert(content, "toml", "json")
}
func (ct *ConvertTool) TOMLToYAML(content string) (string, error) {
	return ct.Convert(content, "toml", "yaml")
}
func (ct *ConvertTool) TOMLToXML(content string) (string, error) {
	return ct.Convert(content, "toml", "xml")
}
func (ct *ConvertTool) XMLToJSON(content string) (string, error) {
	return ct.Convert(content, "xml", "json")
}
func (ct *ConvertTool) XMLToYAML(content string) (string, error) {
	return ct.Convert(content, "xml", "yaml")
}
func (ct *ConvertTool) XMLToTOML(content string) (string, error) {
	return ct.Convert(content, "xml", "toml")
}

// ==========================================================================
// Decoders
// ==========================================================================

func decodeJSON(content string) (interface{}, error) {
	var v interface{}
	if err := json.Unmarshal([]byte(content), &v); err != nil {
		return nil, err
	}
	return v, nil
}

func decodeYAML(content string) (interface{}, error) {
	var v interface{}
	if err := yaml.Unmarshal([]byte(content), &v); err != nil {
		return nil, err
	}
	// yaml.v3 decodes maps as map[string]interface{} already, but nested maps
	// may come out as map[interface{}]interface{} — normalise them.
	return normaliseYAMLValue(v), nil
}

func decodeTOML(content string) (interface{}, error) {
	var v interface{}
	if _, err := toml.Decode(content, &v); err != nil {
		return nil, err
	}
	return v, nil
}

// decodeXML parses XML into a nested map[string]interface{}.
// The outermost element becomes the root key.
func decodeXML(content string) (interface{}, error) {
	dec := xml.NewDecoder(strings.NewReader(content))
	// Read the first start element
	var root xmlToken
	for {
		tok, err := dec.Token()
		if err != nil {
			return nil, fmt.Errorf("empty or invalid XML")
		}
		if se, ok := tok.(xml.StartElement); ok {
			root.Name = se.Name.Local
			if err := parseXMLElement(dec, &root); err != nil {
				return nil, err
			}
			break
		}
	}
	return xmlTokenToMap(root), nil
}

// ==========================================================================
// Encoders
// ==========================================================================

func encodeJSON(v interface{}) (string, error) {
	b, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return "", err
	}
	return string(b), nil
}

func encodeYAML(v interface{}) (string, error) {
	var sb strings.Builder
	enc := yaml.NewEncoder(&sb)
	enc.SetIndent(2)
	if err := enc.Encode(v); err != nil {
		return "", err
	}
	return sb.String(), nil
}

func encodeTOML(v interface{}) (string, error) {
	// BurntSushi/toml requires a map or struct at the top level.
	// Convert interface{} to map[string]interface{} when possible.
	m, err := toMap(v)
	if err != nil {
		return "", fmt.Errorf("TOML requires a top-level object: %w", err)
	}
	var sb strings.Builder
	if err := toml.NewEncoder(&sb).Encode(m); err != nil {
		return "", err
	}
	return sb.String(), nil
}

func encodeXML(v interface{}) (string, error) {
	var sb strings.Builder
	sb.WriteString(xml.Header)
	m, err := toMap(v)
	if err != nil {
		// If not a map, wrap value directly
		sb.WriteString(fmt.Sprintf("<root>%v</root>\n", v))
		return sb.String(), nil
	}
	if err := writeXMLMap(&sb, "root", m, ""); err != nil {
		return "", err
	}
	return sb.String(), nil
}

// ==========================================================================
// XML helpers
// ==========================================================================

// xmlToken is an intermediate tree node for XML parsing.
type xmlToken struct {
	Name     string
	Text     string
	Children []xmlToken
}

// parseXMLElement reads children of the current element until its end tag.
func parseXMLElement(dec *xml.Decoder, node *xmlToken) error {
	for {
		tok, err := dec.Token()
		if err != nil {
			return err
		}
		switch t := tok.(type) {
		case xml.StartElement:
			child := xmlToken{Name: t.Name.Local}
			if err := parseXMLElement(dec, &child); err != nil {
				return err
			}
			node.Children = append(node.Children, child)
		case xml.CharData:
			node.Text += strings.TrimSpace(string(t))
		case xml.EndElement:
			return nil
		}
	}
}

// xmlTokenToMap converts an xmlToken tree to map[string]interface{}.
func xmlTokenToMap(node xmlToken) interface{} {
	if len(node.Children) == 0 {
		// Leaf node: return text value
		return node.Text
	}

	// Group children by name to detect arrays
	order := make([]string, 0)
	groups := make(map[string][]interface{})
	for _, child := range node.Children {
		if _, exists := groups[child.Name]; !exists {
			order = append(order, child.Name)
		}
		groups[child.Name] = append(groups[child.Name], xmlTokenToMap(child))
	}

	result := make(map[string]interface{}, len(order))
	for _, name := range order {
		vals := groups[name]
		if len(vals) == 1 {
			result[name] = vals[0]
		} else {
			result[name] = vals
		}
	}
	return result
}

// writeXMLMap serialises a map[string]interface{} as XML with the given tag.
func writeXMLMap(sb *strings.Builder, tag string, m map[string]interface{}, indent string) error {
	sb.WriteString(indent + "<" + tag + ">\n")
	for k, v := range m {
		if err := writeXMLValue(sb, k, v, indent+"  "); err != nil {
			return err
		}
	}
	sb.WriteString(indent + "</" + tag + ">\n")
	return nil
}

// writeXMLValue writes a single key/value pair as XML element(s).
func writeXMLValue(sb *strings.Builder, tag string, v interface{}, indent string) error {
	switch val := v.(type) {
	case map[string]interface{}:
		return writeXMLMap(sb, tag, val, indent)
	case []interface{}:
		for _, item := range val {
			if err := writeXMLValue(sb, tag, item, indent); err != nil {
				return err
			}
		}
	case nil:
		sb.WriteString(indent + "<" + tag + "/>\n")
	default:
		escaped, err := xmlEscape(fmt.Sprintf("%v", val))
		if err != nil {
			return err
		}
		sb.WriteString(indent + "<" + tag + ">" + escaped + "</" + tag + ">\n")
	}
	return nil
}

// xmlEscape escapes special XML characters in a string.
func xmlEscape(s string) (string, error) {
	var buf strings.Builder
	if err := xml.EscapeText(&buf, []byte(s)); err != nil {
		return "", err
	}
	return buf.String(), nil
}

// ==========================================================================
// Utility helpers
// ==========================================================================

// toMap asserts or converts v to map[string]interface{}.
func toMap(v interface{}) (map[string]interface{}, error) {
	switch m := v.(type) {
	case map[string]interface{}:
		return m, nil
	default:
		return nil, fmt.Errorf("value is %T, not an object", v)
	}
}

// normaliseYAMLValue converts map[interface{}]interface{} (produced by some
// YAML decoders) into map[string]interface{} recursively.
func normaliseYAMLValue(v interface{}) interface{} {
	switch val := v.(type) {
	case map[interface{}]interface{}:
		m := make(map[string]interface{}, len(val))
		for k, kv := range val {
			m[fmt.Sprintf("%v", k)] = normaliseYAMLValue(kv)
		}
		return m
	case map[string]interface{}:
		for k, kv := range val {
			val[k] = normaliseYAMLValue(kv)
		}
		return val
	case []interface{}:
		for i, item := range val {
			val[i] = normaliseYAMLValue(item)
		}
		return val
	default:
		return v
	}
}
