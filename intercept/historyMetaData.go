package intercept

import (
	"encoding/json"
	"net"
	"net/http"
	"strings"
	"time"

	"github.com/empijei/wapty/cli/lg"
	"github.com/empijei/wapty/ui/apis"
)

//DISCLAIMER use original req AFTER editing the new one
//And use it from a thread that has a readlock on the status
func (rr *ReqResp) parseRequest(req *http.Request) {
	this := rr.MetaData
	this.Host = req.Host
	this.Method = req.Method
	this.Path = req.URL.Path
	//TODO implement this in a way that does not consume the body
	//if len(req.Form) == 0 {
	//	_ = req.ParseForm()
	//}
	//this.Params = len(req.Form) > 0
	//this supposes to alread have a RLock on the status.
	this.Edited = status.ReqResps[this.ID].RawEditedReq != nil
	tmp := strings.Split(this.Path, ".")
	if !strings.Contains(tmp[len(tmp)-1], "/") {
		this.Extension = tmp[len(tmp)-1]
	}
	ipport := strings.Split(this.Host, ":")
	ips, err := net.LookupHost(ipport[0])
	if err == nil && len(ips) >= 1 {
		this.IP = ips[0]
		if len(ipport) >= 2 {
			this.Port = ipport[1]
		} else {
			switch req.URL.Scheme {
			case "https":
				this.Port = "443"
			case "http":
				this.Port = "80"
			default:
				lg.Infof("Port not specified: %s\n", this.Host)
			}
		}
	} else {
		lg.Infof("Unable to resolve Host: %s\n", this.Host)
	}
	this.Time = time.Now().String()
	sendMetaData(this)
}

//DISCLAIMER use original res AFTER editing the new one
//And use it from a thread that has a readlock on the status
func (rr *ReqResp) parseResponse(res *http.Response) {
	this := rr.MetaData
	if !this.Edited {
		//this supposes to alread have a RLock on the status.
		this.Edited = status.ReqResps[this.ID].RawEditedRes != nil
	}
	this.Status = res.Status
	//FIXME
	this.Length = res.ContentLength
	this.ContentType = res.Header.Get("Content-Type")
	this.TLS = res.TLS != nil
	tmp := res.Cookies()
	for _, cookie := range tmp {
		this.Cookies += cookie.String() + "; "
	}
	sendMetaData(this)
}

func sendMetaData(metaData *apis.ReqRespMetaData) {
	metaJSON, err := json.Marshal(metaData)
	if err != nil {
		lg.Error(err)
	}
	uiHistory.Send(&apis.Command{Action: apis.HST_METADATA, Args: map[apis.ArgName]string{apis.HST_METADATA: string(metaJSON)}})
}
