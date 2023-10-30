package contracts

import "go.etcd.io/bbolt"

type CellDependencyTree interface {
	// SetDependsOn
	/**
	 * Example, for formula `cell1 = cell2 + cell3`:
	 * `cellId` depends on `dependingList`
	 * cell1 depends on cell2 and cell3
	 *  SetCellDependsOn("cell1", []string{"cell2", "cell3"})
	 * `cell5 = cell1 * cell3`
	 * SetCellDependsOn("cell5", []string{"cell1", "cell3"})
	 */
	SetDependsOn(tx *bbolt.Tx, sheetId []byte, dependantCellId string, dependingOnCellIds []string) error

	// GetDependants
	/**
	 * For formulas
	 *    - `cell1 = cell2 + cell3` => cell2 and cells3 are dependants of cell1;
	 *    - `cell5 = cell1 * cell3` => cell1 and cell3 are dependants of cell5;
	 *      recursively, `cell2` are dependants of cell5 (via `cell1`)
	 * GetDependants("cell2") should returns ["cell1", "cell5"]
	 *
	 * Internally, it is stored as B+tree. It uses prefixed keys to store data in B-tree.
	 * So it is possible to get all dependants of cellId in O(log(n)) time.
	 */
	GetDependants(tx *bbolt.Tx, sheetId []byte, dependingOnCellId string) []string
}
