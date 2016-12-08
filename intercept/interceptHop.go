package intercept

import (
	"log"
	"net/http"
)

//Remove trailers?
//https://github.com/squid-cache/squid/blob/master/src/http/RegisteredHeadersHash.cci
var HopByHopHeaders = []string{
	"Content-Encoding",
	"Connection",
	"TE",
	"HTTP2-Settings",
	"Keep-Alive",
	"Proxy-Authenticate",
	"Proxy-Connection",
	"Proxy-Authorization",
	"Trailer",
	"Upgrade",
	"Transfer-Encoding",
	"Alternate-Protocol",
	"X-Forwarded-For",
	"Proxy-Connection",
}

func stripHTHHeaders(h *http.Header) {
	for _, header := range HopByHopHeaders {
		h.Del(header)
	}
}

//This is a struct that respects the net.RoundTripper interface and just wraps
//the original http.RoundTripper
type Interceptor struct {
	wrappedRT http.RoundTripper
}

//This is a mock RoundTrip used to intercept responses before they are forwarded by the proxy
func (ri *Interceptor) RoundTrip(req *http.Request) (res *http.Response, err error) {
	//log.Println("Request read by proxy")
	intercepted := intercept.Value()
	//log.Println("Preprocessing...")
	backUpURL := req.URL
	req, Id, err := preProcessRequest(req)
	//log.Println("...done")
	if err != nil {
		//TODO handle possible autodrop
		//TODO other errors
		log.Println(err)
	}
	if intercepted {
		var editedReq *http.Request
		editedReq, res, err = editRequest(req, Id)
		if err != nil {
			//TODO
			log.Println(err)
		}
		if editedReq != nil {
			req = editedReq
			req.URL.Scheme = backUpURL.Scheme
			req.URL.Host = backUpURL.Host
		}
	}

	status.RLock()
	status.ReqResps[Id].parseRequest(req)
	status.RUnlock()
	if res != nil {
		//TODO Adding dropped responses should be avoided.
		status.addResponse(Id, res)
		return
	}

	//Perform the request, but disable compressing.
	//The gzip encoding will be used by the http package
	req.Header.Del("Accept-Encoding")
	res, err = ri.wrappedRT.RoundTrip(req)
	if err != nil {
		log.Println("Something went wrong trying to contact the server")
		return
	}
	res = preProcessResponse(req, res, Id)
	if intercepted {
		res, err = editResponse(req, res, Id)
	}
	status.RLock()
	status.ReqResps[Id].parseResponse(res)
	status.RUnlock()
	return
}