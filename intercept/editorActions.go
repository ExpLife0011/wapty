package intercept

import (
	"github.com/empijei/cli/lg"
	"github.com/empijei/wapty/ui"
	"github.com/empijei/wapty/ui/apis"
)

var uiEditor ui.Subscription

func init() {
	uiEditor = ui.Subscribe(apis.CHN_EDITOR)
}

//Invokes the edit action on the proxy ui. When a response is received returns
//the payload and the action in its string form. It does not attempt to validate
//the action, the caller must take care of it.
func editBuffer(p string, b []byte, endpoint string) ([]byte, apis.Action) {
	if !intercept.value() {
		return nil, apis.EDT_FORWARD
	}
	lg.Debugf("Editing: %s", p)
	args := map[apis.ArgName]string{
		apis.ARG_PAYLOADTYPE: p,
		apis.ARG_ENDPOINT:    endpoint}
	uiEditor.Send(&apis.Command{Action: apis.EDT_EDIT, Args: args, Payload: b})
	lg.Debug("Waiting for user interaction")
	result := uiEditor.Receive()
	lg.Debug("User interacted")
	return result.Payload, result.Action
}
