package db

import "embed"

// Migrations embeds all SQL migration files from the migrations directory.
//
//go:embed migrations/*.sql
var Migrations embed.FS
