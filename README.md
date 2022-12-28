# Compute Engine - Cloud Run Reverse Proxy
A container based reverse proxy for Cloud Run.

## Motivation
- Sometimes you need a reverse proxy for Cloud Run to meet specific customer requirements such as private Dialogflow CX -> Cloud Run

## Features
- Provides basic Health Checks at /health/check (HTTPS) for load balancers.  Returns 200, Status OK when the service is up.
- Dynamically creates self-signed certificates at runtime. Certificates are required to serve TLS and are also required by GCP's HTTP(S) Internal Load Balancers for more details please see https://cloud.google.com/load-balancing/docs/ssl-certificates/encryption-to-the-backends#encryption-to-backends
- Dynamically creates and applies new OIDC tokens with the Cloud Run Service as the audience.
- Easy deployment: build this container using gcloud builds submit and then use COS to deploy a single host (don't to this!) or create a container based instance template and deploy as a MIG.  

## Variables and Parameters
When creating an instance template, you can deploy a container and specify environment variables there.  UPSTREAM and PRINCIPAL are REQUIRED.

- PORT: Optional with a default 443; this defines the port the container will be listening on.  Supplying a new port is up to the operator
- UPSTREAM: Required. The upstream Cloud Run service's URL that you're proxying.
- PRINCIPAL: Required.  The principal (email) of the invoking resource.  In the case of Dialogflow, it's Dialogflow's Service Agent service account.

## Deployment
Clone this repo and build it in your GCP project
```shell
git clone https://github.com/YvanJAquino/gce-run-proxy
cd gce-run-proxy
```
Build the container image with Cloud Build
```shell
gcloud builds submit
```

Use the container image however you'd like! 