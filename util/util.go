package util

import (
	"io/ioutil"
	"os"
)

func CreateTempFile(text []byte) (*os.File, error) {
	tmpFile, err := ioutil.TempFile(os.TempDir(), "prefix-")
	if err != nil {
		return nil, err
	}

	if _, err = tmpFile.Write(text); err != nil {
		return nil, err
	}

	if err := tmpFile.Close(); err != nil {
		return nil, err
	}

	return tmpFile, nil
}
