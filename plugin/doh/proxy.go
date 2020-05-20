package doh

import (
	"context"
	"encoding/base64"
	"fmt"
	"github.com/miekg/dns"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"io/ioutil"
	"net/http"
	"strconv"
	"time"
)

// Proxy defines an upstream host.
type Proxy struct {
	addr string

	// connection
	client   *http.Client
}

// newProxy returns a new proxy.
func newProxy(addr string, transport * http.Transport) (*Proxy, error) {
	p := &Proxy{
		addr: addr,
	}


	client := &http.Client{Transport: transport}

	p.client = client

	return p, nil
}

// query sends the request and waits for a response.
func (p *Proxy) query(ctx context.Context, req *dns.Msg) (*dns.Msg, error) {
	start := time.Now()

	msg, err := req.Pack()
	if err != nil {
		return nil, err
	}

	query:= base64.StdEncoding.EncodeToString(msg)
	request,err := http.NewRequestWithContext(ctx,"GET",
		fmt.Sprintf("https://%s/dns-query?dns=%s", p.addr,query),
		nil  )
	request.Header.Add("accept","application/dns-message")
	if err != nil {
		return nil, err
	}
	reply, err := p.client.Do(request)
	if err != nil {
		// if not found message, return empty message with NXDomain code
		if status.Code(err) == codes.NotFound {
			m := new(dns.Msg).SetRcode(req, dns.RcodeNameError)
			return m, nil

		}
		return nil, err
	}
	defer reply.Body.Close()
	body, err := ioutil.ReadAll(reply.Body)
	fmt.Printf("%s",body)

	ret := new(dns.Msg)
	if err := ret.Unpack(body); err != nil {
		return nil, err
	}
	rc, ok := dns.RcodeToString[ret.Rcode]
	if !ok {
		rc = strconv.Itoa(ret.Rcode)
	}

	RequestCount.WithLabelValues(p.addr).Add(1)
	RcodeCount.WithLabelValues(rc, p.addr).Add(1)
	RequestDuration.WithLabelValues(p.addr).Observe(time.Since(start).Seconds())
	fmt.Printf("%s",ret.Answer)

	return ret, nil
}
