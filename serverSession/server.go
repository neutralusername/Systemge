package serverSession

import (
	"github.com/neutralusername/systemge/server"
	"github.com/neutralusername/systemge/tools"
)

type sessionServer[D any] struct {
}

func New[D any](
	sessionManager *tools.SessionManager,
	server *server.Server[D],
) server.Server[D] {

}
