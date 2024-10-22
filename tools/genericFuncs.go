package tools

type Serializer[D any, O any] func(object O) (D, error)
type Deserializer[D any, O any] func(data D) (O, error)

type AcceptHandler[C any] func(C)
type AcceptHandlerWithError[C any] func(C) error

type ReadHandler[D any, C any] func(D, C)

type ReadHandlerWithResult[D any, C any] func(D, C) (D, error)

type ReadHandlerWithError[D any, C any] func(D, C) error

/*
type ByteHandler[D any, C any] func(D, C) error
type ObjectDeserializer[D any, O any, C any] func(D, C) (O, error)
type ObjectHandler[O any, C any] func(O, C) error

type ReadHandlerQueueWrapper[O any, C any] struct {
	object O
	caller C
}

func NewDefaultReadHandler[D any, O any, C any](
	byteHandler ByteHandler[D, C],
	deserializer ObjectDeserializer[D, O, C],
	objectHandler ObjectHandler[O, C],
) ReadHandler[D, C] {
	return func(bytes D, caller C) {

		err := byteHandler(bytes, caller)
		if err != nil {
			return
		}

		object, err := deserializer(bytes, caller)
		if err != nil {
			return
		}

		objectHandler(object, caller)
	}
}

// executes all handlers in order, return error if any handler returns an error
func NewObjectHandler[O any, C any](handlers ...ObjectHandler[O, C]) ObjectHandler[O, C] {
	return func(object O, caller C) error {
		for _, handler := range handlers {
			if err := handler(object, caller); err != nil {
				return err
			}
		}
		return nil
	}
}

type ObtainReadHandlerEnqueueConfigs[O any, C any] func(O, C) (token string, priority uint32, timeout uint32)

func NewQueueObjectHandler[O any, C any](
	priorityTokenQueue *PriorityTokenQueue[*ReadHandlerQueueWrapper[O, C]],
	obtainEnqueueConfigs ObtainReadHandlerEnqueueConfigs[O, C],
) ObjectHandler[O, C] {
	return func(object O, caller C) error {
		token, priority, timeoutMs := obtainEnqueueConfigs(object, caller)
		queueWrapper := &ReadHandlerQueueWrapper[O, C]{
			object,
			caller,
		}
		return priorityTokenQueue.Push(token, queueWrapper, priority, timeoutMs)
	}
}

type ObjectValidator[O any, C any] func(O, C) error

func NewValidationObjectHandler[O any, C any](validator ObjectValidator[O, C]) ObjectHandler[O, C] {
	return func(object O, caller C) error {
		return validator(object, caller)
	}
}

// executes all handlers in order, return error if any handler returns an error
func NewByteHandler[D any, C any](handlers ...ByteHandler[D, C]) ByteHandler[D, C] {
	return func(bytes D, caller C) error {
		for _, handler := range handlers {
			if err := handler(bytes, caller); err != nil {
				return err
			}
		}
		return nil
	}
}

type ObtainTokensFromBytes[D any] func(D) uint64

func NewTokenBucketRateLimitHandler[D any, C any](obtainTokensFromBytes ObtainTokensFromBytes[D], tokenBucketRateLimiterConfig *Config.TokenBucketRateLimiter) ByteHandler[D, C] {
	tokenBucketRateLimiter := NewTokenBucketRateLimiter(tokenBucketRateLimiterConfig)
	return func(bytes D, caller C) error {
		tokens := obtainTokensFromBytes(bytes)
		tokenBucketRateLimiter.Consume(tokens)
		return nil
	}
}
*/

/* // executes all handlers in order, return error if any handler returns an error
func NewAcceptHandler[C any](handlers ...InternalAcceptHandler[C]) AcceptHandler[C] {
	return func(caller C) {
		for _, handler := range handlers {
			if err := handler(caller); err != nil {
				return
			}
		}
	}
} */

/*
type ObtainAcceptHandlerEnqueueConfigs[C any] func(C) (token string, priority uint32, timeout uint32)

func NewQueueAcceptHandler[C any](
	priorityTokenQueue *PriorityTokenQueue[C],
	obtainEnqueueConfigs ObtainAcceptHandlerEnqueueConfigs[C],
) InternalAcceptHandler[C] {
	return func(caller C) error {
		token, priority, timeoutMs := obtainEnqueueConfigs(caller)
		return priorityTokenQueue.Push(token, caller, priority, timeoutMs)
	}
}

type ObtainIp[C any] func(C) string

func NewControlledAcceptHandler[C any](
	ipRateLimiter *IpRateLimiter,
	blacklist *AccessControlList,
	whitelist *AccessControlList,
	obtainIp ObtainIp[C],
) InternalAcceptHandler[C] {
	return func(caller C) error {
		ip := obtainIp(caller)
		if ipRateLimiter != nil && !ipRateLimiter.RegisterConnectionAttempt(ip) {
			return errors.New("rate limited")
		}
		if blacklist != nil && blacklist.Contains(ip) {
			return errors.New("blacklisted")
		}
		if whitelist != nil && whitelist.ElementCount() > 0 && !whitelist.Contains(ip) {
			return errors.New("not whitelisted")
		}
		return nil
	}
} */

/*
func NewAccessControlAcceptionHandler[O any](blacklist *Tools.AccessControlList, whitelist *Tools.AccessControlList, ipRateLimiter *Tools.IpRateLimiter, handshakeHandler func(*WebsocketClient.WebsocketClient) (string, error)) AcceptionHandler[O] {
	return func(websocketServer *WebsocketServer[O], websocketClient *WebsocketClient.WebsocketClient) (string, error) {
		ip, _, err := net.SplitHostPort(websocketClient.GetAddress())
		if err != nil {
			if websocketServer.GetEventHandler() != nil {
				websocketServer.GetEventHandler().Handle(Event.New(
					Event.SplittingHostPortFailed,
					Event.Context{
						Event.Address: websocketClient.GetAddress(),
						Event.Error:   err.Error(),
					},
					Event.Skip,
				))
			}
			return "", err
		}

		if ipRateLimiter != nil && !ipRateLimiter.RegisterConnectionAttempt(ip) {
			if websocketServer.GetEventHandler() != nil {
				event := websocketServer.GetEventHandler().Handle(Event.New(
					Event.RateLimited,
					Event.Context{
						Event.Address:         websocketClient.GetAddress(),
						Event.RateLimiterType: Event.Ip,
					},
					Event.Skip,
					Event.Continue,
				))
				if event.GetAction() == Event.Skip {
					return "", errors.New("rate limited")
				}
			} else {
				return "", errors.New("rate limited")
			}
		}

		if blacklist != nil && blacklist.Contains(ip) {
			if websocketServer.GetEventHandler() != nil {
				event := websocketServer.GetEventHandler().Handle(Event.New(
					Event.Blacklisted,
					Event.Context{
						Event.Address: websocketClient.GetAddress(),
					},
					Event.Skip,
					Event.Continue,
				))
				if event.GetAction() == Event.Skip {
					return "", errors.New("blacklisted")
				}
			} else {
				return "", errors.New("blacklisted")
			}
		}

		if whitelist != nil && whitelist.ElementCount() > 0 && !whitelist.Contains(ip) {
			if websocketServer.GetEventHandler() != nil {
				event := websocketServer.GetEventHandler().Handle(Event.New(
					Event.NotWhitelisted,
					Event.Context{
						Event.Address: websocketClient.GetAddress(),
					},
					Event.Skip,
					Event.Continue,
				))
				if event.GetAction() == Event.Skip {
					return "", errors.New("not whitelisted")
				}
			} else {
				return "", errors.New("not whitelisted")
			}
		}

		identity := ""
		if handshakeHandler != nil {
			identity_, err := handshakeHandler(websocketClient)
			if err != nil {
				if websocketServer.GetEventHandler() != nil {
					event := websocketServer.GetEventHandler().Handle(Event.New(
						Event.HandshakeFailed,
						Event.Context{
							Event.Address:  websocketClient.GetAddress(),
							Event.Identity: identity,
							Event.Error:    err.Error(),
						},
						Event.Skip,
						Event.Continue,
					))
					if event.GetAction() == Event.Skip {
						return "", err
					}
				} else {
					return "", err
				}
			}
			identity = identity_
		}

		return identity, nil
	}
}
*/

/*

func (server *WebsocketListener) acceptionRoutine() {
	defer func() {
		if server.eventHandler != nil {
			server.eventHandler.Handle(Event.New(
				Event.AcceptionRoutineEnds,
				Event.Context{},
				Event.Continue,
			))
		}
		server.waitGroup.Done()
	}()

	if server.eventHandler != nil {
		event := server.eventHandler.Handle(Event.New(
			Event.AcceptionRoutineBegins,
			Event.Context{},
			Event.Continue,
			Event.Cancel,
		))
		if event.GetAction() == Event.Cancel {
			return
		}
	}

	handleAcceptionWrapper := func(websocketClient *WebsocketClient.WebsocketClient) {
		if identity, err := server.acceptionHandler(server, websocketClient); err == nil {
			session := server.createSession(identity, websocketClient)
			if session == nil {
				server.ClientsRejected.Add(1)
				websocketClient.Close()
				return
			}
			server.ClientsAccepted.Add(1)
			server.waitGroup.Add(1)
			go server.websocketClientDisconnect(session, websocketClient)
		} else {
			server.ClientsRejected.Add(1)
			websocketClient.Close()
		}
	}

	for {
		websocketClient, err := server.websocketListener.Accept(server.config.WebsocketClientConfig, server.config.AcceptTimeoutMs)
		if err != nil {
			websocketClient.Close()
			if server.eventHandler != nil {
				event := server.eventHandler.Handle(Event.New(
					Event.AcceptClientFailed,
					Event.Context{
						Event.Address: websocketClient.GetAddress(),
					},
					Event.Skip,
					Event.Cancel,
				))
				if event.GetAction() == Event.Cancel {
					break
				}
			}
			continue
		}

		if server.config.HandleClientsSequentially {
			handleAcceptionWrapper(websocketClient)
		} else {
			server.waitGroup.Add(1)
			go func(websocketClient *WebsocketClient.WebsocketClient) {
				handleAcceptionWrapper(websocketClient)
				server.waitGroup.Done()
			}(websocketClient)
		}
	}
}

func (server *WebsocketServer[O]) createSession(identity string, websocketClient *WebsocketClient.WebsocketClient) *Tools.Session {
	for {
		if server.eventHandler != nil {
			event := server.eventHandler.Handle(Event.New(
				Event.CreatingSession,
				Event.Context{
					Event.Address: websocketClient.GetAddress(),
				},
				Event.Continue,
				Event.Skip,
			))
			if event.GetAction() == Event.Skip {
				return nil
			}
		}

		session, err := server.sessionManager.CreateSession(identity, map[string]any{
			"websocketClient": websocketClient,
		})
		if err != nil {
			if server.eventHandler != nil {
				event := server.eventHandler.Handle(Event.New(
					Event.CreateSessionFailed,
					Event.Context{
						Event.Address:  websocketClient.GetAddress(),
						Event.Identity: identity,
					},
					Event.Skip,
					Event.Retry,
				))
				if event.GetAction() == Event.Retry {
					continue
				}
			}
			return nil
		}

		if server.eventHandler != nil {
			event := server.eventHandler.Handle(Event.New(
				Event.CreatedSession,
				Event.Context{
					Event.Address:   websocketClient.GetAddress(),
					Event.SessionId: session.GetId(),
					Event.Identity:  session.GetIdentity(),
				},
				Event.Continue,
				Event.Skip,
				Event.Retry,
			))
			if event.GetAction() == Event.Skip {
				session.GetTimeout().Trigger()
				return nil
			}
			if event.GetAction() == Event.Retry {
				session.GetTimeout().Trigger()
				continue
			}
		}
		return session
	}
}

func (server *WebsocketServer[O]) websocketClientDisconnect(session *Tools.Session, websocketClient *WebsocketClient.WebsocketClient) {
	defer server.waitGroup.Done()

	select {
	case <-websocketClient.GetCloseChannel():
	case <-session.GetTimeout().GetTriggeredChannel():
	case <-server.stopChannel:
	}

	if server.eventHandler != nil {
		server.eventHandler.Handle(Event.New(
			Event.OnDisconnect,
			Event.Context{
				Event.Address:   websocketClient.GetAddress(),
				Event.Identity:  session.GetId(),
				Event.SessionId: session.GetIdentity(),
			},
			Event.Continue,
		))
	}

	session.GetTimeout().Trigger()
	websocketClient.Close()
}

*/