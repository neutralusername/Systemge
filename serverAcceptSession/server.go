package serverAcceptSession

import (
	"github.com/neutralusername/systemge/systemge"
	"github.com/neutralusername/systemge/tools"
)

func New[D any](
	sessionManager *tools.SessionManager,
) func(connection systemge.Connection[D]) error {

}
