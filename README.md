# Summary

`golang` based API leveraging [gorm](https://github.com/jinzhu/gorm) and [gin](https://github.com/gin-gonic/gin) libraries with postgres.

## Tools used

* [golang 1.10.1](https://golang.org/)
* [docker](https://www.docker.com/)
* [docker-compose](https://docs.docker.com/compose/)
* [postgres 9.6](https://hub.docker.com/_/postgres/)
* [kompose](https://github.com/kubernetes/kompose)
* [minikube](https://github.com/kubernetes/minikube)

## Quickstart

`docker-compose` is used for TDD workflow, and `minikube` is used for ensuring dev -> prod compatibility.

### Local

#### Run the API

```
git clone https://github.com/mblhaunted/gobooks.git
docker-compose up
```

#### Run tests

```
git clone https://github.com/mblhaunted/gobooks.git
docker-compose api run go test

```

#### Run the API (minikube)


```
minikube delete
minikube start
helm init
git clone https://github.com/mblhaunted/gobooks.git
helm install kubernetes/helm/gobooks
```

### Kubernetes

```
git clone https://github.com/mblhaunted/gobooks.git
helm install kubernetes/helm/gobooks
```


## Code

All application logic is in `main.go` and all test logic is in `main_test.go`.

All code exists in `package main`.

_Note: This breaks `go doc`. Docs can be generated, but it's a little bit of a hack._
_See the issue [here](https://github.com/golang/go/issues/5727)._

### Dependencies

Dependencies were created/managed using `dep`. 

The `Gopkg.yml` file, along with the `vendor` folder are included in this repo for convenience.

`dep ensure` should build a `Gopkg.lock` and pull the latest source.

## Bootstrapping

### minikube

`minikube delete && minikube start && eval $(minikube docker-env) && docker-compose build --no-cache --force-rm`

Note, you can revert the docker environment by issuing the following command.

`eval "$(docker-machine env -u)"`

### docker-compose

`docker-compose build --no-cache --force-rm`

## Local Usage

### Running tests and TDD

The `docker-compose.yml` is designed to leverage a volume, meaning you ought to be able to edit your code/tests and quickly get feedback.

- Create/modify `main_test.go`, `main.go`, etc
- _Optional_: Bootstrap env: `docker-compose down --rmi all --remove-orphans`
- Run `docker-compose run api go test`

### Running the API (docker-compose)

- `docker-compose up`

### Running the API (minikube) - *READ CAREFULLY*

### Kompose

_Unfortunately, there's a kompose bug getting this working on OSX. This approach has not been tested._
_See the bug [here](https://github.com/kubernetes/kompose/issues/911)._

- `minikube delete`
- `minikube start`
- `eval $(minikube docker-env)`
- `kompose up`

Using this approach would require modifying the `docker-compose.yml` to deal with volumes.

### Helm

From the root of the project, this beastly command will deploy the API to a local `minikube`.

`$ minikube delete && minikube start && eval $(minikube docker-env) && docker-compose build --no-cache --force-rm && helm init && cd kubernetes/helm/ && sleep 90 && helm install gobooks && minikube dashboard && cd ../..`

### Stopping the API (docker-compose)

- `docker-compose down --rmi all --remove-orphans`

### Stopping the API (helm)

Destroy the `minikube` cluster.

- `minikube delete`

Less aggressively ...

- `minikube stop`

## "Prod Usage"

The kubernetes content provided is an attempt to generate generic content that should work without much modification.

`kompose` was used to generate all kubernetes content for this project.

* Download and install [Kompose](https://github.com/kubernetes/kompose)

### Helm

A helm chart has been provided in `./kubernetes/helm/gobooks`.

* `cd kubernetes/helm && helm install gobooks`

_Note, the volume has been manually removed post `kompose convert -c`._

### kubectl

Files have been provided via `kompose convert` for using `kubectl apply` to deploy the API.

The files are in the `kubernetes/manual` folder.

## Notes

### Nice-to-haves

#### Binary

A workflow to ship the compiled binary to a image for containers instead of using `go run main.go`.

- Rework docker-compose, or create Dockerfiles, etc

#### More tests

I consider the tests a first draft. I would normally build out more of a component test suite.

#### Testing with other SQL databases

##### Experimenting with other datastores, such as ElasticSearch

#### Monitoring and alerting

I've done a lot in this area, and would have enjoyed building out application level monitoring.

#### A/B Deployment/Testing

I've designed daemons in the past to operate in "test mode" with prod data, which I find super useful for testing and deployment purposes.

#### A cloud infrastructure provider

I can dream, right? :)
