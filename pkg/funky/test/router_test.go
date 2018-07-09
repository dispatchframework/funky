package funky

import (
	"testing"

	"github.com/dispatchframework/funky/pkg/funky"
	"github.com/dispatchframework/funky/pkg/funky/mocks"
)

func TestNewRouterSuccess(t *testing.T) {
	expected := 2
	serverFactory := new(mocks.ServerFactory)
	router := funky.NewRouter(expected, serverFactory)
	router.Shutdown()

	t.Errorf("Invalid number of servers expecting %d and found %d", expected, 2)
}
