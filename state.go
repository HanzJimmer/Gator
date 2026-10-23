package main

import (
	"github.com/HanzJimmer/Gator/internal/config"
	"github.com/HanzJimmer/Gator/internal/database"
)

type state struct {
	State *config.Config
	db    *database.Queries
}
