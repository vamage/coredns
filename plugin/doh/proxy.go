package doh

import (
	"context"
	"crypto/tls"
	"encoding/base64"
	"fmt"
	"github.com/miekg/dns"
	"google.golang.org/grpc"
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
	dialOpts []grpc.DialOption
}

// newProxy returns a new proxy.
func newProxy(addr string, tlsConfig *tls.Config) (*Proxy, error) {
	p := &Proxy{
		addr: addr,
	}


	tr := &http.Transport{
		MaxIdleConns:       10,
		IdleConnTimeout:    30 * time.Second,
		DisableCompression: true,
	}
	client := &http.Client{Transport: tr}

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
	fmt.Println(fmt.Sprintf("https://1.1.1.1/dns-query?dns=%s,%s, %v",query,req))
	reply, err := p.client.Get(fmt.Sprintf("https://1.1.1.1/dns-query?dns=%s",query))
	fmt.Println(reply)
	fmt.Printf("%v\n", err)

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
	ret := new(dns.Msg)
	if err := ret.Unpack(body); err != nil {
		return nil, err
	}
	fmt.Printf("%v\n",ret)
	fmt.Println(body)
	rc, ok := dns.RcodeToString[ret.Rcode]
	if !ok {
		rc = strconv.Itoa(ret.Rcode)
	}

	RequestCount.WithLabelValues(p.addr).Add(1)
	RcodeCount.WithLabelValues(rc, p.addr).Add(1)
	RequestDuration.WithLabelValues(p.addr).Observe(time.Since(start).Seconds())

	return ret, nil
}
