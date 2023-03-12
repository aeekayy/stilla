package models

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"go.mongodb.org/mongo-driver/bson"
)

func TestIngestPositive(t *testing.T) {
	var cr ConfigResponse

	bmap := bson.M{
		"_id": "638b0a89e7693d00937122ef",
		"config": bson.M{
			"checksum": "666f614e722f616d5672462f746e436c61565837624863434358386e55593966427a572f32706d6e6f706f3d",
			"config": bson.M{
				"url": "https://backstage.aeekay.co",
			},
		},
		"config_name":    "backstage",
		"config_version": "638b0a88e7693d00937122ee",
		"created":        "2022-12-03T08:36:24.972Z",
		"host":           "3112593a-b61e-4835-ae58-c44d3eebca5a",
		"created_by":     "aeekayy",
		"config_id":      "8b9a54ea-d931-43d9-8f6a-84065964208f",
		"version":        3,
		"modified":       "2022-12-03T08:36:24.972Z",
	}

	err := cr.Ingest(&bmap)

	if err != nil {
		t.Errorf("error ingesting the map: %s", err)
	}

	assert.Equal(t, cr.ConfigName, "backstage", "the two config names should be the same.")
	assert.Equal(t, cr.CreatedBy, "aeekayy", "the two authors should be the same.")
}

func TestIngestNil(t *testing.T) {
	var cr ConfigResponse

	err := cr.Ingest(nil)

	if err == nil {
		t.Error("there should be an error ingesting the map")
	}
}
