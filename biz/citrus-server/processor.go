package citrus_server

import (
	"fmt"
	"strconv"

	"github.com/cloudwego/hertz/pkg/common/hlog"
)

func (e *EventBreak) Process() error {
	return fmt.Errorf("should never receive EventBreak")
}

func (e *EventBindToServer) Process() error {
	return fmt.Errorf("should never receive EventBindToServer")
}

func (e *EventBindResult) Process() error {
	return fmt.Errorf("should never receive EventBindResult")
}

func (e *EventError) Process() error {
	hlog.Warnf("[Processor] Received error: appId = %s, thirdPartyId = %s, message = %s", e.TargetId, e.ClientId, e.Message)
	return nil
}

func (e *EventHeartbeat) Process() error {
	hlog.Infof("[Processor] Received heartbeat: appId = %s, thirdPartyId = %s", e.TargetId, e.ClientId)
	return nil
}

func (e *EventBindAppToThirdParty) Process() error {
	hlog.Infof("[Processor] Received bind app to third party: appId = %s, thirdPartyId = %s", e.TargetId, e.ClientId)
	event := &EventBindResult{
		ClientId: e.ClientId,
		TargetId: e.TargetId,
	}
	err := citrusServer.bindClients(e.TargetId, e.ClientId)
	if err != nil {
		hlog.Errorf("[Processor] Failed to bind app to third party: appId = %s, thirdPartyId = %s, error = %v", e.TargetId, e.ClientId, err)
		event.Code = 400
		err = citrusServer.sendEvent(e.TargetId, event)
		if err != nil {
			return err
		}
	}

	event.Code = 200
	err = citrusServer.sendEvent(e.TargetId, event)
	if err != nil {
		return err
	}
	client, err := citrusServer.getClientSecure(e.ClientId)
	if err != nil {
		return err
	}
	if client.typ == ClientTypeThirdPartyWS {
		err = citrusServer.sendEvent(e.ClientId, event)
		if err != nil {
			return err
		}
	}
	return nil
}

func (e *EventReportStrength) Process() error {
	hlog.Infof("[Processor] Received report strength: appId = %s, thirdPartyId = %s (ignored), strength = %+v", e.TargetId, e.ClientId, e.Strength)
	bindings, err := citrusServer.getClientBindings(e.TargetId)
	if err != nil {
		err = failWithCode(e.ClientId, e.TargetId, 403)
		if err != nil {
			return err
		}
	}
	for _, binding := range bindings {
		// Only forward to websocket clients
		if binding.typ == ClientTypeThirdPartyWS {
			hlog.Infof("[Processor] Forwarding report strength to third party: appId = %s, thirdPartyId = %s", e.TargetId, binding.secureId)
			err = citrusServer.sendEvent(binding.secureId, e)
			if err != nil {
				hlog.Errorf("[Processor] Failed to forward report strength to third party: appId = %s, thirdPartyId = %s, error = %v", e.TargetId, binding.secureId, err)
			}
		}
	}
	return nil
}

func (e *EventAdjustStrength) Process() error {
	hlog.Infof("[Processor] Received adjust strength: thirdPartyId = %s, appId = %s (ignored), strength = %+v", e.ClientId, e.TargetId, e.Strength)
	bindings, err := citrusServer.getClientBindings(e.ClientId)
	if err != nil {
		err = failWithCode(e.ClientId, e.TargetId, 403)
		if err != nil {
			return err
		}
	}
	for _, binding := range bindings {
		hlog.Infof("[Processor] Forwarding adjust strength to DG-LAB app: thirdPartyId = %s, appId = %s", e.ClientId, binding.secureId)
		err = citrusServer.sendEvent(binding.secureId, e)
		if err != nil {
			hlog.Errorf("[Processor] Failed to forward adjust strength to DG-LAB app: thirdPartyId = %s, appId = %s, error = %v", e.ClientId, binding.secureId, err)
		}
	}
	return nil
}

func (e *EventExecutePulse) Process() error {
	hlog.Infof("[Processor] Received execute pulse: thirdPartyId = %s, appId = %s (ignored), channel = %d, pulseSequences = %+v", e.ClientId, e.TargetId, e.Channel, e.PulseSequences)
	bindings, err := citrusServer.getClientBindings(e.ClientId)
	if err != nil {
		err = failWithCode(e.ClientId, e.TargetId, 403)
		if err != nil {
			return err
		}
	}
	for _, binding := range bindings {
		hlog.Infof("[Processor] Forwarding execute pulse to DG-LAB app: thirdPartyId = %s, appId = %s", e.ClientId, binding.secureId)
		err = citrusServer.sendEvent(binding.secureId, e)
		if err != nil {
			hlog.Errorf("[Processor] Failed to forward execute pulse to DG-LAB app: thirdPartyId = %s, appId = %s, error = %v", e.ClientId, binding.secureId, err)
		}
	}
	return nil
}

func (e *EventStopPulse) Process() error {
	hlog.Infof("[Processor] Received stop pulse: thirdPartyId = %s, appId = %s (ignored), channel = %d", e.ClientId, e.TargetId, e.Channel)
	bindings, err := citrusServer.getClientBindings(e.ClientId)
	if err != nil {
		err = failWithCode(e.ClientId, e.TargetId, 403)
		if err != nil {
			return err
		}
	}
	for _, binding := range bindings {
		hlog.Infof("[Processor] Forwarding stop pulse to DG-LAB app: thirdPartyId = %s, appId = %s", e.ClientId, binding.secureId)
		err = citrusServer.sendEvent(binding.secureId, e)
		if err != nil {
			hlog.Errorf("[Processor] Failed to forward stop pulse to DG-LAB app: thirdPartyId = %s, appId = %s, error = %v", e.ClientId, binding.secureId, err)
		}
	}
	return nil
}

func (e *EventReportFeedback) Process() error {
	hlog.Infof("[Processor] Received report feedback: appId = %s, thirdPartyId = %s (ignored), button = %+v", e.TargetId, e.ClientId, e.Button)
	bindings, err := citrusServer.getClientBindings(e.TargetId)
	if err != nil {
		err = failWithCode(e.ClientId, e.TargetId, 403)
		if err != nil {
			return err
		}
	}
	for _, binding := range bindings {
		// Only forward to websocket clients
		if binding.typ == ClientTypeThirdPartyWS {
			hlog.Infof("[Processor] Forwarding report feedback to third party: appId = %s, thirdPartyId = %s", e.TargetId, binding.secureId)
			err = citrusServer.sendEvent(binding.secureId, e)
			if err != nil {
				hlog.Errorf("[Processor] Failed to forward report feedback to third party: appId = %s, thirdPartyId = %s, error = %v", e.TargetId, binding.secureId, err)
			}
		}
	}
	return nil
}

func failWithCode(clientId ClientSecureId, targetId ClientSecureId, code int) error {
	event := &EventError{
		ClientId: clientId,
		TargetId: targetId,
		Message:  strconv.Itoa(code),
	}
	return citrusServer.sendEvent(targetId, event)
}
