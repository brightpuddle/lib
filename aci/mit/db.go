// Package mit models the ACI Management Information Tree (MIT) as a key-value store.
package mit

import (
	_ "embed"
	"fmt"
	"strings"

	"lib/json"

	"github.com/tidwall/buntdb"
	"github.com/tidwall/gjson"
)

// RNs are generated with pyACI
// https://pyaci.readthedocs.io/en/latest/user/installation.html
//
//go:embed rns.json
var rnTemplateData string

// DB is a key value db for the ACI MIT
// keys are class:dn
// values are the full JSON record
type DB struct {
	db *buntdb.DB
	// rnTemplates map[string]gjson.Result
}

// var rnTemplates map[string]gjson.Result

// NewNDO creates a new DB for NDO from a temp folder path
func NewNDO(src Source) (db DB, err error) {
	d, err := buntdb.Open(":memory:")
	if err != nil {
		return
	}
	db.db = d
	entries, err := src.Entries()
	if err != nil {
		return db, err
	}

	for _, entry := range entries {
		body, err := entry.Read()
		if err != nil {
			return db, err
		}

		gjson.ForEachLine(string(body), func(j gjson.Result) bool {
			oid := j.Get("_id.$oid")
			key := fmt.Sprintf("%s:%s", entry.Class, oid.Str)
			if err = db.Set(key, j.Raw); err != nil {
				return false
			}
			return true
		})

		if err != nil {
			return db, err
		}
	}
	return db, nil
}

// New creates a new DB from a temp folder path.
func New(src Source) (db DB, err error) {
	d, err := buntdb.Open(":memory:")
	if err != nil {
		return
	}
	db.db = d

	entries, _ := src.Entries()

	for _, entry := range entries {
		body, err := entry.Read()
		if err != nil {
			return db, err
		}
		j := gjson.ParseBytes(body)

		switch imdata := j.Get("imdata"); {
		case imdata.Get("0.moCount").Exists():
			record := imdata.Get("0.moCount.attributes")
			key := fmt.Sprintf("%s:%s", entry.Class, record.Get("dn").Str)
			if err := db.Set(key, record.Raw); err != nil {
				return db, err
			}
		case imdata.Exists() && imdata.IsArray():
			for _, mo := range imdata.Array() {
				for class, record := range mo.Map() {
					attrs := record.Get("attributes")
					dn := attrs.Get("dn").Str
					if err := db.Set(class+":"+dn, attrs.Raw); err != nil {
						return db, err
					}
					children := record.Get("children")
					if children.Exists() && children.IsArray() {
						if err := db.setMeta(mo); err != nil {
							return db, err
						}
					}
				}
			}
		default:
			// Fall back to building DNs recursively - this is *much* slower
			if err := db.setMeta(j); err != nil {
				return db, err
			}
		}
	}
	return db, nil
}

// Close closes the DB.
func (db *DB) Close() error {
	return db.db.Close()
}

// Get return a value or an error
func (db *DB) Get(key string, a ...interface{}) (res gjson.Result, err error) {
	key = fmt.Sprintf(key, a...)
	if err := db.db.View(func(tx *buntdb.Tx) error {
		val, err := tx.Get(key)
		if err != nil {
			return err
		}
		res = gjson.Parse(val)
		return nil
	}); err != nil {
		return res, fmt.Errorf("DB:GET:%s:%s", key, err)
	}
	return res, nil
}

// Set sets a value by key
func (db *DB) Set(key, value string) error {
	return db.db.Update(func(tx *buntdb.Tx) error {
		if _, _, err := tx.Set(key, value, nil); err != nil {
			return fmt.Errorf("cannot set key: %v", err)
		}
		return nil
	})
}

// SetMany sets multiple values.
func (db *DB) SetMany(vals map[string]interface{}) error {
	return db.db.Update(func(tx *buntdb.Tx) error {
		for k, v := range vals {
			res := json.Marshal(v)
			if _, _, err := tx.Set(k, res, nil); err != nil {
				return fmt.Errorf("cannot set key: %v", err)
			}
		}
		return nil
	})
}

// SetRaw ingests raw JSON
func (db *DB) SetRaw(val string) error {
	return db.db.Update(func(tx *buntdb.Tx) error {
		for k, v := range gjson.Parse(val).Map() {
			if _, _, err := tx.Set(k, v.Raw, nil); err != nil {
				return fmt.Errorf("cannot set key: %v", err)
			}
		}
		return nil
	})
}

// Find searches for values by pattern.
func (db *DB) Find(pattern string, a ...interface{}) (res []gjson.Result, err error) {
	pattern = fmt.Sprintf(pattern, a...)
	pattern = strings.Replace(pattern, "//", "/", -1)
	if err := db.db.View(func(tx *buntdb.Tx) error {
		return tx.AscendKeys(pattern, func(_, v string) bool {
			res = append(res, gjson.Parse(v))
			return true
		})
	}); err != nil {
		return res, fmt.Errorf("DB:FIND:%s:%s", pattern, err)
	}
	if strings.HasSuffix(pattern, ":*") && len(res) == 0 {
		return res, fmt.Errorf("DB:FIND:%s:%s", pattern, "result is empty")
	}
	return res, nil
}

// FindOne searches for a value by pattern.
func (db *DB) FindOne(pattern string, a ...interface{}) (res gjson.Result, err error) {
	pattern = fmt.Sprintf(pattern, a...)
	pattern = strings.Replace(pattern, "//", "/", -1)
	if err := db.db.View(func(tx *buntdb.Tx) error {
		return tx.AscendKeys(pattern, func(_, v string) bool {
			res = gjson.Parse(v)
			return false
		})
	}); err != nil {
		return res, fmt.Errorf("DB:FIND_ONE:%s:%s", pattern, err)
	}
	return res, nil
}
