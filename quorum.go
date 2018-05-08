package main

func buildQuorum(addresses []string) map[string][]string {
	n := len(addresses)
	q := make(map[string][]string, n)
	if prebuilt, ok := prebuiltQuorums[len(addresses)]; ok {
		// Lucky we have some prebuilt quorums....
		for i, addr := range addresses {
			quorum := make([]string, len(prebuilt[0]))
			for j, index := range prebuilt[i] {
				quorum[j] = addresses[index]
			}
			q[addr] = quorum
		}
	}
	return q
}

var prebuiltQuorums map[int]map[int][]int

// spent way too much time trying to generate these programmatically.
func init() {
	prebuiltQuorums = map[int]map[int][]int{
		2: {
			0: {1},
			1: {0},
		},
		3: {
			0: {1},
			1: {2},
			2: {0},
		},
		4: {
			0: {1, 2},
			1: {2, 3},
			2: {3, 0},
			3: {0, 1},
		},
		5: {
			0: {1, 2},
			1: {2, 3},
			2: {3, 4},
			3: {4, 0},
			4: {0, 1},
		},
		7: {
			0: {1, 2},
			1: {3, 5},
			2: {4, 5},
			3: {0, 4},
			4: {1, 6},
			5: {0, 6},
			6: {2, 3},
		},
		13: {
			0:  {1, 2, 3},
			1:  {4, 7, 10},
			2:  {5, 7, 12},
			3:  {5, 9, 10},
			4:  {0, 5, 6},
			5:  {1, 8, 11},
			6:  {1, 9, 12},
			7:  {0, 8, 9},
			8:  {2, 6, 10},
			9:  {2, 4, 11},
			10: {0, 11, 12},
			11: {3, 6, 7},
			12: {3, 4, 8},
		},
		21: {
			0:  {1, 2, 3, 4},
			1:  {5, 9, 13, 17},
			2:  {6, 9, 15, 20},
			3:  {6, 12, 13, 19},
			4:  {7, 11, 16, 17},
			5:  {0, 6, 8, 16},
			6:  {1, 10, 14, 18},
			7:  {1, 11, 15, 19},
			8:  {1, 12, 16, 20},
			9:  {0, 10, 11, 12},
			10: {2, 5, 16, 19},
			11: {2, 8, 13, 18},
			12: {2, 7, 14, 17},
			13: {0, 14, 15, 16},
			14: {3, 5, 11, 20},
			15: {3, 8, 10, 17},
			16: {3, 7, 9, 18},
			17: {0, 18, 19, 20},
			18: {4, 5, 12, 15},
			19: {4, 8, 9, 14},
			20: {4, 7, 10, 13},
		},
	}
}

//func buildQuorums(addresses []string) {
//	size := int(math.Floor(math.Sqrt(float64(c.Nodes))))
//	cutoff := size * size
//	length := c.K //c.K*(c.K-1)/size - 2
//	if c.Nodes > cutoff {
//		length++
//	}
//	// create quorums until requirements are satisfied
//	for !c.complete() {
//		for node := 1; node <= c.Nodes; node++ {
//			// create quorum map with capacity 'size'
//			c.Quorums[node] = make(map[int]bool, length)
//			number := node
//
//			// add the node to the quorum
//			c.Quorums[node][node] = true
//			col, row := (node-1)%size, (node-1)/size
//
//			// add column of nodes to quorum
//			for j := 0; j < size; j++ {
//				number = j*size + col + 1
//				if number != node {
//					c.Quorums[node][number] = true
//				}
//			}
//
//			// add node if the square matrix isn't so square
//			if node <= cutoff {
//				// add row of nodes to quorum
//				for i := 1; i <= size; i++ {
//					number = row*size + i
//					if number != node {
//						c.Quorums[node][number] = true
//					}
//				}
//			}
//			// randomly add needed nodes to quorum
//			for len(c.Quorums[node]) < length {
//				if c.Nodes <= cutoff {
//					number = rand.Intn(c.Nodes-cutoff) + cutoff + 1
//				} else {
//					number = rand.Intn(c.Nodes) + 1
//				}
//				c.Quorums[node][number] = true
//			}
//		}
//	}
//}
//
//func (c Coterie) complete() bool {
//	// clear counts
//	for k := range c.Counts {
//		c.Counts[k] = 0
//	}
//
//	// roll through quorums, get counts of nodes
//	for i := 1; i <= c.Nodes; i++ {
//		// add counts
//		for node := range c.Quorums[i] {
//			c.Counts[node]++
//		}
//		for j := i + 1; j < c.Nodes; j++ {
//			// check for intersection of quorum[i] & quorum[j]
//			if !intersection(c.Quorums[i], c.Quorums[j]) {
//				// if no intersection, incomplete quorum
//				return false
//			}
//		}
//	}
//	// make sure all counts are above K
//	for _, count := range c.Counts {
//		if count < c.K {
//			return false
//		}
//	}
//	return true
//}
