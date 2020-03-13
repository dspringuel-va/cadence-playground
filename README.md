#Start Cadence server

To start a Cadence server instance locally, run the following

```
docker-compose up
```

*Make sure that the Docker daemon is running.*

The `docker-compose.yaml` was copied from [Github Cadence Docker directory](https://github.com/uber/cadence/tree/master/docker)

It will start Cadence itself (service + web page), along with a Cassandra instance and a Graphite-statsd instance.

To open the Cadence web page, go to `localhost:8088`.

#Register Cadence domain
 
Cadence has a cli docker image hosted on DockerHub.

It can operate on domains, workflows, tasklists, and run various admin commands.

To register a domain, run 

```
docker run --rm ubercadence/cli:master --address host.docker.internal:7933 --domain test-domain domain register
```

A Cadence domain acts like a namespace for workflows.

#Start the worker server

To start the worker server (i.e. the go server in this repository), run 

```
go run main.go
```

It'll create a workflow worker on the `test-domain` domain, and `test-task-list` list.

The server also listen to `:1234`, where the endpoint `/start-workflow` can be called to start the test workflow.

The code in main is meant to be a playground to play around and test Cadence.