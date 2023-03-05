package models

import (
	"testing"

	"go.mongodb.org/mongo-driver/bson"
	"github.com/stretchr/testify/assert"
)

func TestIngestPositive(t *testing.T) {
	var cr ConfigResponse

	bmap := bson.M{
		"_id": "638b0a89e7693d00937122ef",
		"config": bson.M{
			"checksum": bson.M{
				"Subtype": 0,
				"Data":    "foaNr/amVrF/tnClaVX7bHcCCX8nUY9fBzW/2pmnopo=",
			},
			"config": bson.M{
				"url": "https://backstage.aeekay.co",
			},
			"config_name": "backstage",
			"created":     "2022-12-03T08:36:24.972Z",
			"created_by":  "aeekayy",
		},
		"config_name":    "backstage",
		"config_version": "638b0a88e7693d00937122ee",
		"created":        "2022-12-03T08:36:24.972Z",
		"created_by":     "aeekayy",
		"modified":       "2022-12-03T08:36:24.972Z",
	}

	err := cr.Ingest(&bmap)

	if err != nil {
		t.Errorf("error ingesting the map: %s", err)
	}

	assert.Equal(t, cr.ConfigName, "backstage", "the two config names should be the same.")
	assert.Equal(t, cr.CreatedBy, "aeekayy", "the two authors should be the same.")
}

func TestIngestNegative(t *testing.T) {
	var cr ConfigResponse

	err := cr.Ingest(nil)

	if err == nil {
		t.Error("there should be an error ingesting the map")
	}
}