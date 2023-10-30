package main

import (
	"github.com/stretchr/testify/assert"
	"go.etcd.io/bbolt"
	"testing"
)

type TransactionCellDependencyTreeDecorator struct {
	t  *testing.T
	db *bbolt.DB
	CellDependencyTree
}

func (tree *TransactionCellDependencyTreeDecorator) SetDependsOn(sheetId []byte, dependantCellId string, dependingOnCellIds []string) (returnErr error) {
	tx, err := tree.db.Begin(true)
	assert.NoError(tree.t, err)

	returnErr = tree.CellDependencyTree.SetDependsOn(tx, sheetId, dependantCellId, dependingOnCellIds)
	assert.NoError(tree.t, tx.Commit())
	return
}

func (tree *TransactionCellDependencyTreeDecorator) GetDependants(sheetId []byte, dependingOnCellId string) (returnList []string) {
	tx, err := tree.db.Begin(false)
	assert.NoError(tree.t, err)

	returnList = tree.CellDependencyTree.GetDependants(tx, sheetId, dependingOnCellId)
	assert.NoError(tree.t, tx.Rollback())
	return
}

func NewTransactionCellDependencyTreeDecorator(t *testing.T, db *bbolt.DB) *TransactionCellDependencyTreeDecorator {
	return &TransactionCellDependencyTreeDecorator{t, db, CellDependencyTree{}}
}

func TestCellDependencyTree_GetDependants(t *testing.T) {
	db, closeDb := _createTmpDb()
	defer closeDb()

	t.Run("single-level-deep", func(t *testing.T) {
		tree := NewTransactionCellDependencyTreeDecorator(t, db)
		sheetId := []byte(t.Name())

		err := tree.SetDependsOn(sheetId, "cell1", []string{"cell100", "cell2", "cell3"})
		assert.NoError(t, err)

		assert.Empty(t, tree.GetDependants(sheetId, "cell1"))
		assert.Empty(t, tree.GetDependants(sheetId, "cellUnknown"))

		assert.Equal(t, []string{"cell1"}, tree.GetDependants(sheetId, "cell2"))
		assert.Equal(t, []string{"cell1"}, tree.GetDependants(sheetId, "cell3"))

		err = tree.SetDependsOn(sheetId, "cell1", []string{"cell5", "cell99", "cell100"})
		assert.NoError(t, err)

		assert.Equal(t, []string{"cell1"}, tree.GetDependants(sheetId, "cell5"))
		assert.Empty(t, tree.GetDependants(sheetId, "cell2"))
		assert.Empty(t, tree.GetDependants(sheetId, "cell3"))

		err = tree.SetDependsOn(sheetId, "cell1", []string{})
		assert.NoError(t, err)

		assert.Empty(t, tree.GetDependants(sheetId, "cell1"))

		assert.Empty(t, tree.GetDependants(sheetId, "cell2"))
		assert.Empty(t, tree.GetDependants(sheetId, "cell3"))
	})

	t.Run("circular-reference", func(t *testing.T) {
		tree := NewTransactionCellDependencyTreeDecorator(t, db)
		sheetId := []byte(t.Name())

		err := tree.SetDependsOn(sheetId, "cell1", []string{"cell20", "cell21"})
		assert.NoError(t, err)

		err = tree.SetDependsOn(sheetId, "cell20", []string{"cell40", "cell41"})
		assert.NoError(t, err)

		err = tree.SetDependsOn(sheetId, "cell40", []string{"cell1"})
		assert.NoError(t, err)

		assert.Equal(t,
			[]string{"cell40", "cell20", "cell1"},
			tree.GetDependants(sheetId, "cell1"),
		)
	})

	t.Run("error-empty-bucket", func(t *testing.T) {
		//		tree := CellDependencyTree{db: db}
		tree := NewTransactionCellDependencyTreeDecorator(t, db)
		err := tree.SetDependsOn(nil, "cell1", []string{"cell2", "cell3"})
		assert.Error(t, err)

		assert.Empty(t, tree.GetDependants(nil, "cell1"))
	})

	t.Run("error-db-put", func(t *testing.T) {
		tree := NewTransactionCellDependencyTreeDecorator(t, db)
		sheetId := []byte(t.Name())
		bucketId := tree.makeBucketId(sheetId)

		err := db.Update(func(tx *bbolt.Tx) error {
			bucket, err := tx.CreateBucketIfNotExists(bucketId)
			if err != nil {
				return err
			}
			_, err = bucket.CreateBucket(tree.makeDependantKey("cell1", "cell2"))
			return err
		})
		assert.NoError(t, err)

		err = tree.SetDependsOn(sheetId, "cell1", []string{"cell2"})
		assert.Error(t, err)
	})

	t.Run("error-db-delete", func(t *testing.T) {
		tree := NewTransactionCellDependencyTreeDecorator(t, db)
		sheetId := []byte(t.Name())
		bucketId := tree.makeBucketId(sheetId)

		err := tree.SetDependsOn(sheetId, "cell1", []string{"cell3"})

		err = db.Update(func(tx *bbolt.Tx) error {
			var bucket *bbolt.Bucket
			bucket, err = tx.CreateBucketIfNotExists(bucketId)
			assert.NoError(t, err)

			_, err = bucket.CreateBucket(tree.makeDependantKey("cell1", "cell2"))
			assert.NoError(t, err)

			_ = bucket.Delete(tree.makeDependantKey("cell1", "cell3"))
			_, err = bucket.CreateBucket(tree.makeDependantKey("cell1", "cell3"))
			assert.NoError(t, err)

			return nil
		})

		err = tree.SetDependsOn(sheetId, "cell1", []string{})
		assert.Error(t, err)
	})
}
