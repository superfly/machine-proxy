# Boot idle Fly VMs on demand via an http proxy

This is a proof-of-concept Go proxy that runs as a normal Fly app and proxies
requests to a single Fly VM. Should the VM be shut down, the proxy will start
the VM, blocking on HTTP requests until the VM is ready to handle them.

While this example is designed for a single VM backend, it could be easily
adapted to map subdomains or other request parameters to specific VMs. [The Fly
API Go adapter](https://pkg.go.dev/github.com/superfly/flyctl/api) allows
storing metadata on individual VM records (machines). These could store information
such a subdomain, and the subdomain-to-machine mapping might be fetched
at boot time by the proxy.

Currently, boot times vary between 1-2 seconds due the polling health checks against the API.
This could be improved by switching to a new internal API that allows direct communication
with the VM management service (flyd). 

# Setup

Install Flyctl and get a Fly account if you don't have one yet.

Setup this app on Fly by cloning this repo and running `fly launch`. Don't deploy yet.

Next, add your Fly API token (fetch with `fly auth token`) as a secret:

`fly secrets set FLY_AUTH_TOKEN=mytoken`

Then, create a new app and a VM to run upstream from the proxy.

```
fly apps create myappname
fly machines run nginx -a myappname 
```

Finally, update `fly.toml` `env` section to reflect the new app name and VM internal IPv6 address.
The IPv6 address will be picked up automatically by the next version of this proxy. For now,
you can grab it via the [GraphQL API playground](https://app.fly.io/graphql).

