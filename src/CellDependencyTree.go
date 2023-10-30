package main

import (
	"bytes"
	"go.etcd.io/bbolt"
)

type CellDependencyTree struct{}

const Delimiter = byte(0x00)

var bucketPrefix = [4]byte{'_', '_', 'd', '_'}

func (t *CellDependencyTree) SetDependsOn(tx *bbolt.Tx, sheetId []byte, dependantCellId string, dependingOnCellIds []string) (err error) {
	cellDependingListKey := t.makeDependingListKey(dependantCellId)

	bucketId := t.makeBucketId(sheetId)
	var bucket *bbolt.Bucket
	bucket, err = tx.CreateBucketIfNotExists(bucketId)
	if err != nil {
		return err
	}

	previousDependingListToDelete := map[string]bool{}
	previous := bucket.Get(cellDependingListKey)
	if previous != nil {
		for _, oldDependantCellId := range bytes.Split(previous, []byte{Delimiter}) {
			previousDependingListToDelete[string(oldDependantCellId)] = true
		}
	}

	addedRecords := false
	for _, dependingOnCellId := range dependingOnCellIds {
		if previousDependingListToDelete[dependingOnCellId] {
			// dependingOnCellId is already in the list and saved in the database. Remove it from delete list
			delete(previousDependingListToDelete, dependingOnCellId)
		} else {
			addedRecords = true
			err = bucket.Put(t.makeDependantKey(dependantCellId, dependingOnCellId), []byte{})
			if err != nil {
				return err
			}
		}
	}

	if addedRecords == false && len(previousDependingListToDelete) == 0 {
		return nil
	}

	// delete old dependants which is not configured anymore
	for oldDependantCellId := range previousDependingListToDelete {
		err = bucket.Delete(t.makeDependantKey(dependantCellId, oldDependantCellId))
		if err != nil {
			return err
		}
	}

	if len(dependingOnCellIds) == 0 {
		return bucket.Delete(cellDependingListKey)
	}

	newDependingOnCellIds := make([][]byte, 0, len(dependingOnCellIds))
	for _, dependingOnCellId := range dependingOnCellIds {
		newDependingOnCellIds = append(newDependingOnCellIds, []byte(dependingOnCellId))
	}
	return bucket.Put(cellDependingListKey, bytes.Join(newDependingOnCellIds, []byte{Delimiter}))
}

func (t *CellDependencyTree) GetDependants(tx *bbolt.Tx, sheetId []byte, dependingOnCellId string) []string {
	bucketId := t.makeBucketId(sheetId)

	bucket := tx.Bucket(bucketId)
	if bucket == nil {
		return []string{}
	}

	return t.fetchDependantsRecursive(bucket, dependingOnCellId, map[string]bool{
		dependingOnCellId: true,
	})
}

func (t *CellDependencyTree) makeBucketId(sheetId []byte) []byte {
	if sheetId == nil || len(sheetId) == 0 {
		return nil
	}

	return append(bucketPrefix[:], sheetId...)
}

func (t *CellDependencyTree) fetchDependantsRecursive(bucket *bbolt.Bucket, dependingOnCellId string, alreadyFetched map[string]bool) []string {
	dependants := t.fetchCellDependants(bucket, dependingOnCellId)

	for _, dependantCellId := range dependants {
		if !alreadyFetched[dependantCellId] {
			alreadyFetched[dependantCellId] = true
			dependants = append(dependants, t.fetchDependantsRecursive(bucket, dependantCellId, alreadyFetched)...)
		}
	}

	return dependants
}

func (t *CellDependencyTree) fetchCellDependants(bucket *bbolt.Bucket, dependingOnCellId string) []string {
	dependantCellIds := make([]string, 0, 5)
	c := bucket.Cursor()

	prefix := t.makeDependingOnPrefixKey(dependingOnCellId)
	prefixLength := len(prefix)
	for k, _ := c.Seek(prefix); k != nil && bytes.HasPrefix(k, prefix); k, _ = c.Next() {
		dependantCellIds = append(dependantCellIds, string(k[prefixLength:]))
	}

	return dependantCellIds
}

func (t *CellDependencyTree) makeDependingListKey(dependantCellId string) []byte {
	return append(
		[]byte{Delimiter, Delimiter},
		[]byte(dependantCellId)...,
	)
}

func (t *CellDependencyTree) makeDependingOnPrefixKey(dependingOnCellId string) []byte {
	return append([]byte(dependingOnCellId), Delimiter)
}

func (t *CellDependencyTree) makeDependantKey(dependantCellId string, dependingOnCellId string) []byte {
	return append(t.makeDependingOnPrefixKey(dependingOnCellId), []byte(dependantCellId)...)
}

/** Terms:
 * dependant of - залежний від - зависимый от
 * depending on - в залежності від - в зависимости от
 */
