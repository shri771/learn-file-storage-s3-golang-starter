package main

import (
	"fmt"
	"path/filepath"

	"github.com/google/uuid"
)

func (cfg *apiConfig) GetFilePath(id uuid.UUID, fileType string) string {

	fileName := fmt.Sprintf("%v.%v", id, fileType)

	filePath := filepath.Join(cfg.assetsRoot, fileName)

	return filePath
}
