package main

import (
	"embed"
)

//go:embed ../../deployments/sqlite/migrations/*.sql
var embedMigrations embed.FS
