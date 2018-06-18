RoadRunner
==========
[![Latest Stable Version](https://poser.pugx.org/spiral/roadrunner/version)](https://packagist.org/packages/spiral/roadrunner)
[![GoDoc](https://godoc.org/github.com/spiral/roadrunner?status.svg)](https://godoc.org/github.com/spiral/roadrunner)
[![Build Status](https://travis-ci.org/spiral/roadrunner.svg?branch=master)](https://travis-ci.org/spiral/roadrunner)
[![Go Report Card](https://goreportcard.com/badge/github.com/spiral/roadrunner)](https://goreportcard.com/report/github.com/spiral/roadrunner)
[![Scrutinizer Code Quality](https://scrutinizer-ci.com/g/spiral/roadrunner/badges/quality-score.png)](https://scrutinizer-ci.com/g/spiral/roadrunner/?branch=master)
[![Codecov](https://codecov.io/gh/spiral/roadrunner/branch/master/graph/badge.svg)](https://codecov.io/gh/spiral/roadrunner/)

RoadRunner is an open source (MIT licensed), high-performance PSR-7 PHP application server, load balancer and process manager.
It supports service model with ability to extend it's functionality on a project basis.

Features:
--------
- PSR-7 HTTP server (file uploads, error handling, static files, hot reload, middlewares, event listeners)
- extendable service model (plus PHP compatible RPC server)
- no external PHP dependencies, drop-in (based on [Goridge](https://github.com/spiral/goridge))
- load balancer, process manager and task pipeline
- frontend agnostic (queue, REST, PSR-7, async php, etc)
- works over TCP, unix sockets and standard pipes
- automatic worker replacement and safe PHP process destruction
- worker lifecycle management (create/allocate/destroy timeouts)
- payload context and body
- control over max jobs per worker
- protocol, worker and job level error management (including PHP errors)
- memory leak failswitch
- very fast (~250k rpc calls per second on Ryzen 1700X over 16 threads)
- works on Windows

Getting Started:
--------

#### Downloading RoadRunner
The easiest way to get the latest RoadRunner version is to use one of the pre-built release binaries which are available for
OSX, Linux, FreeBSD, and Windows. Instructions for using these binaries are on the GitHub [releases page](https://github.com/spiral/roadrunner/releases).

#### Building RoadRunner
RoadRunner can be compiled on Linux, OSX, Windows and other 64 bit environments as the only requirement is Go 1.8+ itself.

To build:

```
$ make
```

To test:

```
$ make test
```

Using RoadRunner:
--------

In order to use RoadRunner you only have to place `.rr.yaml` file in a root of your php project:

```yaml
# rpc bus allows php application and external clients to talk to rr services.
rpc:
  # enable rpc server
  enable: true

  # rpc connection DSN. Supported TCP and Unix sockets.
  listen: tcp://127.0.0.1:6001

# http service configuration.
http:
  # set to false to disable http server.
  enable:     true

  # http host to listen.
  address:    0.0.0.0:8080

  # max POST request size, including file uploads in MB.
  maxRequest: 200

  # file upload configuration.
  uploads:
    # list of file extensions which are forbidden for uploading.
    forbid: [".php", ".exe", ".bat"]

  # http worker pool configuration.
  workers:
    # php worker command.
    command:  "php psr-worker.php pipes"

    # connection method (pipes, tcp://:9000, unix://socket.unix).
    relay:    "pipes"

    # worker pool configuration.
    pool:
      # number of workers to be serving.
      numWorkers: 4

      # maximum jobs per worker, 0 - unlimited.
      maxJobs:  0

      # for how long pool should attempt to allocate free worker (request timeout). Nanoseconds atm. (60s)
      allocateTimeout: 60000000000

      # amount of time given to worker to gracefully destruct itself. Nanoseconds atm. (30s)
      destroyTimeout:  30000000000

# static file serving.
static:
  # serve http static files
  enable:  true

  # root directory for static file (http would not serve .php and .htaccess files).
  dir:   "public"

  # list of extensions to forbid for serving.
  forbid: [".php", ".htaccess"]
```

> You can use json or any config type supported by `spf13/viper`.

Where `psr-worker.php`:

```php
$psr7 = new RoadRunner\PSR7Client(new RoadRunner\Worker($relay));

while ($req = $psr7->acceptRequest()) {
    try {
        $resp = new \Zend\Diactoros\Response()
        $resp->getBody()->write("hello world");

        $psr7->respond($resp);
    } catch (\Throwable $e) {
        $psr7->getWorker()->error((string)$e);
    }
}
```

> Check how to init relay [here](./php-src/tests/client.php).

Working with RoadRunner service:
--------

RoadRunner application can be started by calling simple command from the root of your PHP application.

```
$ rr serve -v
```

You can also run RR in debug mode to view all incoming requests.

```
$ rr serve -d -v
```

You can force RR service to reload it's http workers.

```
$ rr http:reset
```

> You can attach this command as file watcher in your IDE.

To view status of all active workers in interactive mode.

```
$ rr http:workers -i
```

```
+---------+-----------+---------+---------+--------------------+
|   PID   |  STATUS   |  EXECS  | MEMORY  |      CREATED       |
+---------+-----------+---------+---------+--------------------+
|    9440 | ready     |  42,320 | 31 MB   | 22 minutes ago     |
|    9447 | ready     |  42,329 | 31 MB   | 22 minutes ago     |
|    9454 | ready     |  42,306 | 31 MB   | 22 minutes ago     |
|    9461 | ready     |  42,316 | 31 MB   | 22 minutes ago     |
+---------+-----------+---------+---------+--------------------+
```

Standalone Usage:
--------
You can also use RoadRunner as library in order to drive your application without any additional protocol at top of it.

```go
srv := NewServer(
    &ServerConfig{
        Command: "php client.php echo pipes",
        Relay:   "pipes",
        Pool: &Config{
            NumWorkers:      int64(runtime.NumCPU()),
            AllocateTimeout: time.Second,
            DestroyTimeout:  time.Second,
        },
    })
defer srv.Stop()

srv.Start()

res, err := srv.Exec(&Payload{Body: []byte("hello")})
```

```php
<?php
/**
 * @var Goridge\RelayInterface $relay
 */

use Spiral\Goridge;
use Spiral\RoadRunner;

$rr = new RoadRunner\Worker($relay);

while ($body = $rr->receive($context)) {
    try {
        $rr->send((string)$body, (string)$context);
    } catch (\Throwable $e) {
        $rr->error((string)$e);
    }
}
```
> Check how to init relay [here](./php-src/tests/client.php).

You can find more examples in tests and `php-src` directory.

License:
--------
The MIT License (MIT). Please see [`LICENSE`](./LICENSE) for more information.
