package serverSession

import (
	"github.com/neutralusername/systemge/systemge"
	"github.com/neutralusername/systemge/tools"
)

type sessionServer[D any] struct {
}

func New[D any](
	sessionManager *tools.SessionManager,
	server systemge.Server[D],
) systemge.Server[D] {

}
