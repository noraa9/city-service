package repository

import "errors"

// ErrNotFound is returned when a DB query expected a row but got none.
//
// We expose a shared sentinel error so services can map it to a 404 (or a friendly message)
// without depending on driver-specific errors like sql.ErrNoRows.
var ErrNotFound = errors.New("not found")
