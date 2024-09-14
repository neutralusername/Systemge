package DashboardHelpers

import (
	"github.com/neutralusername/Systemge/Error"
	"github.com/neutralusername/Systemge/Status"
)

func (page *Page) GetCachedCommands() map[string]bool {
	switch page.Type {
	case CLIENT_TYPE_COMMAND:
		return page.Data.(*CommandClient).Commands
	case CLIENT_TYPE_CUSTOMSERVICE:
		return page.Data.(*CustomServiceClient).Commands
	case CLIENT_TYPE_SYSTEMGECONNECTION:
		return page.Data.(*SystemgeConnectionClient).Commands
	default:
		return nil
	}
}

func (page *Page) SetCachedCommands(commands map[string]bool) error {
	if commands == nil {
		return Error.New("Commands is nil", nil)
	}
	switch page.Type {
	case CLIENT_TYPE_COMMAND:
		page.Data.(*CommandClient).Commands = commands
		return nil
	case CLIENT_TYPE_CUSTOMSERVICE:
		page.Data.(*CustomServiceClient).Commands = commands
		return nil
	case CLIENT_TYPE_SYSTEMGECONNECTION:
		page.Data.(*SystemgeConnectionClient).Commands = commands
		return nil
	default:
		return Error.New("Unknown client type", nil)
	}
}

func (page *Page) GetCachedStatus() int {
	switch page.Type {
	case CLIENT_TYPE_CUSTOMSERVICE:
		return page.Data.(*CustomServiceClient).Status
	case CLIENT_TYPE_SYSTEMGECONNECTION:
		return page.Data.(*SystemgeConnectionClient).Status
	default:
		return Status.NON_EXISTENT
	}
}

func (page *Page) SetCachedStatus(status int) error {
	switch page.Type {
	case CLIENT_TYPE_CUSTOMSERVICE:
		page.Data.(*CustomServiceClient).Status = status
	case CLIENT_TYPE_SYSTEMGECONNECTION:
		page.Data.(*SystemgeConnectionClient).Status = status
	default:
		return Error.New("Unknown client type", nil)
	}
	return nil
}

func (page *Page) SetCachedIsProcessingLoopRunning(isProcessingLoopRunning bool) error {
	switch page.Type {
	case CLIENT_TYPE_SYSTEMGECONNECTION:
		page.Data.(*SystemgeConnectionClient).IsProcessingLoopRunning = isProcessingLoopRunning
	default:
		return Error.New("Unknown client type", nil)
	}
	return nil
}

func (page *Page) SetCachedUnprocessedMessageCount(unprocessedMessageCount uint32) error {
	switch page.Type {
	case CLIENT_TYPE_SYSTEMGECONNECTION:
		page.Data.(*SystemgeConnectionClient).UnprocessedMessageCount = unprocessedMessageCount
	default:
		return Error.New("Unknown client type", nil)
	}
	return nil
}
