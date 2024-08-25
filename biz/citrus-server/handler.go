package citrus_server

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"net/http"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/common/hlog"
	"github.com/cloudwego/hertz/pkg/protocol/consts"
	"github.com/hertz-contrib/websocket"
	"github.com/tundrawork/DG-citrus/biz/handler"
	"github.com/tundrawork/DG-citrus/config"
	"golang.org/x/crypto/blake2b"
)

var (
	insecureIdSalt = generateSalt(8)
	citrusServer   = NewCitrusServer()
)

func DGAppHandler(ctx context.Context, c *app.RequestContext) {
	err := wsConnectionHandler(ctx, c, ClientTypeDGApp)
	if err != nil {
		hlog.CtxInfof(ctx, "RootHandler: try to handle connection as websocket failed: %v", err)
		wsUpgradeFailed(ctx, c)
	}
}

func ThirdPartyWSHandler(ctx context.Context, c *app.RequestContext) {
	err := wsConnectionHandler(ctx, c, ClientTypeThirdPartyWS)
	if err != nil {
		hlog.CtxInfof(ctx, "RootHandler: try to handle connection as websocket failed: %v", err)
		wsUpgradeFailed(ctx, c)
	}
}

func HTTPRegister(ctx context.Context, c *app.RequestContext) {
	insecureId := getInsecureIdFromRequest(c.ClientIP(), ClientTypeThirdPartyHTTP)
	if config.Conf.AllowInsecureClientId {
		_, err := citrusServer.getClientInsecure(insecureId)
		if err == nil {
			fail(ctx, c, "HTTPRegister", "We can not register you on this server as insecure client ID is enabled and your IP address is already registered.")
			return
		}
	}
	client := citrusServer.newHTTPClient(insecureId)
	event := &EventBindToServer{
		ClientId: client.secureId,
	}
	rawEvent, err := event.ToRawEvent()
	if err != nil {
		fail(ctx, c, "HTTPRegister", fmt.Sprintf("sendEvent: Failed to convert event to raw event: %v", err))
		return
	}
	c.JSON(http.StatusOK, rawEvent)
}

func HTTPBindingQrcode(ctx context.Context, c *app.RequestContext) {
	secureId, err := getSecureIdFromHTTPRequest(c)
	if err != nil {
		fail(ctx, c, "HTTPCommand", fmt.Sprintf("Failed to get client ID: %v", err))
		return
	}
	err = sendDGAppBindingCode(c.Response.BodyWriter(), config.Conf.HostName, secureId)
	if err != nil {
		fail(ctx, c, "HTTPBindingQrcode", fmt.Sprintf("Failed to generate DG-LAB app bindings code: %v", err))
		return
	}
	c.Response.Header.SetContentType(consts.MIMEImageJPEG)
}

func HTTPCommand(ctx context.Context, c *app.RequestContext) {
	secureId, err := getSecureIdFromHTTPRequest(c)
	if err != nil {
		fail(ctx, c, "HTTPCommand", fmt.Sprintf("Failed to get client ID: %v", err))
		return
	}
	var message string
	if message = c.Query("message"); message == "" {
		fail(ctx, c, "HTTPCommand", "No message provided")
		return
	}
	rawEvent := &RawEvent{
		Type:     EventTypeMsg,
		ClientId: string(secureId),
		TargetId: "",
		Message:  message,
	}
	event, err := rawEvent.ToEvent()
	if err != nil {
		fail(ctx, c, "HTTPCommand", fmt.Sprintf("Failed to parse event: %v", err))
		return
	}
	err = event.Process()
	if err != nil {
		fail(ctx, c, "HTTPCommand", fmt.Sprintf("Failed to process event: %v", err))
		return
	}
	c.JSON(http.StatusOK, map[string]interface{}{"code": 200, "message": "success"})
}

func HTTPHeartbeat(ctx context.Context, c *app.RequestContext) {
	secureId, err := getSecureIdFromHTTPRequest(c)
	if err != nil {
		fail(ctx, c, "HTTPHeartbeat", fmt.Sprintf("Failed to get client ID: %v", err))
		return
	}
	rawEvent := &RawEvent{
		Type:     EventTypeHeartbeat,
		ClientId: string(secureId),
		TargetId: "",
		Message:  "",
	}
	event, err := rawEvent.ToEvent()
	if err != nil {
		fail(ctx, c, "HTTPHeartbeat", fmt.Sprintf("Failed to parse event: %v", err))
		return
	}
	err = event.Process()
	if err != nil {
		fail(ctx, c, "HTTPHeartbeat", fmt.Sprintf("Failed to process event: %v", err))
		return
	}
	c.JSON(http.StatusOK, map[string]interface{}{"code": 200, "message": "success"})
}

func wsConnectionHandler(ctx context.Context, c *app.RequestContext, typ CitrusClientType) error {
	upgrader := websocket.HertzUpgrader{}
	err := upgrader.Upgrade(c, func(conn *websocket.Conn) {
		insecureId := getInsecureIdFromRequest(c.ClientIP(), typ)
		if config.Conf.AllowInsecureClientId {
			_, err := citrusServer.getClientInsecure(insecureId)
			if err == nil {
				fail(ctx, c, "wsConnectionHandler", "We can not register you on this server as insecure client ID is enabled and your IP address is already registered.")
				return
			}
		}
		dgClient := citrusServer.newWSClient(typ, insecureId, conn)
		defer citrusServer.purgeClient(insecureId)
		dgClient.serve()
	})
	if err != nil {
		return fmt.Errorf("wsConnectionHandler: Failed to upgrade connection: %v", err)
	}
	return nil
}

func generateSalt(length int) string {
	bytes := make([]byte, length)
	if _, err := rand.Read(bytes); err != nil {
		hlog.Errorf("generateSalt: Failed to generate salt: %v", err)
		panic(err)
	}
	return hex.EncodeToString(bytes)
}

func getInsecureIdFromRequest(clientIP string, clientType CitrusClientType) ClientInsecureId {
	hash, _ := blake2b.New(16, []byte(insecureIdSalt))
	hash.Write([]byte{byte(clientType)})
	return ClientInsecureId(hex.EncodeToString(hash.Sum([]byte(clientIP))))
}

func getSecureIdFromHTTPRequest(c *app.RequestContext) (ClientSecureId, error) {
	var secureId ClientSecureId
	if clientId := c.Query("clientId"); clientId == "" {
		if config.Conf.AllowInsecureClientId {
			insecureId := getInsecureIdFromRequest(c.ClientIP(), ClientTypeThirdPartyHTTP)
			dgClient, err := citrusServer.getClientInsecure(insecureId)
			if err != nil {
				return "", fmt.Errorf("can not match you with an existing client, this may caused by an IP address change of your device or network: %v", err)
			}
			secureId = dgClient.secureId
		} else {
			return "", fmt.Errorf("no client ID provided, insecure client ID is not allowed on this server")
		}
	} else {
		secureId = ClientSecureId(clientId)
		_, err := citrusServer.getClientSecure(secureId)
		if err != nil {
			return "", fmt.Errorf("can not find the client ID provided: %v", err)
		}
	}
	return secureId, nil
}

func wsUpgradeFailed(ctx context.Context, c *app.RequestContext) {
	c.Response.ResetBody()
	handler.HomeHandler(ctx, c)
}

func fail(ctx context.Context, c *app.RequestContext, context string, message string) {
	hlog.CtxWarnf(ctx, "%s: %s", context, message)
	c.JSON(http.StatusBadRequest, map[string]interface{}{"code": 400, "message": message})
}
