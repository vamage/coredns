package doh

import (
	"context"
	"encoding/base64"
	"fmt"
	"github.com/coredns/coredns/pb"
	"github.com/miekg/dns"
	"google.golang.org/grpc/status"
	"io/ioutil"
	"net/http"
	"strconv"
	"time"
)

// Proxy defines an upstream host.
type Proxy struct {
	addr string

	client   DnsServiceClient

	// connection
	clientHTTP *http.Client
}

type DnsServiceClient interface {
	Query(ctx context.Context, in *pb.DnsPacket) (*pb.DnsPacket, error)
}
type dnsServiceClient struct {
	cc *http.Client
	addr string
}
func NewDnsServiceClient(cc *http.Client,addr string) DnsServiceClient {
	return &dnsServiceClient{cc, addr}
}
func (c *dnsServiceClient) Query(ctx context.Context, req *pb.DnsPacket) (*pb.DnsPacket, error) {


	query:= base64.StdEncoding.EncodeToString(req.Msg)
	request,err := http.NewRequestWithContext(ctx,"GET",
		fmt.Sprintf("https://%s/dns-query?dns=%s", c.addr,query),
		nil  )
	request.Header.Add("accept","application/dns-message")
	if err != nil {
		return nil, err
	}
	reply, err := c.cc.Do(request)
	if err != nil {
		return nil, err
	}
	defer reply.Body.Close()
	body, err := ioutil.ReadAll(reply.Body)

	return &pb.DnsPacket{Msg: body}, nil
}
// newProxy returns a new proxy.
func newProxy(addr string, transport * http.Transport) (*Proxy, error) {
	p := &Proxy{
		addr: addr,
	}


	client := &http.Client{Transport: transport}

	p.client = NewDnsServiceClient(client,addr)

	return p, nil
}

// query sends the request and waits for a response.
func (p *Proxy) query(ctx context.Context, req *dns.Msg) (*dns.Msg, error) {
	start := time.Now()
	msg, err := req.Pack()
	if err != nil {
		return nil, err
	}
	reply, err := p.client.Query(ctx,&pb.DnsPacket{Msg: msg})
	if err != nil {
		return nil, err
	}
	// if not found message, return empty message with NXDomain code
	if status.Code(err) == http.StatusNotFound {
		m := new(dns.Msg).SetRcode(req, dns.RcodeNameError)
		return m, nil
	}
	ret := new(dns.Msg)
	if err := ret.Unpack(reply.Msg); err != nil {
		return nil, err
	}

	rc, ok := dns.RcodeToString[ret.Rcode]
	if !ok {
		rc = strconv.Itoa(ret.Rcode)
	}

	RequestCount.WithLabelValues(p.addr).Add(1)
	RcodeCount.WithLabelValues(rc, p.addr).Add(1)
	RequestDuration.WithLabelValues(p.addr).Observe(time.Since(start).Seconds())

	return ret, nil
}
