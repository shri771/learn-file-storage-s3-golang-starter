package main

import (
	"fmt"
	"path/filepath"

	"github.com/google/uuid"
)

func (cfg *apiConfig) GetFilePath(id uuid.UUID, ext string) string {

	fileName := fmt.Sprintf("%v%v", id, ext)

	filePath := filepath.Join(cfg.assetsRoot, fileName)

	return filePath
}
