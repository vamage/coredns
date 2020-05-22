# doh

## Name

*doh* - facilitates proxying DNS messages to upstream resolvers via dOH protocol.

## Description

The *doh* plugin supports DoH.

This plugin can only be used once per Server Block.

## Syntax

In its most basic form:

~~~
doh FROM TO...
~~~

* **FROM** is the base domain to match for the request to be proxied.
* **TO...** are the destination endpoints to proxy to. The number of upstreams is
  limited to 15.

Multiple upstreams are randomized (see `policy`) on first use. When a proxy returns an error
the next upstream in the list is tried.

Extra knobs are available with an expanded syntax:

~~~
doh FROM TO... {
    except IGNORED_NAMES...

    policy random|round_robin|sequential
}
~~~

* **FROM** and **TO...** as above.
* **IGNORED_NAMES** in `except` is a space-separated list of domains to exclude from proxying.
  Requests that match none of these names will be passed through.

* `policy` specifies the policy to use for selecting upstream servers. The default is `random`.

Also note the TLS config is "global" for the whole doh proxy if you need a different
`tls-name` for different upstreams you're out of luck.

## Metrics

If monitoring is enabled (via the *prometheus* plugin) then the following metric are exported:

* `coredns_doh_request_duration_seconds{to}` - duration per upstream interaction.
* `coredns_doh_requests_total{to}` - query count per upstream.
* `coredns_doh_responses_total{to, rcode}` - count of RCODEs per upstream.
  and we are randomly (this always uses the `random` policy) spraying to an upstream.

## Examples

Proxy all requests within `example.org.` to a nameserver running on a different port:

~~~ corefile
example.org {
    doh . 127.0.0.1:443
}
~~~

Load balance all requests between three resolvers, one of which has a IPv6 address.

~~~ corefile
. {
    doh . 10.0.0.10:443 10.0.0.11:443 [2003::1]:443
}
~~~

Forward everything except requests to `example.org`

~~~ corefile
. {
    doh . 10.0.0.10:443 {
        except example.org
    }
}
~~~

Proxy everything except `example.org` using the host's `resolv.conf`'s nameservers:

~~~ corefile
. {
    doh . /etc/resolv.conf {
        except example.org
    }
}
~~~



## Bugs


