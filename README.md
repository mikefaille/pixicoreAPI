# pixicoreAPI

## How to run the API

- Install [Go](https://nats.io/documentation/tutorials/go-install/).

- Git clone the repo into `$GOPATH/src/github.com/ClubCedille/`.

- Enter the `pixicoreAPI` directory.

- Install [Dep](https://golang.github.io/dep/docs/installation.html), a program that will manage Go dependencies.

- Add `dep` to your PATH like this: `export PATH=$PATH:$GOPATH/bin`.

- Now you can install all the package dependencies with `dep ensure`.

- `go test ./... && go build ./cmd/pixicoreAPI && ./pixicoreAPI` will run the tests, build the program and run it.

## Run with vagrant

- Install Vagrant and VirtualBox

- Create a hostonly adapter with VBoxManage with `VBoxManage hostonlyif create`

- Copy the returned name and replaced in the Vagrantfile where `vboxnet0` is used;

- Run `vagrant up` or separately with (`vagrant up master`), (`vagrant up vboxNode1`), (`vagrant up vboxNode2`).

- ssh into the master with `vagrant ssh master`

#### Using Docker

- `docker build -t pixicoreapi .`

- `docker run -d -p 3000:3000 pixicoreapi`

## Usage

- Change the IP address for each server in the `servers-config.yaml` file

- You can run `curl -i http://localhost:3000/v1/install/SERVER_MAC_ADDRESS`: this will collect info and install coreOS for the server

- You can run `curl -i http://localhost:3000/v1/all/`: this will collect info  and install coreOS for each server

## API Endpoints

### `GET v1/boot/:macAddress`

- Used by pixicore to get PXE config and boot each server (each server have a IP address assigned).

### `GET v1/install/:macAddress`

- Gets information (cores, RAM, etc) from the server using its macAddress as ID and install coresOS.

### `GET v1/all`

- Gets information (cores, RAM, etc) from each the server using its macAddress as ID and install coresOS for each one.

### `GET v1/servers`

- Show information about all the registered servers.