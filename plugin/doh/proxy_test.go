package doh

import (
	"context"
	"errors"
	"github.com/coredns/coredns/pb"
	"testing"


	"github.com/miekg/dns"
)

func TestProxy(t *testing.T) {
	tests := map[string]struct {
		p       *Proxy
		res     *dns.Msg
		wantErr bool
	}{
		"response_ok": {
			p:       &Proxy{},
			res:     &dns.Msg{},
			wantErr: false,
		},
		"nil_response": {
			p:       &Proxy{},
			res:     nil,
			wantErr: true,
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			var mock *testServiceClient
			if tt.res != nil {
				msg, err := tt.res.Pack()
				if err != nil {
					t.Fatalf("Error packing response: %s", err.Error())
				}
				mock = &testServiceClient{&pb.DnsPacket{Msg: msg}, nil}
			} else {
				mock = &testServiceClient{nil, errors.New("server error")}
			}
			tt.p.client = mock

			_, err := tt.p.query(context.TODO(), new(dns.Msg))
			if err != nil && !tt.wantErr {
				t.Fatalf("Error query(): %s", err.Error())
			}
		})
	}
}

type testServiceClient struct {
	dnsPacket *pb.DnsPacket
	err       error
}

func (m testServiceClient) Query(ctx context.Context, in *pb.DnsPacket) (*pb.DnsPacket, error) {
	return m.dnsPacket, m.err
}
