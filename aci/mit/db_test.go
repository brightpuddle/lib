package mit

import (
	"testing"

	"github.com/brightpuddle/goaci"
	"github.com/stretchr/testify/assert"
	"github.com/tidwall/buntdb"
)

func newTestDB() DB {
	db, _ := buntdb.Open(":memory:")
	err := db.Update(func(tx *buntdb.Tx) error {
		_, _, err := tx.Set("fvTenant:uni/tn-a", goaci.Body{}.Set("name", "a").Str, nil)
		if err != nil {
			return err
		}
		_, _, err = tx.Set("fvTenant:uni/tn-b", goaci.Body{}.Set("name", "b").Str, nil)
		if err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		panic(err)
	}
	return DB{db: db}
}

func TestNewFolder(t *testing.T) {
	a := assert.New(t)
	db, err := New(NewFolderSource("testdata"))
	a.NoError(err)

	// verify set/get
	j := `{"b":"c"}`
	err = db.Set("a", j)
	a.NoError(err)
	res, err := db.Get("a")
	a.NoError(err)
	a.Equal(res.Raw, j)
	a.Equal(res.Get("b").Str, "c")

	// close
	err = db.Close()
	a.NoError(err)
}

func TestNewMem(t *testing.T) {
	a := assert.New(t)
	db, err := New(NewMemSource())
	a.NoError(err)

	// Verify set/get
	j := `{"b":"c"}`
	err = db.Set("a", j)
	a.NoError(err)
	res, err := db.Get("a")
	a.NoError(err)
	a.Equal(res.Raw, j)
	a.Equal(res.Get("b").Str, "c")
}

func TestDBGet(t *testing.T) {
	a := assert.New(t)
	mit := newTestDB()
	defer mit.Close()

	// key
	res, err := mit.Get("fvTenant:uni/tn-a")
	a.Equal("a", res.Get("name").Str)
	a.NoError(err)

	// key not found
	_, err = mit.Get("fvTenant:%s", "uni/tn-c")
	a.Error(err)
}

func TestDBFind(t *testing.T) {
	a := assert.New(t)
	mit := newTestDB()
	defer mit.Close()
	res, err := mit.Find("%s:*", "fvTenant")
	a.NoError(err)
	a.Equal(2, len(res))
}

func TestDBFindOne(t *testing.T) {
	a := assert.New(t)
	mit := newTestDB()
	defer mit.Close()
	res, err := mit.FindOne("%s:*-a", "fvTenant")
	a.NoError(err)
	a.Equal("a", res.Get("name").Str)
	res, err = mit.FindOne("fvTenant:uni/tn-c")
	a.NoError(err)
	a.False(res.IsObject())
}
