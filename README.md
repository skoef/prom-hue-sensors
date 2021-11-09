# Prometheus exporter for Hue Sensors

This program allows you to gather generic metrics on all your Philips Hue sensors with Prometheus.

## Installation
Installation of `prom-hue-sensors` can be done by running this command:

```
go install github.com/skoef/prom-hue-sensors@latest
```

Alternatively, a Docker container is available as well:

```
docker run rschoof/prom-hue-sensors
```

## Registering a new user
To be able to gather metrics on the sensors, a user should be registered with Hue. This is quite simple actually. Run the following command:

```
prom-hue-sensors -register
```

and press the big button on the front of the Hue bridge within one minute. If it takes more than one minute to walk over to your Hue bridge and press the button, increase the timeout with `-register-timeout 5m` for instance.

The registration command can directly store the user key in a file by passing `-user-key-path /path/to/file` to the register command 

## Running 
When a user key was registered, there are several ways you can invoke `prom-hue-sensors`:

- Pass the user key from the commandline like this: 

```
prom-hue-sensors -user s3cr3tus3rk3y
```

- Pass the user key as an environment variable like this:

```
HUE_USER=s3cr3tus3rk3y prom-hue-sensors
```

- Pass the user key stored in a file like this:
 
```
prom-hue-sensors -user-key-path /path/to/file
```

## Kubernetes
The container is built for ARM64 (as well as AMD64 and ARM7) so suitable for running on kubernetes on a Raspberry Pi for instance. In the `manifests/` directory are 2 examples of deployments: `deployment-key.yaml` and `deployment-register.yaml` for running the container with a preregistered key or registering a key on the spot respectively.