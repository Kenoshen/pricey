package pricey

import (
	"context"
	"testing"
	"time"

	"cloud.google.com/go/firestore"
	"github.com/stretchr/testify/assert"
)

func TestFirebaseStoreImplementation(t *testing.T) {
	fb, _ := NewFirebase(context.Background(), OrgGroupExtractorConfig("orgId", "groupId"), nil)
	// if this line does not compile, then there is a problem with the implementation of the interface
	_ = New(fb, fb, nil)
}

func TestUpdateField(t *testing.T) {
	now := time.Now()
	testCases := []struct {
		name     string
		original Item
		fields   []field
		updates  []firestore.Update
		final    Item
	}{
		{
			name: "no change",
			original: Item{
				Name: "ItemName",
			},
			fields:  []field{{"Name", "ItemName"}},
			updates: nil,
			final: Item{
				Name: "ItemName",
			},
		},
		{
			name: "single simple field",
			original: Item{
				Name: "ItemName",
			},
			fields: []field{{"Name", "ItemNameUpdated"}},
			updates: []firestore.Update{
				{
					Path:  "name",
					Value: "ItemNameUpdated",
				},
				{
					Path:  "updated",
					Value: now,
				},
			},
			final: Item{
				Name:    "ItemNameUpdated",
				Updated: now,
			},
		},
		{
			name: "multiple simple field",
			original: Item{
				Name:        "ItemName",
				Description: "ItemDescription",
			},
			fields: []field{{"Name", "ItemNameUpdated"}, {"Description", "ItemDescriptionUpdated"}},
			updates: []firestore.Update{
				{
					Path:  "name",
					Value: "ItemNameUpdated",
				},
				{
					Path:  "description",
					Value: "ItemDescriptionUpdated",
				},
				{
					Path:  "updated",
					Value: now,
				},
			},
			final: Item{
				Name:        "ItemNameUpdated",
				Description: "ItemDescriptionUpdated",
				Updated:     now,
			},
		},
		{
			name: "add to single slice field",
			original: Item{
				Name:        "ItemName",
				Description: "ItemDescription",
				ParentIds:   []ID{"1", "2"},
			},
			fields: []field{{"ParentIds", []ID{"1", "2", "3"}}},
			updates: []firestore.Update{
				{
					Path:  "parentIds",
					Value: []ID{"1", "2", "3"},
				},
				{
					Path:  "updated",
					Value: now,
				},
			},
			final: Item{
				Name:        "ItemName",
				Description: "ItemDescription",
				ParentIds:   []ID{"1", "2", "3"},
				Updated:     now,
			},
		},
		{
			name: "remove from single slice field",
			original: Item{
				Name:        "ItemName",
				Description: "ItemDescription",
				ParentIds:   []ID{"1", "2"},
			},
			fields: []field{{"ParentIds", []ID{"2"}}},
			updates: []firestore.Update{
				{
					Path:  "parentIds",
					Value: []ID{"2"},
				},
				{
					Path:  "updated",
					Value: now,
				},
			},
			final: Item{
				Name:        "ItemName",
				Description: "ItemDescription",
				ParentIds:   []ID{"2"},
				Updated:     now,
			},
		},
		{
			name: "add and remove from single slice field",
			original: Item{
				Name:        "ItemName",
				Description: "ItemDescription",
				ParentIds:   []ID{"1", "2"},
			},
			fields: []field{{"ParentIds", []ID{"1", "3"}}},
			updates: []firestore.Update{
				{
					Path:  "parentIds",
					Value: []ID{"1", "3"},
				},
				{
					Path:  "updated",
					Value: now,
				},
			},
			final: Item{
				Name:        "ItemName",
				Description: "ItemDescription",
				ParentIds:   []ID{"1", "3"},
				Updated:     now,
			},
		},
		{
			name: "add and remove from multiple slice fields",
			original: Item{
				Name:        "ItemName",
				Description: "ItemDescription",
				ParentIds:   []ID{"1", "2"},
				TagIds:      []ID{"A", "B"},
			},
			fields: []field{{"ParentIds", []ID{"1", "3"}}, {"TagIds", []ID{"A", "C"}}},
			updates: []firestore.Update{
				{
					Path:  "parentIds",
					Value: []ID{"1", "3"},
				},
				{
					Path:  "tagIds",
					Value: []ID{"A", "C"},
				},
				{
					Path:  "updated",
					Value: now,
				},
			},
			final: Item{
				Name:        "ItemName",
				Description: "ItemDescription",
				ParentIds:   []ID{"1", "3"},
				TagIds:      []ID{"A", "C"},
				Updated:     now,
			},
		},
		{
			name: "add and remove from complex slice fields",
			original: Item{
				Name:        "ItemName",
				Description: "ItemDescription",
				SubItems: []SubItem{{
					SubItemID: "1",
					Quantity:  1,
				}, {
					SubItemID: "2",
					Quantity:  2,
				}},
			},
			fields: []field{{"SubItems", []SubItem{{
				SubItemID: "1",
				Quantity:  1,
			}, {
				SubItemID: "3",
				Quantity:  3,
			}}}},
			updates: []firestore.Update{
				{
					Path: "subItems",
					Value: []SubItem{{
						SubItemID: "1",
						Quantity:  1,
					}, {
						SubItemID: "3",
						Quantity:  3,
					}},
				},
				{
					Path:  "updated",
					Value: now,
				},
			},
			final: Item{
				Name:        "ItemName",
				Description: "ItemDescription",
				SubItems: []SubItem{{
					SubItemID: "1",
					Quantity:  1,
				}, {
					SubItemID: "3",
					Quantity:  3,
				}},
				Updated: now,
			},
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			updates := innerUpdate(&tc.original, now, tc.fields...)

			assert.Equal(t, tc.final, tc.original)
			assert.Equal(t, tc.updates, updates)
		})
	}
}
