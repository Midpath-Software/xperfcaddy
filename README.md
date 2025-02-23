# Caddy X-Perf

A caddy module that will add `x-perf-caddy` header to all responses to indicate how much time the request has spent inside caddy.

Assuming your `reverse_proxy` upstream adds `x-perf-upstream` header, that information coupled with `x-perf-caddy` can help quickly diagnosing latency issues.

For example let us consider these cases where timings are in seconds:

duration in browser | x-perf-caddy | x-perf-upstream | case
---|---|---|---
5 | 4 | 4 | there is some network latency between browser and caddy
5 | 5 | 4 | the upstream is overloaded. The request spent 1s inside caddy waiting for upstream to be available
5 | 5 | - | When you use `response_header_timeout`, this request spent all 5 seconds waiting for an upstream and then was returned with an error since nobody was avaiable. This can be used to make sure your backend does not get overloaded in case of spikes.
