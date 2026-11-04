package anki

import (
	"github.com/Sobaze/Shippai-kizuki/backend/internal/storage"
	"github.com/Sobaze/Shippai-kizuki/backend/internal/testutil"
)

var testDatabase = testutil.Database
var openDatabase = storage.Open
