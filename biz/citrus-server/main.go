package citrus_server

import (
	"fmt"
	"sync"

	"github.com/cloudwego/hertz/pkg/common/hlog"
	"github.com/google/uuid"
	"github.com/hertz-contrib/websocket"
)

type CitrusClientType int
type ClientSecureId string
type ClientInsecureId string

type CitrusServer struct {
	clients CitrusClients
}

type CitrusClients struct {
	secureMapping   map[ClientSecureId]*CitrusClient
	insecureMapping map[ClientInsecureId]*CitrusClient
	mutex           sync.RWMutex
}

type CitrusClient struct {
	typ        CitrusClientType
	secureId   ClientSecureId
	insecureId ClientInsecureId
	bindings   map[ClientSecureId]bool
	conn       *websocket.Conn
}

const (
	ClientTypeDGApp CitrusClientType = iota
	ClientTypeThirdPartyWS
	ClientTypeThirdPartyHTTP
)

func NewCitrusServer() *CitrusServer {
	return &CitrusServer{
		clients: CitrusClients{
			secureMapping:   make(map[ClientSecureId]*CitrusClient),
			insecureMapping: make(map[ClientInsecureId]*CitrusClient),
		},
	}
}

func (client *CitrusClient) serve() {
	event := &EventBindToServer{
		ClientId: client.secureId,
	}
	err := citrusServer.sendEvent(client.secureId, event)
	if err != nil {
		hlog.Errorf("serve: failed to send EventBindToServer: %s", err)
		return
	}
	for {
		typ, message, err := client.conn.ReadMessage()
		if err != nil {
			hlog.Errorf("serve: read message from conn failed: %v", err)
			break
		}

		switch typ {
		case websocket.TextMessage:
			rawEvent := &RawEvent{}
			err := rawEvent.FromByteArray(message)
			if err != nil {
				hlog.Errorf("serve: failed to parse message: %v", err)
				continue
			}
			event, err := rawEvent.ToEvent()
			if err != nil {
				hlog.Errorf("serve: failed to convert raw event to event: %v", err)
				continue
			}
			err = event.Process()
			if err != nil {
				hlog.Errorf("serve: failed to process event: %v", err)
				continue
			}
		case websocket.CloseMessage:
			hlog.Infof("serve: received close message")
			err := client.conn.Close()
			if err != nil {
				hlog.Errorf("serve: failed to close connection: %v", err)
			}
			break
		default:
			hlog.Errorf("serve: received unsupported message type: %d", typ)
		}
	}
}

func (server *CitrusServer) newWSClient(typ CitrusClientType, insecureId ClientInsecureId, conn *websocket.Conn) CitrusClient {
	server.clients.mutex.Lock()
	defer server.clients.mutex.Unlock()

	secureID := ClientSecureId(uuid.NewString())
	client := &CitrusClient{
		typ:        typ,
		secureId:   secureID,
		insecureId: insecureId,
		bindings:   make(map[ClientSecureId]bool),
		conn:       conn,
	}

	server.clients.secureMapping[secureID] = client
	server.clients.insecureMapping[insecureId] = client

	return *client
}

func (server *CitrusServer) newHTTPClient(insecureId ClientInsecureId) CitrusClient {
	server.clients.mutex.Lock()
	defer server.clients.mutex.Unlock()

	secureID := ClientSecureId(uuid.NewString())
	client := &CitrusClient{
		typ:        ClientTypeThirdPartyHTTP,
		secureId:   secureID,
		insecureId: insecureId,
		bindings:   make(map[ClientSecureId]bool),
	}

	server.clients.secureMapping[secureID] = client
	server.clients.insecureMapping[insecureId] = client

	return *client
}

func (server *CitrusServer) purgeClient(insecureId ClientInsecureId) {
	server.clients.mutex.Lock()
	defer server.clients.mutex.Unlock()

	hlog.Infof("purgeClient: purging client with insecure ID %s", insecureId)
	client, ok := server.clients.insecureMapping[insecureId]
	if !ok {
		hlog.Errorf("purgeClient: Client with insecure ID %s not found", insecureId)
		return
	}

	err := server.unbindClientFromAllBindings(client.secureId)
	if err != nil {
		return
	}

	delete(server.clients.secureMapping, client.secureId)
	delete(server.clients.insecureMapping, insecureId)
}

func (server *CitrusServer) getClientSecure(secureId ClientSecureId) (*CitrusClient, error) {
	server.clients.mutex.RLock()
	defer server.clients.mutex.RUnlock()

	client, ok := server.clients.secureMapping[secureId]
	if !ok {
		return nil, fmt.Errorf("getClientSecure: Client with secure ID %s not found", secureId)
	}

	return client, nil
}

func (server *CitrusServer) getClientInsecure(insecureId ClientInsecureId) (*CitrusClient, error) {
	server.clients.mutex.RLock()
	defer server.clients.mutex.RUnlock()

	client, ok := server.clients.insecureMapping[insecureId]
	if !ok {
		return nil, fmt.Errorf("getClientInsecure: Client with insecure ID %s not found", insecureId)
	}

	return client, nil
}

func (server *CitrusServer) bindClients(dgAppClientId ClientSecureId, thirdPartyClientId ClientSecureId) error {
	server.clients.mutex.Lock()
	defer server.clients.mutex.Unlock()

	dgAppClient, ok := server.clients.secureMapping[dgAppClientId]
	if !ok {
		return fmt.Errorf("bindClients: DG App client with secure ID %s not found", dgAppClientId)
	}
	if dgAppClient.typ != ClientTypeDGApp {
		return fmt.Errorf("bindClients: client with secure ID %s is not a DG App client", dgAppClientId)
	}
	thirdPartyClient, ok := server.clients.secureMapping[thirdPartyClientId]
	if !ok {
		return fmt.Errorf("bindClients: Third Party client with secure ID %s not found", thirdPartyClientId)
	}
	if thirdPartyClient.typ != ClientTypeThirdPartyWS && thirdPartyClient.typ != ClientTypeThirdPartyHTTP {
		return fmt.Errorf("bindClients: client with secure ID %s is not a Third Party client", thirdPartyClientId)
	}

	if _, ok := dgAppClient.bindings[thirdPartyClientId]; ok {
		return fmt.Errorf("bindClients: Clients with secure IDs %s and %s are already bound", dgAppClientId, thirdPartyClientId)
	}

	dgAppClient.bindings[thirdPartyClientId] = true
	thirdPartyClient.bindings[dgAppClientId] = true
	return nil
}

func (server *CitrusServer) unbindClientFromAllBindings(secureId ClientSecureId) error {
	server.clients.mutex.Lock()
	defer server.clients.mutex.Unlock()

	client, ok := server.clients.secureMapping[secureId]
	if !ok {
		return fmt.Errorf("unbindClientFromAllBindings: Client with secure ID %s not found", secureId)
	}

	for binding, _ := range client.bindings {
		peerClient, ok := server.clients.secureMapping[binding]
		if !ok {
			hlog.Errorf("unbindClientFromAllBindings: Client with secure ID %s not found", binding)
			continue
		}
		delete(peerClient.bindings, secureId)
	}
	client.bindings = make(map[ClientSecureId]bool)

	return nil
}

func (server *CitrusServer) getClientBindings(secureId ClientSecureId) ([]CitrusClient, error) {
	server.clients.mutex.RLock()
	defer server.clients.mutex.RUnlock()

	client, ok := server.clients.secureMapping[secureId]
	if !ok {
		return nil, fmt.Errorf("getClientBindings: Client with secure ID %s not found", secureId)
	}

	bindings := make([]CitrusClient, 0)
	for bindingId, _ := range client.bindings {
		binding, ok := server.clients.secureMapping[bindingId]
		if !ok {
			hlog.Errorf("getClientBindings: Binding with secure ID %s not found", bindingId)
			continue
		}
		bindings = append(bindings, *binding)
	}

	return bindings, nil
}

func (server *CitrusServer) sendEvent(secureId ClientSecureId, event Event) error {
	server.clients.mutex.RLock()
	defer server.clients.mutex.RUnlock()

	client, ok := server.clients.secureMapping[secureId]
	if !ok {
		return fmt.Errorf("sendEvent: Client with secure ID %s not found", secureId)
	}

	rawEvent, err := event.ToRawEvent()
	if err != nil {
		hlog.Errorf("sendEvent: Failed to convert event to raw event: %v", err)
		return err
	}
	data, err := rawEvent.ToByteArray()
	if err != nil {
		return fmt.Errorf("sendEvent: Failed to serialize event: %v", err)
	}
	err = client.conn.WriteMessage(websocket.TextMessage, data)
	if err != nil {
		return fmt.Errorf("sendEvent: WriteMessage failed: %v", err)
	}
	return nil
}
