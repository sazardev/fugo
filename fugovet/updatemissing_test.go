package fugovet_test

import (
	"testing"

	"golang.org/x/tools/go/analysis/analysistest"

	"github.com/sazardev/fugo/fugovet"
)

func TestUpdateMissing(t *testing.T) {
	analysistest.Run(t, analysistest.TestData(), fugovet.UpdateMissing, "updatemissing")
}
