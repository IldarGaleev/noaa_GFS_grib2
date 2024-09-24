// Package indexfile parse .idx files
package indexfile

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"
)

type rangeInfo struct {
	from uint64
	to   uint64
}

type IndexFile struct {
	offsets map[string]rangeInfo
}

func createKey(tag, layer string) string {
	return fmt.Sprintf("%s:%s", tag, layer)
}

func (f *IndexFile) GetOffset(tag string, layer string) (uint64, uint64, error) {
	if r, ok := f.offsets[createKey(tag, layer)]; ok {
		return r.from, r.to, nil
	} else {
		return 0, 0, ErrOffsetNotFound
	}
}

func newIndexFile(src map[string]*rangeInfo) *IndexFile {
	res := &IndexFile{
		offsets: make(map[string]rangeInfo, len(src)),
	}

	for k, v := range src {
		res.offsets[k] = *v
	}

	return res
}

var (
	ErrOpenIndexFile  = errors.New("failed to open index file")
	ErrCloseIndexFile = errors.New("failed to close index file")
	ErrParseIndexFile = errors.New("failed to parse index file")
	ErrOffsetNotFound = errors.New("offset not found")
)

func New(indexFileName string) (res *IndexFile, rErr error) {
	file, err := os.Open(indexFileName)

	defer func() {
		err := file.Close()
		if err != nil {
			rErr = errors.Join(ErrCloseIndexFile, err)
		}
	}()

	if err != nil {
		return nil, errors.Join(ErrOpenIndexFile, err)
	}

	scanner := bufio.NewScanner(file)

	offsets := make(map[string]*rangeInfo, 800)

	prevKey := ""
	for scanner.Scan() {
		cols := strings.Split(scanner.Text(), ":")

		col1, err := strconv.Atoi(cols[1])
		if err != nil {
			return nil, errors.Join(ErrParseIndexFile, err)
		}
		offset := uint64(col1)

		key := createKey(cols[3], cols[4])
		offsets[key] = &rangeInfo{
			from: offset,
		}

		if prevKey != "" {
			offsets[prevKey].to = offset - 1
		}

		prevKey = key
	}

	result := newIndexFile(offsets)
	return result, nil

}
