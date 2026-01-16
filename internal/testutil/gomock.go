package testutil

import (
	"testing"

	"go.uber.org/mock/gomock"
)

func NewController(t *testing.T) *gomock.Controller {
	t.Helper()

	ctrl := gomock.NewController(t)
	t.Cleanup(ctrl.Finish)

	return ctrl
}
