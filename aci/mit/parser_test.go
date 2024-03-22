package mit

import (
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParseNDO(t *testing.T) {
	a := assert.New(t)
	db, err := NewNDO(NewFolderSource(filepath.Join("testdata", "NDO")))
	a.NoError(err)
	defer db.Close()

	a.NoError(err)
	res, err := db.Find("msc_anpEpgRels:*")
	a.NoError(err)
	a.Greater(len(res), 0)
	res1, err := db.Get("msc_anpEpgRels:5ee97b394174748e7dd74eb8")
	a.Equal(`{"_id":{"$oid":"5ee97b394174748e7dd74eb8"},"siteId":"5d784dbf10000091016d97af","dn":"uni/tn-Enterprise/ap-PubSafety","epgs":["uni/tn-Enterprise/ap-PubSafety/epg-Migration"]}`, res1.Raw)
	a.NoError(err)

	res2, err := db.Find("msc_emptyFile:*")
	a.Error(err)
	a.Nil(res2)
}

func TestParseFlat(t *testing.T) {
	a := assert.New(t)
	db, err := New(NewFolderSource(filepath.Join("testdata", "flat")))
	a.NoError(err)
	defer db.Close()

	a.NoError(err)
	res, err := db.Find("topSystem:*")
	a.NoError(err)
	a.Greater(len(res), 0)
}

func TestParseChildren(t *testing.T) {
	a := assert.New(t)
	db, err := New(NewFolderSource(filepath.Join("testdata", "children")))
	a.NoError(err)
	defer db.Close()

	a.NoError(err)
	topSystems, err := db.Find("topSystem:*")
	a.NoError(err)
	a.Equal(4, len(topSystems))
	for _, topSystem := range topSystems {
		role := topSystem.Get("role").Str
		if role != "leaf" && role != "spine" {
			continue
		}
		healthInst, err := db.FindOne("healthInst:%s/health", topSystem.Get("dn").Str)
		a.NoError(err)
		a.True(healthInst.Exists(), "healthInst not found")
	}
}
