package main

import (
	"devChallengeExcel/contracts"
	"devChallengeExcel/mocks"
	"errors"
	"fmt"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.etcd.io/bbolt"
	"os"
	"strings"
	"testing"
)

func TestSheet_SetCell(t *testing.T) {
	expectedCellsMatcher := func(expectedCells ...contracts.Cell) interface{} {
		return mock.MatchedBy(func(cells []*contracts.Cell) bool {
			if len(cells) != len(expectedCells) {
				fmt.Fprintf(os.Stderr, "len(cells): %d, len(expectedCells): %d\n", len(cells), len(expectedCells))
				return false
			}

			for i, cell := range cells {
				if cell == nil {
					fmt.Fprintf(os.Stderr, "cell[%d] is nil\n", i)
					return false
				}

				if cell.CanonicalKey != expectedCells[i].CanonicalKey {
					fmt.Fprintf(os.Stderr, "cell.CanonicalKey: %s, expectedCells[%d].CanonicalKey: %s\n", cell.CanonicalKey, i, expectedCells[i].CanonicalKey)
					return false
				}
				if cell.Value != expectedCells[i].Value {
					fmt.Fprintf(os.Stderr, "cell.Value: %s, expectedCells[%d].Value: %s\n", cell.Value, i, expectedCells[i].Value)
					return false
				}
				if cell.Result != expectedCells[i].Result {
					fmt.Fprintf(os.Stderr, "cell.Result: %s, expectedCells[%d].Result: %s\n", cell.Result, i, expectedCells[i].Result)
					return false
				}
			}
			return true
		})
	}
	canonicalizer := NewCanonicalizer()

	sheetId := "sheet1"

	cell1 := "cell1"
	canonical1 := canonicalizer.Canonicalize(cell1)
	value := "value"

	cell2 := "cell2"
	canonical2 := canonicalizer.Canonicalize(cell2)
	value2 := "=cell3"

	cell3 := "cell3"
	canonical3 := canonicalizer.Canonicalize(cell3)
	value3 := "value3"

	serializer := NewCellBinarySerializer()

	t.Run("success", func(t *testing.T) {
		db, dbClose := _createTmpDb()
		defer dbClose()

		t.Run("first_write", func(t *testing.T) {
			executor := mocks.NewExpressionExecutor(t)
			webhookDispatcher := mocks.NewWebhookDispatcher(t)

			sheetRepository := &SheetRepository{
				db:                db,
				executor:          executor,
				canonicalizer:     canonicalizer,
				serializer:        serializer,
				dependencyTree:    &CellDependencyTree{},
				webhookDispatcher: webhookDispatcher,
			}

			executor.On("MultiEvaluate", contracts.ExpressionsMap{canonical1: &value}, mock.Anything, true).
				Return(func(expressions contracts.ExpressionsMap, getter contracts.CellValuesGetter, breakOnError bool) error {
					*expressions[canonical1] = "result"
					return nil
				})

			executor.On("ExtractDependingOnList", value).Return([]string{})

			expectCell := contracts.Cell{
				CanonicalKey: canonical1,
				Value:        value,
				Result:       "result",
			}
			webhookDispatcher.On("Notify", sheetId, expectedCellsMatcher(expectCell)).Return()

			cell, err, _ := sheetRepository.SetCell(sheetId, cell1, value, true)

			assert.NotNil(t, cell)
			assert.NoError(t, err)

			assert.Equal(t, "value", cell.Value)
			assert.Equal(t, "result", cell.Result)
		})

		t.Run("repeat_write", func(t *testing.T) {
			executor := mocks.NewExpressionExecutor(t)
			sheetRepository := &SheetRepository{
				db:             db,
				executor:       executor,
				canonicalizer:  canonicalizer,
				serializer:     serializer,
				dependencyTree: &CellDependencyTree{},
			}

			executor.On("Evaluate", value, mock.Anything).Return("result", nil)

			cell, err, _ := sheetRepository.SetCell(sheetId, cell1, value, true)

			assert.NotNil(t, cell)
			assert.NoError(t, err)

			assert.Equal(t, "value", cell.Value)
			assert.Equal(t, "result", cell.Result)
		})
	})

	t.Run("success_with_execute_dependants", func(t *testing.T) {
		db, dbClose := _createTmpDb()
		defer dbClose()
		executor := mocks.NewExpressionExecutor(t)
		webhookDispatcher := mocks.NewWebhookDispatcher(t)

		sheetRepository := &SheetRepository{
			db:                db,
			executor:          executor,
			canonicalizer:     canonicalizer,
			serializer:        serializer,
			dependencyTree:    &CellDependencyTree{},
			webhookDispatcher: webhookDispatcher,
		}

		executor.On("MultiEvaluate", contracts.ExpressionsMap{canonical2: &value2}, mock.Anything, true).
			Return(func(expressions contracts.ExpressionsMap, getter contracts.CellValuesGetter, breakOnError bool) error {
				*expressions[canonical2] = "result2"
				return nil
			})
		executor.On("ExtractDependingOnList", value2).Return([]string{canonical3})

		webhookDispatcher.On("Notify", sheetId, expectedCellsMatcher(contracts.Cell{
			CanonicalKey: canonical2,
			Value:        value2,
			Result:       "result2",
		})).Return().Once()

		cell, err, _ := sheetRepository.SetCell(sheetId, cell2, value2, true)
		assert.NotNil(t, cell)
		assert.NoError(t, err)

		expectedMaps := contracts.ExpressionsMap{
			canonical2: &value2,
			canonical3: &value3,
		}

		executor.On("MultiEvaluate", expectedMaps, mock.Anything, true).
			Return(func(expressions contracts.ExpressionsMap, getter contracts.CellValuesGetter, breakOnError bool) error {
				assert.Equal(t, []*string{nil}, getter([]string{"test"}))

				for key, expressionValue := range expressions {
					*expressionValue = key + "_result"
				}
				return nil
			})
		executor.On("ExtractDependingOnList", value3).Return([]string{""})

		webhookDispatcher = mocks.NewWebhookDispatcher(t)
		sheetRepository.webhookDispatcher = webhookDispatcher
		expectedCell2 := contracts.Cell{
			CanonicalKey: canonical2,
			Value:        value2,
			Result:       "cell2_result",
		}

		expectedCell3 := contracts.Cell{
			CanonicalKey: canonical3,
			Value:        value3,
			Result:       "cell3_result",
		}

		webhookDispatcher.On("Notify", sheetId, expectedCellsMatcher(expectedCell3, expectedCell2)).Return().Once()

		cell, err, _ = sheetRepository.SetCell(sheetId, cell3, value3, true)

		assert.NotNil(t, cell)
		assert.NoError(t, err)

		assert.Equal(t, value3, cell.Value)
		assert.Equal(t, canonical3+"_result", cell.Result)
	})

	t.Run("execute_error", func(t *testing.T) {
		isolatedDb, closeIsolatedDB := _createTmpDb()
		defer closeIsolatedDB()

		executor := mocks.NewExpressionExecutor(t)
		sheet := &SheetRepository{
			db:             isolatedDb,
			executor:       executor,
			canonicalizer:  canonicalizer,
			serializer:     serializer,
			dependencyTree: &CellDependencyTree{},
		}

		executor.On("MultiEvaluate", contracts.ExpressionsMap{canonical1: &value}, mock.Anything, true).
			Return(errors.New("execute-error"))

		executor.On("ExtractDependingOnList", value).Return([]string{}).Maybe()

		cell, err, _ := sheet.SetCell(sheetId, cell1, value, true)

		assert.NotNil(t, cell)
		assert.Error(t, err)

		assert.Equal(t, value, cell.Value)
		assert.Equal(t, value, cell.Result)
	})

	t.Run("set_depends_on_error", func(t *testing.T) {
		isolatedDb, closeIsolatedDB := _createTmpDb()
		defer closeIsolatedDB()

		expectedErr := errors.New("set-depends-on-error")

		executor := mocks.NewExpressionExecutor(t)
		executor.On("MultiEvaluate", contracts.ExpressionsMap{canonical1: &value}, mock.Anything, true).
			Return(func(expressions contracts.ExpressionsMap, getter contracts.CellValuesGetter, breakOnError bool) error {
				*expressions[canonical1] = "result"
				return nil
			})
		executor.On("ExtractDependingOnList", "value").Return([]string{}).Maybe()

		webhookDispatcher := mocks.NewWebhookDispatcher(t)
		webhookDispatcher.On("Notify", sheetId, expectedCellsMatcher(contracts.Cell{
			CanonicalKey: canonical1,
			Value:        value,
			Result:       "result",
		})).Return().Once()

		tree := mocks.NewCellDependencyTree(t)
		tree.On("SetDependsOn", mock.Anything, []byte(sheetId), canonical1, []string{}).Return(expectedErr)
		tree.On("GetDependants", mock.Anything, []byte(sheetId), canonical1).Return([]string{}).Maybe()

		sheet := &SheetRepository{
			db:                isolatedDb,
			executor:          executor,
			canonicalizer:     NewCanonicalizer(),
			serializer:        NewCellBinarySerializer(),
			dependencyTree:    tree,
			webhookDispatcher: webhookDispatcher,
		}

		cell, err, _ := sheet.SetCell(sheetId, cell1, value, true)

		assert.NotNil(t, cell)
		assert.Error(t, err)
		assert.Equal(t, expectedErr, err)

		assert.Equal(t, value, cell.Value)
		assert.Equal(t, "result", cell.Result)
	})

	t.Run("save_value_error", func(t *testing.T) {
		dbWithError, closeWithError := _createTmpDb()
		defer closeWithError()

		_ = dbWithError.Update(func(tx *bbolt.Tx) error {
			bucket, err := tx.CreateBucket([]byte(sheetId))
			assert.NoError(t, err)

			_, err = bucket.CreateBucket([]byte(canonical1))
			assert.NoError(t, err)

			return nil
		})

		executor := mocks.NewExpressionExecutor(t)
		executor.On("MultiEvaluate", contracts.ExpressionsMap{canonical1: &value}, mock.Anything, true).
			Return(func(expressions contracts.ExpressionsMap, getter contracts.CellValuesGetter, breakOnError bool) error {
				*expressions[canonical1] = "result"
				return nil
			})
		executor.On("ExtractDependingOnList", "value").Return([]string{})

		webhookDispatcher := mocks.NewWebhookDispatcher(t)
		webhookDispatcher.On("Notify", sheetId, expectedCellsMatcher(contracts.Cell{
			CanonicalKey: canonical1,
			Value:        value,
			Result:       "result",
		})).Return().Once()

		tree := mocks.NewCellDependencyTree(t)
		tree.On("SetDependsOn", mock.Anything, []byte(sheetId), canonical1, []string{}).Return(nil)
		tree.On("GetDependants", mock.Anything, []byte(sheetId), canonical1).Return([]string{})

		sheet := &SheetRepository{
			db:                dbWithError,
			executor:          executor,
			canonicalizer:     NewCanonicalizer(),
			serializer:        NewCellBinarySerializer(),
			dependencyTree:    tree,
			webhookDispatcher: webhookDispatcher,
		}

		cell, err, _ := sheet.SetCell(sheetId, cell1, value, true)

		assert.NotNil(t, cell)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "incompatible value")

		assert.Equal(t, value, cell.Value)
		assert.Equal(t, "result", cell.Result)
	})

	t.Run("blacklist_char", func(t *testing.T) {
		sheet := &SheetRepository{}
		cell, err, _ := sheet.SetCell(sheetId, "cell1+cell1", "value", true)

		assert.NotNil(t, cell)
		assert.Error(t, err)

		assert.Equal(t, "value", cell.Value)
		assert.Empty(t, cell.Result)
	})

	t.Run("fail_clear_bucket", func(t *testing.T) {
		db, dbClose := _createTmpDb()
		defer dbClose()

		executor := mocks.NewExpressionExecutor(t)
		webhookDispatcher := mocks.NewWebhookDispatcher(t)

		webhookDispatcher.On("Notify", "", expectedCellsMatcher(contracts.Cell{
			CanonicalKey: canonical1,
			Value:        value,
			Result:       "result",
		})).Return().Once()

		sheet := &SheetRepository{
			db:                db,
			executor:          executor,
			canonicalizer:     NewCanonicalizer(),
			serializer:        NewCellBinarySerializer(),
			dependencyTree:    &CellDependencyTree{},
			webhookDispatcher: webhookDispatcher,
		}

		executor.On("MultiEvaluate", contracts.ExpressionsMap{canonical1: &value}, mock.Anything, true).
			Return(func(expressions contracts.ExpressionsMap, getter contracts.CellValuesGetter, breakOnError bool) error {
				*expressions[canonical1] = "result"
				return nil
			}).Maybe()

		executor.On("ExtractDependingOnList", value).Return([]string{}).Maybe()

		cell, err, _ := sheet.SetCell("", "cell1", "value", true)

		assert.NotNil(t, cell)
		assert.Error(t, err)

		assert.EqualError(t, err, "bucket name required")
	})

}

func TestSheet_GetCell(t *testing.T) {
	sheetId := "SHeetId"
	db := _prepareSheet(t, sheetId)
	defer db.Close()

	t.Run("success", func(t *testing.T) {
		executor := mocks.NewExpressionExecutor(t)
		sheet := &SheetRepository{
			db:             db,
			executor:       executor,
			canonicalizer:  NewCanonicalizer(),
			serializer:     NewCellBinarySerializer(),
			dependencyTree: &CellDependencyTree{},
		}

		executor.On("Evaluate", "value1", mock.Anything).Return("result1", nil)
		executor.On("Evaluate", "value2", mock.Anything).Return("result2", nil)

		cell, err := sheet.GetCell(strings.ToUpper(sheetId), "cell1")

		assert.NotNil(t, cell)
		assert.NoError(t, err)

		assert.Equal(t, "value1", cell.Value)
		assert.Equal(t, "result1", cell.Result)

		executor.AssertNumberOfCalls(t, "Evaluate", 1)

		camelCaseSheetId := strings.ToLower(sheetId[0:2]) + strings.ToUpper(sheetId[2:3]) + strings.ToLower(sheetId[3:])

		cell, err = sheet.GetCell(camelCaseSheetId, "cell2")

		assert.NotNil(t, cell)
		assert.NoError(t, err)

		assert.Equal(t, "value2", cell.Value)
		assert.Equal(t, "result2", cell.Result)
	})

	t.Run("execute-error", func(t *testing.T) {
		executor := mocks.NewExpressionExecutor(t)
		sheet := &SheetRepository{
			db:             db,
			executor:       executor,
			canonicalizer:  NewCanonicalizer(),
			serializer:     NewCellBinarySerializer(),
			dependencyTree: &CellDependencyTree{},
		}

		executor.On("Evaluate", "value1", mock.Anything).Return("ERROR", errors.New("execute-error"))

		cell, err := sheet.GetCell(sheetId, "cell1")

		assert.NotNil(t, cell)
		assert.Error(t, err)
		assert.Equal(t, "execute-error", err.Error())

		assert.Equal(t, "value1", cell.Value)
		assert.Equal(t, "ERROR", cell.Result)
	})

	t.Run("sheet_not_found", func(t *testing.T) {
		sheet := &SheetRepository{
			db:             db,
			executor:       nil,
			canonicalizer:  NewCanonicalizer(),
			serializer:     NewCellBinarySerializer(),
			dependencyTree: &CellDependencyTree{},
		}

		cell, err := sheet.GetCell("not-exists", "cell1")

		assert.NotNil(t, cell)
		assert.Error(t, err)
		assert.ErrorIs(t, err, contracts.SheetNotFoundError)

		assert.Equal(t, "", cell.Value)
		assert.Equal(t, "", cell.Result)
	})

	t.Run("cell_not_found", func(t *testing.T) {
		sheet := &SheetRepository{
			db:             db,
			executor:       nil,
			canonicalizer:  NewCanonicalizer(),
			serializer:     NewCellBinarySerializer(),
			dependencyTree: &CellDependencyTree{},
		}

		cell, err := sheet.GetCell(sheetId, "cell-not-exists")

		assert.NotNil(t, cell)
		assert.Error(t, err)
		assert.ErrorIs(t, err, contracts.CellNotFoundError)

		assert.Equal(t, "", cell.Value)
		assert.Equal(t, "", cell.Result)
	})

	t.Run("wrong-data-storage", func(t *testing.T) {
		errorSheet := "error-sheet"

		errorCell := NewCanonicalizer().Canonicalize("errorCell")

		err := db.Update(func(tx *bbolt.Tx) error {
			bucket, err := tx.CreateBucketIfNotExists([]byte(errorSheet))
			assert.NoError(t, err)

			err = bucket.Put([]byte(errorCell), []byte{'C', 'U'})
			assert.NoError(t, err)

			return nil
		})
		assert.NoError(t, err)

		sheet := &SheetRepository{
			db:             db,
			canonicalizer:  NewCanonicalizer(),
			serializer:     NewCellBinarySerializer(),
			dependencyTree: &CellDependencyTree{},
		}

		cell, err := sheet.GetCell(errorSheet, errorCell)

		assert.NotNil(t, cell)
		assert.Error(t, err)
		assert.ErrorIs(t, err, SerializerError)

		assert.Equal(t, "", cell.Value)
		assert.Equal(t, "", cell.Result)
	})
}

func TestSheet_GetCellList(t *testing.T) {
	canonicalizer := NewCanonicalizer()

	sheetId := "sheetId"
	db := _prepareSheet(t, sheetId)
	defer db.Close()

	t.Run("success", func(t *testing.T) {
		executor := mocks.NewExpressionExecutor(t)
		sheet := &SheetRepository{
			db:             db,
			executor:       executor,
			canonicalizer:  NewCanonicalizer(),
			serializer:     NewCellBinarySerializer(),
			dependencyTree: &CellDependencyTree{},
		}

		expectedMaps := contracts.ExpressionsMap{
			canonicalizer.Canonicalize("cell1"): _makeStringRef("value1"),
			canonicalizer.Canonicalize("cell2"): _makeStringRef("value2"),
		}

		executor.On("MultiEvaluate", expectedMaps, mock.Anything, false).
			Return(func(expressions contracts.ExpressionsMap, getter contracts.CellValuesGetter, breakOnError bool) error {
				for _, expressionValue := range expressions {
					*expressionValue = *expressionValue + "_result"
				}
				return nil
			})

		camelCaseSheetId := strings.ToLower(sheetId[0:2]) + strings.ToUpper(sheetId[2:3]) + strings.ToLower(sheetId[3:])

		cellListRef, err := sheet.GetCellList(camelCaseSheetId)
		cellList := *cellListRef

		assert.NotNil(t, cellList)
		assert.NoError(t, err)

		assert.Len(t, cellList, 2)

		assert.Equal(t, "value1", cellList["cell1"].Value)
		assert.Equal(t, "value1_result", cellList["cell1"].Result)

		assert.Equal(t, "value2", cellList["cell2"].Value)
		assert.Equal(t, "value2_result", cellList["cell2"].Result)
	})

	t.Run("not-exists-sheet", func(t *testing.T) {
		sheet := &SheetRepository{
			db:             db,
			executor:       nil,
			canonicalizer:  NewCanonicalizer(),
			serializer:     NewCellBinarySerializer(),
			dependencyTree: &CellDependencyTree{},
		}

		cellListRef, err := sheet.GetCellList("not-exits-sheet1")
		cellList := *cellListRef

		assert.NotNil(t, cellList)
		assert.Empty(t, cellList)
		assert.Error(t, err)
		assert.ErrorIs(t, err, contracts.SheetNotFoundError)
	})

	t.Run("execute-error", func(t *testing.T) {
		executor := mocks.NewExpressionExecutor(t)
		sheet := &SheetRepository{
			db:             db,
			executor:       executor,
			canonicalizer:  NewCanonicalizer(),
			serializer:     NewCellBinarySerializer(),
			dependencyTree: &CellDependencyTree{},
		}

		expectedMaps := contracts.ExpressionsMap{
			canonicalizer.Canonicalize("cell1"): _makeStringRef("value1"),
			canonicalizer.Canonicalize("cell2"): _makeStringRef("value2"),
		}

		executor.On("MultiEvaluate", expectedMaps, mock.Anything, false).Return(errors.New("execute-error"))

		cellListRef, err := sheet.GetCellList(sheetId)
		cellList := *cellListRef

		assert.NotNil(t, cellList)
		assert.Error(t, err)
		assert.Equal(t, "execute-error", err.Error())

		assert.Len(t, cellList, 2)

		assert.Equal(t, "value1", cellList["cell1"].Value)
		assert.Equal(t, "value1", cellList["cell1"].Result)

		assert.Equal(t, "value2", cellList["cell2"].Value)
		assert.Equal(t, "value2", cellList["cell2"].Result)
	})
}

func _prepareSheet(t *testing.T, sheetId string) *bbolt.DB {
	db, dbClose := _createTmpDb()
	defer dbClose()

	executor := mocks.NewExpressionExecutor(t)
	executor.On("MultiEvaluate", mock.Anything, mock.Anything, mock.Anything).Return(nil).Maybe()
	executor.On("ExtractDependingOnList", mock.Anything).Return([]string{})

	webhookDispatcher := mocks.NewWebhookDispatcher(t)
	webhookDispatcher.On("Notify", mock.Anything, mock.Anything).Return().Maybe()

	sheet := &SheetRepository{
		db:                db,
		executor:          executor,
		canonicalizer:     NewCanonicalizer(),
		serializer:        NewCellBinarySerializer(),
		dependencyTree:    &CellDependencyTree{},
		webhookDispatcher: webhookDispatcher,
	}

	_, err, _ := sheet.SetCell(sheetId, "cell1", "value1", true)
	assert.NoError(t, err)

	_, err, _ = sheet.SetCell(sheetId, "cell2", "value2", true)
	assert.NoError(t, err)

	// finish prepare sheet

	path := db.Path()
	db.Close()
	// re-open DB to ensure it stored at disk
	db, err = bbolt.Open(path, 0600, nil)
	assert.NoError(t, err)

	return db
}

func _createTmpDb() (*bbolt.DB, func()) {
	f, _ := os.CreateTemp("", "db_*.db")
	os.Remove(f.Name())

	db, dbErr := bbolt.Open(f.Name(), 0600, nil)
	if dbErr != nil {
		panic(dbErr)
	}

	return db, func() {
		db.Close()
		os.Remove(f.Name())
	}
}
