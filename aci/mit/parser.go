package mit

import (
	"fmt"
	"strings"

	"lib/json"

	"github.com/tidwall/gjson"
)

// buildDN derives the DN using an RN template and record attributes
func buildDN(record gjson.Result, parentDn []string, rnTemplate string) []string {
	// If record already has a DN just return it
	dn := record.Get("dn").Str
	if dn != "" {
		return strings.Split(dn, "/")
	}

	// String templating state machine
	type state struct {
		inVariable  bool
		isBracketed bool
		varName     string
	}
	var (
		s  state
		rn string
	)

	// Iterate through characters and build the RN
	for _, c := range rnTemplate {
		switch {
		case c == '{': // start of a variable
			s.inVariable = true
		case c == '[' || c == ']':
			s.isBracketed = true
		case c == '}': // end of a variable
			value := record.Get(s.varName).Str
			if s.isBracketed {
				value = "[" + value + "]"
			}
			rn += value
			// Reset variable state
			s = state{}
		case s.inVariable:
			s.varName += string(c)
		default:
			rn += string(c)
		}
	}

	return append(parentDn, rn)
}

// setMeta creates all records in the db for a meta record
// e.g. rsp-subtree=full
func (db *DB) setMeta(root gjson.Result) error {
	type mo struct {
		object   gjson.Result
		parentDn []string
		class    string
	}
	// Create stack and populate root node
	stack := []mo{{object: root}}

	// Read RN template map
	rnTemplates := gjson.Parse(rnTemplateData).Map()

	for len(stack) > 0 {
		// Pop item off stack j
		o := stack[len(stack)-1]
		stack = stack[:len(stack)-1]

		// Get class/body for the current object
		var moBody gjson.Result
		o.object.ForEach(func(key, value gjson.Result) bool {
			o.class = key.String()
			moBody = value
			return false
		})

		// If the DN exists, use what's there
		dn := moBody.Get("attributes.dn").Str
		var thisDn []string
		if dn != "" {
			thisDn = strings.Split(dn, "/")
		} else {
			// Get the RN template from the lookup table
			rnTemplate, ok := rnTemplates[o.class]
			if !ok {
				continue
			}
			// Get DN of current object
			thisDn = buildDN(moBody.Get("attributes"), o.parentDn, rnTemplate.Str)
			dn = strings.Join(thisDn, "/")
		}

		key := fmt.Sprintf("%s:%s", o.class, dn)
		body := json.Set(moBody.Get("attributes").Raw, "dn", dn)
		if err := db.Set(key, body); err != nil {
			return err
		}

		// Add any children of this MO to stack
		for _, child := range moBody.Get("children").Array() {
			stack = append(stack, mo{object: child, parentDn: thisDn})
		}
	}
	return nil
}
