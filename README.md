# DevicesAPI

## Getting started

### What you will need

- Go v1.25+
- Docker
- Kubernetes cluster (tested with minikube)
    - Docker compose also works and compose file is also included in this repository

### Compose route

- Clone this repository
- Build the api image with `docker build -t devices-api -f Dockerfile .` and run `docker compose up`
- If `bash` is installed in your system, you can also execute `build.sh` to build and run composer

If minikube is installed you can also `apply` all .yaml files inside the kubernetes folder `kubectl apply -f filename.yaml`


## Proposal
develop a REST API capable of persisting and managing device resources.

### Device Domain
In PostgreSQL

- ID `UUID`
- Name `varchar(250)`
- Brand `varchar(250)`
- CreatedAt `TIMESTAMPTZ`
- State: `ENUM`
    - Available
    - In-Use
    - Inactive

### Supported Functionalities
- Create a new device. `POST`
- Fully and/or partially update an existing device. `PUT`
- Fetch a single device. `GET`
- Fetch all devices. `GET`
- Fetch devices by brand. `GET`
- Fetch devices by state. `GET`
- Delete a single device. `DELETE`

### Domain Validations
- Creation time cannot be updated.
- Name and brand properties cannot be updated if the device is in use.
- In use devices cannot be deleted.

## Test coverage
`go test $(go list ./... | grep -v '/docs') -coverprofile=cover.out && go tool cover -html=cover.out -o cover.html`