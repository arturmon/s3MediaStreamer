package tree

import (
	"fmt"
	"s3MediaStreamer/app/model"
	"sort"
	"strconv"
	"strings"

	"github.com/emirpasic/gods/maps/treemap"
)

type Repository interface {
	RebalanceTreePositions(tree *treemap.Map) error
	FillTree(tree *treemap.Map, items []model.PlaylistStruct) error
	AddToTree(tree *treemap.Map, items []model.PlaylistStruct, rebalance bool) error
}

type TreeService struct{}

func NewTreeService() *TreeService {
	return &TreeService{}
}

// RebalanceTreePositions reassigns sequential positions to the nodes of a tree.
//
// This function rebalances the positions of nodes in a tree so that they are assigned
// consecutive integers starting from 1, based on their current positions in ascending order.
//
// Parameters:
//   - tree: *treemap.Map
//     A treemap.Map instance representing the tree structure.
//
// Return Values:
//   - error: Returns an error if any issue occurs during the operation (e.g., invalid data types).
//     Returns nil if the operation is successful.
func (s *TreeService) RebalanceTreePositions(tree *treemap.Map) error {
	// Create a slice to hold the nodes for sorting by position
	var nodes []*model.Node
	var keys []string

	// Collect all nodes from the tree along with their keys
	tree.Each(func(key interface{}, value interface{}) {
		node := value.(*model.Node)
		nodes = append(nodes, node)
		keys = append(keys, key.(string)) // Save the corresponding key
	})

	// Sort nodes by their current position
	sort.Slice(nodes, func(i, j int) bool {
		return nodes[i].Position < nodes[j].Position
	})

	// Create a new tree to store the rebalanced nodes with updated paths
	newTree := treemap.NewWithStringComparator()

	// Reassign sequential positions starting from 1 and update keys
	for i, node := range nodes {
		// Update the node's position
		node.Position = i + 1

		// Split the original key and update the last segment (position)
		components := strings.Split(keys[i], ".")
		if len(components) != 4 {
			return fmt.Errorf("invalid key format: %s", keys[i])
		}

		// Generate new key with the updated position
		newKey := fmt.Sprintf("%s.%s.%s.%d", components[0], components[1], components[2], node.Position)

		// Add the node to the new tree with the updated key
		newTree.Put(newKey, node)
	}

	// Replace the original tree with the new tree
	*tree = *newTree

	return nil
}

// FillTree populates a tree structure using a list of PlaylistStruct items.
//
// This function iterates through a list of PlaylistStruct items, parses each item's
// path to extract the track/playlist details, and inserts them into the given tree map.
//
// Parameters:
//   - tree: *treemap.Map
//     A treemap.Map instance representing the tree structure.
//   - items: []model.PlaylistStruct
//     A slice of PlaylistStruct items to be inserted into the tree.
//
// Return Values:
//   - error: Returns an error if any of the paths are in an invalid format or if there is
//     any issue during insertion into the tree.
func (s *TreeService) FillTree(tree *treemap.Map, items []model.PlaylistStruct) error {
	for _, item := range items {
		// Convert Ltree to string and split into components by dot (.)
		pathStr := item.Path.String
		components := strings.Split(pathStr, ".")

		// We expect the format <trackType>.<parentID>.<trackID>.<position>
		if len(components) != 4 {
			return fmt.Errorf("invalid path format: %s", pathStr)
		}

		trackType := components[1]                   // 'track' or 'playlist'
		parentID := components[0]                    // Parent playlist ID
		trackID := components[2]                     // Track or Playlist ID
		position, err := strconv.Atoi(components[3]) // Position in playlist
		if err != nil {
			return fmt.Errorf("invalid position format: %s", components[3])
		}

		// Add node to the tree
		tree.Put(pathStr, &model.Node{
			ID:       trackID,
			ParentID: parentID,
			Type:     trackType,
			Position: position,
		})
	}

	return nil
}

// AddToTree adds one or multiple items to the tree and optionally rebalances positions.
//
// This function accepts a slice of PlaylistStruct items and inserts each of them
// into the given tree structure. It can also optionally rebalance the positions of
// the nodes in the tree after insertion.
//
// Parameters:
//   - tree: *treemap.Map
//     A treemap.Map instance representing the tree structure.
//   - items: []model.PlaylistStruct
//     A slice of PlaylistStruct items to be added to the tree.
//   - rebalance: bool
//     A boolean flag that determines whether to rebalance the tree's positions after insertion.
//
// Return Values:
//   - error: Returns an error if any issue occurs during the operation (e.g., invalid data types).
//     Returns nil if the operation is successful.
func (s *TreeService) AddToTree(tree *treemap.Map, items []model.PlaylistStruct, rebalance bool) error {
	for _, item := range items {
		err := s.addItemToTree(tree, item)
		if err != nil {
			return err
		}
	}

	if rebalance {
		return s.RebalanceTreePositions(tree)
	}

	return nil
}

// addItemToTree adds a single PlaylistStruct item to the tree.
//
// This is a helper function used to add an individual item to the tree. It parses
// the path string from the PlaylistStruct, extracts the necessary node details,
// and inserts the node into the tree.
//
// Parameters:
//   - tree: *treemap.Map
//     A treemap.Map instance representing the tree structure.
//   - item: model.PlaylistStruct
//     A PlaylistStruct item to be added to the tree.
//
// Return Values:
//   - error: Returns an error if the path format is invalid or if any issue occurs during
//     insertion into the tree.
func (s *TreeService) addItemToTree(tree *treemap.Map, item model.PlaylistStruct) error {
	pathStr := item.Path.String
	components := strings.Split(pathStr, ".")

	if len(components) != 4 {
		return fmt.Errorf("invalid path format: %s", pathStr)
	}

	trackType := components[1]                   // 'track' or 'playlist'
	parentID := components[0]                    // Parent playlist ID
	trackID := components[2]                     // Track or Playlist ID
	position, err := strconv.Atoi(components[3]) // Position in playlist
	if err != nil {
		return fmt.Errorf("invalid position format: %s", components[3])
	}

	tree.Put(pathStr, &model.Node{
		ID:       trackID,
		ParentID: parentID,
		Type:     trackType,
		Position: position,
	})

	return nil
}
