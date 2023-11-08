package main

import (
	"bytes"
	"devChallengeExcel/contracts"
	"fmt"
	"go.etcd.io/bbolt"
	"strings"
)

type SheetRepository struct {
	db                *bbolt.DB
	executor          contracts.ExpressionExecutor
	serializer        contracts.CellSerializer
	canonicalizer     contracts.Canonicalizer
	dependencyTree    contracts.CellDependencyTree
	webhookDispatcher contracts.WebhookDispatcher
}

var errorNoChanges = fmt.Errorf("no changes")

func NewSheetRepository(
	db *bbolt.DB, executor contracts.ExpressionExecutor,
	serializer contracts.CellSerializer, canonicalizer contracts.Canonicalizer,
	webhookDispatcher contracts.WebhookDispatcher,
) *SheetRepository {
	return &SheetRepository{
		db:                db,
		executor:          executor,
		serializer:        serializer,
		canonicalizer:     canonicalizer,
		dependencyTree:    &CellDependencyTree{},
		webhookDispatcher: webhookDispatcher,
	}
}

func (s *SheetRepository) GetCanonicalSheetId(sheetId string) string {
	return strings.ToLower(sheetId)
}

func (s *SheetRepository) SetCell(sheetId string, cellId string, value string, skipNotChanged bool) (cell *contracts.Cell, err error, isUpdated bool) {
	sheetId = s.GetCanonicalSheetId(sheetId)
	sheetIdByte := []byte(sheetId)

	if strings.ContainsAny(cellId, contracts.CellIdBlacklist) {
		err = fmt.Errorf("cell_id `%s`: %w", cellId, contracts.CellIdBlacklistError)
		cell = &contracts.Cell{Value: value}
		return
	}

	cellCanonicalKey := s.canonicalizer.Canonicalize(cellId)
	cellCanonicalKeyByte := []byte(cellCanonicalKey)
	serializedData := s.serializer.Marshal(cellId, value)

	cell = &contracts.Cell{
		CanonicalKey: cellCanonicalKey,
		Value:        value,
		Result:       value,
	}

	var dependants []string
	var dependantsCellList []*contracts.Cell

	err = s.db.View(func(tx *bbolt.Tx) (err error) {
		readBucket := tx.Bucket(sheetIdByte)
		if readBucket == nil {
			dependants = make([]string, 0)
		} else {
			if skipNotChanged && bytes.Equal(readBucket.Get(cellCanonicalKeyByte), serializedData) {
				cell.Result, err = s.executor.Evaluate(cell.Value, s.makeValuesGetter(tx, sheetIdByte))
				return errorNoChanges
			}

			dependants = s.dependencyTree.GetDependants(tx, sheetIdByte, cellCanonicalKey)
		}

		dependantsCellList = s.makeDependantsCellList(tx, sheetIdByte, cell, dependants)
		expressions := make(contracts.ExpressionsMap, len(dependantsCellList))
		for i := range dependantsCellList {
			expressions[dependantsCellList[i].CanonicalKey] = &dependantsCellList[i].Result
		}

		err = s.executor.MultiEvaluate(expressions, s.makeValuesGetter(tx, sheetIdByte), true)
		return err
	})

	isUpdated = err == nil
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

		err = s.dependencyTree.SetDependsOn(tx, sheetIdByte, cellCanonicalKey, dependingOnList)
		if err != nil {
			return
		}

		return bucket.Put(cellCanonicalKeyByte, serializedData)
	})

	s.webhookDispatcher.Notify(sheetId, dependantsCellList)

	return
}

func (s *SheetRepository) makeDependantsCellList(tx *bbolt.Tx, sheetId []byte, thisCell *contracts.Cell, dependants []string) []*contracts.Cell {
	values := s.getCellValues(tx, sheetId, dependants)

	dependantsCellList := make([]*contracts.Cell, 0, len(dependants)+1)
	dependantsCellList = append(dependantsCellList, thisCell)

	for index, dependantCanonicalCellId := range dependants {
		if values[index] != nil {
			dependantsCellList = append(dependantsCellList, &contracts.Cell{
				CanonicalKey: dependantCanonicalCellId,
				Value:        *values[index],
				Result:       *values[index],
			})
		}
	}

	return dependantsCellList
}

func (s *SheetRepository) GetCell(sheetId string, cellId string) (cell *contracts.Cell, err error) {
	sheetId = s.GetCanonicalSheetId(sheetId)

	var byteValue []byte
	sheetIdByte := []byte(sheetId)
	cell = &contracts.Cell{
		CanonicalKey: s.canonicalizer.Canonicalize(cellId),
	}
	canonicalKey := []byte(cell.CanonicalKey)
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
					CanonicalKey: canonicalCellId,
					Value:        value,
					Result:       value,
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
