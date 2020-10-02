# Gloo OSS ExtAuth: Custom Auth Server for Gloo API Gateway (open source version of Gloo)

`gloo-oss-extauth` is a custom Envoy External Authorization server. It provides a gRPC server implementing the Envoy [AuthorizationServer interface](https://github.com/envoyproxy/go-control-plane/blob/master/envoy/service/auth/v2/external_auth.pb.go).

This custom authorization server has been tested with Gloo `1.4.11` (and should work with lower and higher versions), using JWT issued using a test Auth0 tenant.

## Why

Gloo open source version does not come with an Envoy External Authorization server (you need to subscribe to Gloo Enterprise). Goal here is to show how to develop your own external server (and how easy it is). The code is not production ready but provides a basis as it already deals with the gRPC plumbing and anything from Gloo configuration to docker image and Kubernetes manifests are also included.

## Features

- JWT validation and claims extraction

## How to build (optional step)

Local build (you need Golang installed): `make`

To build the code and the image (you don't need Golang): `make image`

> The image can also be pulled directly from Docker hub: *`docker pull asaintsever/gloo-oss-extauth`*

## How to deploy

With Gloo already installed in your Kubernetes cluster, deploy `gloo-oss-extauth`:

```sh
kubectl apply -f deploy/gloo-settings.yaml
kubectl apply -f deploy/gloo-oss-extauth.yaml
```

Check `gloo-oss-extauth` is started:

```sh
kubectl -n gloo-system logs <gloo-oss-extauth pod>
```

## Test

Deploy the Echoserver service in your cluster:

```sh
kubectl apply -f samples/echoserver.yaml
```

Then, deploy Gloo VirtualService resource enabling JWT validation through `gloo-oss-extauth`:

```sh
# Before deploying this manifest: update it with proper values for 'jwks_url', 'jwt_issuer' and 'jwt_audience'
kubectl apply -f samples/jwt_validation/vs-remote-jwks.yaml
```

Test by invoking Echoserver with and without a JWT:

```sh
$ curl -i $(glooctl proxy url)/echo
HTTP/1.1 401 Unauthorized
content-length: 46
content-type: text/plain
date: Fri, 02 Oct 2020 12:23:07 GMT
server: envoy

Authorization Header malformed or not provided

$ curl -i -H "Authorization: Bearer <invalid JWT>" $(glooctl proxy url)/echo
HTTP/1.1 403 Forbidden
date: Fri, 02 Oct 2020 12:28:25 GMT
server: envoy
content-length: 0

$ curl -i -H "Authorization: Bearer <your JWT>" $(glooctl proxy url)/echo
HTTP/1.1 200 OK
date: Fri, 02 Oct 2020 12:24:16 GMT
content-type: text/plain
server: envoy
x-envoy-upstream-service-time: 3
transfer-encoding: chunked

...

Request Headers:
        accept=*/*
        authorization=Bearer <your JWT>
        ...
        x-jwt-scope=read:client_grants create:client_grants delete:client_grants update:client_grants read:users update:users delete:users create:users read:users_app_metadata update:users_app_metadata delete:users_app_metadata create:users_app_metadata create:user_tickets read:clients update:clients delete:clients create:clients read:client_keys update:client_keys delete:client_keys create:client_keys read:connections update:connections delete:connections create:connections read:resource_servers update:resource_servers delete:resource_servers create:resource_servers read:device_credentials update:device_credentials delete:device_credentials create:device_credentials read:rules update:rules delete:rules create:rules read:rules_configs update:rules_configs delete:rules_configs read:hooks update:hooks delete:hooks create:hooks read:email_provider update:email_provider delete:email_provider create:email_provider blacklist:tokens read:stats read:tenant_settings update:tenant_settings read:logs read:shields create:shields delete:shields read:anomaly_blocks delete:anomaly_blocks update:triggers read:triggers read:grants delete:grants read:guardian_factors update:guardian_factors read:guardian_enrollments delete:guardian_enrollments create:guardian_enrollment_tickets read:user_idp_tokens create:passwords_checking_job delete:passwords_checking_job read:custom_domains delete:custom_domains create:custom_domains read:email_templates create:email_templates update:email_templates read:mfa_policies update:mfa_policies read:roles create:roles delete:roles update:roles read:prompts update:prompts read:branding update:branding read:log_streams create:log_streams delete:log_streams update:log_streams create:requested_scopes read:requested_scopes delete:requested_scopes update:requested_scopes
        x-jwt-sub=URirPO9Fo1evKMbPwqd8VpvySgOq7aCA@clients
        x-request-id=fc555f12-3340-492a-b285-68e25b78e013

Request Body:
        -no body in request-
```

The provided test VirtualService enables forwarding of the JWT to the upstream service (Echoserver here) and ask for extraction of the scope and sub claims. The configuration can be easily changed to suit your needs.
