package main

import (
	"bytes"
	"devChallengeExcel/contracts"
	"fmt"
	"go.etcd.io/bbolt"
	"strings"
)

type SheetRepository struct {
	db             *bbolt.DB
	executor       contracts.ExpressionExecutor
	serializer     contracts.CellSerializer
	canonicalizer  contracts.Canonicalizer
	dependencyTree contracts.CellDependencyTree
}

var errorNoChanges = fmt.Errorf("no changes")

func NewSheetRepository(
	db *bbolt.DB, executor contracts.ExpressionExecutor,
	serializer contracts.CellSerializer, canonicalizer contracts.Canonicalizer,
) *SheetRepository {
	return &SheetRepository{
		db:             db,
		executor:       executor,
		serializer:     serializer,
		canonicalizer:  canonicalizer,
		dependencyTree: &CellDependencyTree{},
	}
}

func (s *SheetRepository) SetCell(sheetId string, cellId string, value string) (cell *contracts.Cell, err error) {
	sheetId = strings.ToLower(sheetId)
	sheetIdByte := []byte(sheetId)

	cell = &contracts.Cell{Value: value}

	if strings.ContainsAny(cellId, contracts.CellIdBlacklist) {
		err = fmt.Errorf("cell_id `%s`: %w", cellId, contracts.CellIdBlacklistError)
		return
	}

	canonicalKey := s.canonicalizer.Canonicalize(cellId)
	canonicalKeyByte := []byte(canonicalKey)
	serializedData := s.serializer.Marshal(cellId, value)

	var dependants []string

	err = s.db.View(func(tx *bbolt.Tx) (err error) {
		readBucket := tx.Bucket(sheetIdByte)
		if readBucket == nil {
			dependants = make([]string, 0)
		} else {
			if bytes.Equal(readBucket.Get(canonicalKeyByte), serializedData) {
				cell.Result, err = s.executor.Evaluate(cell.Value, s.makeValuesGetter(tx, sheetIdByte))
				return errorNoChanges
			}

			dependants = s.dependencyTree.GetDependants(tx, sheetIdByte, canonicalKey)
		}

		cell.Result, err = s.executeWithDependantsCells(tx, sheetIdByte, canonicalKey, value, dependants)
		if err != nil {
			return
		}

		return err
	})

	if err != nil {
		if err == errorNoChanges {
			err = nil
		}
		return
	}

	dependingOnList := s.executor.ExtractDependingOnList(value)

	err = s.db.Batch(func(tx *bbolt.Tx) (err error) {
		var bucket *bbolt.Bucket
		bucket, err = tx.CreateBucketIfNotExists(sheetIdByte)
		if err != nil {
			return err
		}

		err = s.dependencyTree.SetDependsOn(tx, sheetIdByte, canonicalKey, dependingOnList)
		if err != nil {
			return
		}

		return bucket.Put(canonicalKeyByte, serializedData)
	})

	return
}

func (s *SheetRepository) GetCell(sheetId string, cellId string) (cell *contracts.Cell, err error) {
	sheetId = strings.ToLower(sheetId)

	cell = &contracts.Cell{}
	var byteValue []byte

	sheetIdByte := []byte(sheetId)

	canonicalKey := []byte(s.canonicalizer.Canonicalize(cellId))
	err = s.db.View(func(tx *bbolt.Tx) error {
		bucket := tx.Bucket(sheetIdByte)
		if bucket == nil {
			return fmt.Errorf("%s: %w", sheetId, contracts.SheetNotFoundError)
		}

		byteValue = bucket.Get(canonicalKey)

		if byteValue == nil {
			return fmt.Errorf("%s: %w", cellId, contracts.CellNotFoundError)
		}

		_, cell.Value, err = s.serializer.Unmarshal(byteValue)
		if err != nil {
			return err
		}

		cell.Result, err = s.executor.Evaluate(cell.Value, s.makeValuesGetter(tx, sheetIdByte))

		return err
	})

	return
}

func (s *SheetRepository) GetCellList(sheetId string) (*contracts.CellList, error) {
	sheetId = strings.ToLower(sheetId)

	cellList := contracts.CellList{}
	expressions := contracts.ExpressionsMap{}

	err := s.db.View(func(tx *bbolt.Tx) error {
		bucket := tx.Bucket([]byte(sheetId))
		if bucket == nil {
			return fmt.Errorf("%s: %w", sheetId, contracts.SheetNotFoundError)
		}

		c := bucket.Cursor()
		for k, v := c.First(); k != nil; k, v = c.Next() {
			canonicalCellId := string(k)
			key, value, err := s.serializer.Unmarshal(v)
			if err == nil {
				cellList[key] = &contracts.Cell{
					Value:  value,
					Result: value,
				}
				expressions[canonicalCellId] = &cellList[key].Result
			}
		}
		return nil
	})

	if err == nil {
		err = s.executor.MultiEvaluate(expressions, nil, false)
	}

	return &cellList, err
}

func (s *SheetRepository) makeValuesGetter(tx *bbolt.Tx, sheetId []byte) contracts.CellValuesGetter {
	return func(cellIds []string) []*string {
		return s.getCellValues(tx, sheetId, cellIds)
	}
}

func (s *SheetRepository) getCellValues(tx *bbolt.Tx, sheetId []byte, canonicalCellIds []string) []*string {
	values := make([]*string, len(canonicalCellIds))

	bucket := tx.Bucket(sheetId)

	if bucket == nil {
		return values
	}

	var byteValue []byte
	for index, canonicalCellId := range canonicalCellIds {
		byteValue = bucket.Get([]byte(canonicalCellId))
		if byteValue != nil {
			_, value, err := s.serializer.Unmarshal(byteValue)
			if err == nil {
				values[index] = &value
			}
		}
	}

	return values
}

func (s *SheetRepository) executeWithDependantsCells(tx *bbolt.Tx, sheetId []byte, cellId string, cellResult string, dependants []string) (string, error) {
	expressions := contracts.ExpressionsMap{
		cellId: &cellResult,
	}

	values := s.getCellValues(tx, sheetId, dependants)
	for index, dependingCellId := range dependants {
		if values[index] != nil {
			expressions[dependingCellId] = values[index]
		}
	}

	err := s.executor.MultiEvaluate(expressions, s.makeValuesGetter(tx, sheetId), true)
	return *expressions[cellId], err
}
