# Hotel

An ssh server that puts the users in a container and mounts their home directory.

This project was inspired by bad security practices I've seen out in the wild.

It uses PAM to authenticate users so if the user exists on the system, it'll work with hotel.

## dependencies

- pam libraries(`libpam0g-dev` on debian based distros)
- docker

## Configuration

`HOTEL_PORT` is the only thing to configure. It is the port on which the server runs on. If not given, `2222` will be used.

`HOTEL_HOST_KEY_PATH` is the path to the host key. It is the ssh host key for all encrypted messages. If not given, one will be generated.

## Running

You can use

```console
go run main.go
```

or you can run the prebuilt `./hotel`

```console
./hotel
```

## Connecting

Connect like you would a normal ssh server, it has the exact same experience
