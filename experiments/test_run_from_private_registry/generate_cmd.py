private_registry = "202.114.10.146:9999/"
apppath = "/home/seveir/Desktop/experiments/test_run_from_private_registry"

def cmdGenerate(repo, tag):
    if repo == "alpine":
        cmd = {
            "image": private_registry+"alpine:"+tag, 
            "environment": [
            ],
            "ports": {
            }, 
            "working_dir": "",
            "volumes": {
            },
            "command": "echo hello", 
            "waitline": "hello", 
        }
    elif repo == "amazonlinux":
        cmd = {
            "image": private_registry+"amazonlinux:"+tag, 
            "environment": [
            ],
            "ports": {
            }, 
            "working_dir": "",
            "volumes": {
            },
            "command": "echo hello", 
            "waitline": "hello", 
        }
    elif repo == "buildpack-deps":
        cmd = {
            "image": private_registry+"buildpack-deps:"+tag, 
            "environment": [
            ],
            "ports": {
            }, 
            "working_dir": "",
            "volumes": {
            },
            "command": "echo hello", 
            "waitline": "hello", 
        }
    elif repo == "busybox":
        cmd = {
            "image": private_registry+"busybox:"+tag, 
            "environment": [
            ],
            "ports": {
            }, 
            "working_dir": "",
            "volumes": {
            },
            "command": "echo hello", 
            "waitline": "hello", 
        }
    elif repo == "centos":
        cmd = {
            "image": private_registry+"centos:"+tag, 
            "environment": [
            ],
            "ports": {
            }, 
            "working_dir": "",
            "volumes": {
            },
            "command": "echo hello", 
            "waitline": "hello", 
        }
    elif repo == "debian":
        cmd = {
            "image": private_registry+"debian:"+tag, 
            "environment": [
            ],
            "ports": {
            }, 
            "working_dir": "",
            "volumes": {
            },
            "command": "echo hello", 
            "waitline": "hello", 
        }
    elif repo == "fedora":
        cmd = {
            "image": private_registry+"fedora:"+tag, 
            "environment": [
            ],
            "ports": {
            }, 
            "working_dir": "",
            "volumes": {
            },
            "command": "echo hello", 
            "waitline": "hello", 
        }
    elif repo == "ubuntu":
        cmd = {
            "image": private_registry+"ubuntu:"+tag, 
            "environment": [
            ],
            "ports": {
            }, 
            "working_dir": "",
            "volumes": {
            },
            "command": "echo hello", 
            "waitline": "hello", 
        }
    elif repo == "cassandra":
        cmd = {
            "image": private_registry+"cassandra:"+tag, 
            "environment": [
            ],
            "ports": {
            }, 
            "working_dir": "",
            "volumes": {
            },
            "command": "", 
            "waitline": "on /0.0.0.0:9042", 
        }
    elif repo == "elasticsearch":
        cmd = {
            "image": private_registry+"elasticsearch:"+tag, 
            "environment": [
                "discovery.type=single-node",
            ],
            "ports": {
                "9200/tcp": 9200,
                "9300/tcp": 9300,
            }, 
            "working_dir": "",
            "volumes": {
            },
            "command": "", 
            "waitline": "started", 
        }
    elif repo == "influxdb":
        cmd = {
            "image": private_registry+"influxdb:"+tag, 
            "environment": [
            ],
            "ports": {
                "8086/tcp": 8086,
            }, 
            "working_dir": "",
            "volumes": {
            },
            "command": "", 
            "waitline": "Sending usage statistics to usage", 
        }
    elif repo == "mariadb":
        cmd = {
            "image": private_registry+"mariadb:"+tag, 
            "environment": [
                "MYSQL_ROOT_PASSWORD=my-secret-pw",
            ],
            "ports": {
            }, 
            "working_dir": "",
            "volumes": {
            },
            "command": "", 
            "waitline": "ready for connections", 
        }
    # need to optimize(OK)
    elif repo == "memcached":
        cmd = {
            "image": private_registry+"memcached:"+tag, 
            "environment": [
            ],
            "ports": {
                "11211/tcp": 8080,
            }, 
            "working_dir": "",
            "volumes": {
            },
            "command": "", 
            "waitline": "", 
        }
    elif repo == "mongo":
        cmd = {
            "image": private_registry+"mongo:"+tag, 
            "environment": [
            ],
            "ports": {
            }, 
            "working_dir": "",
            "volumes": {
            },
            "command": "", 
            "waitline": "waiting for connections", 
        }
    elif repo == "mysql":
        cmd = {
            "image": private_registry+"mysql:"+tag, 
            "environment": [
                "MYSQL_ROOT_PASSWORD=my-secret-pw",
            ],
            "ports": {
            }, 
            "working_dir": "",
            "volumes": {
            },
            "command": "", 
            "waitline": "ready for connections", 
        }
    elif repo == "neo4j":
        cmd = {
            "image": private_registry+"neo4j:"+tag, 
            "environment": [
                "MYSQL_ROOT_PASSWORD=my-secret-pw",
                "NEO4J_ACCEPT_LICENSE_AGREEMENT=yes",
            ],
            "ports": {
            }, 
            "working_dir": "",
            "volumes": {
            },
            "command": "", 
            "waitline": "Started", 
        }
    elif repo == "postgres":
        cmd = {
            "image": private_registry+"postgres:"+tag, 
            "environment": [
                "POSTGRES_PASSWORD=mysecretpassword",
            ],
            "ports": {
            }, 
            "working_dir": "",
            "volumes": {
            },
            "command": "", 
            "waitline": "database system is ready to accept connections", 
        }
    elif repo == "redis":
        cmd = {
            "image": private_registry+"redis:"+tag, 
            "environment": [
            ],
            "ports": {
            }, 
            "working_dir": "",
            "volumes": {
            },
            "command": "", 
            "waitline": "Ready to accept connections", 
        }
    elif repo == "rethinkdb":
        cmd = {
            "image": private_registry+"rethinkdb:"+tag, 
            "environment": [
            ],
            "ports": {
            }, 
            "working_dir": "",
            "volumes": {
            },
            "command": "", 
            "waitline": "Server ready", 
        }
    # these are contaienrs that can not startup alone(they need other containers to help)
    elif repo == "consul":
        cmd = {}
    elif repo == "jenkins":
        cmd = {}
    elif repo == "maven":
        cmd = {}
    elif repo == "kibana":
        cmd = {}
    elif repo == "sonarqube":
        cmd = {}
    elif repo == "telegraf":
        cmd = {}
    elif repo == "sentry":
        cmd = {}
    # need to optimize(OK)
    elif repo == "docker":
        cmd = {
            "image": private_registry+"docker:"+tag, 
            "environment": [
            ],
            "ports": {
            }, 
            "working_dir": "",
            "volumes": {
            },
            "command": "docker version", 
            "waitline": "Client", 
        }
    elif repo == "drupal":
        cmd = {
            "image": private_registry+"drupal:"+tag, 
            "environment": [
            ],
            "ports": {
                "80/tcp":8080,
            }, 
            "working_dir": "",
            "volumes": {
            },
            "command": "", 
            "waitline": "pid 1", 
        }
    elif repo == "gradle":
        cmd = {
            "image": private_registry+"gradle:"+tag, 
            "environment": [
            ],
            "ports": {
            }, 
            "working_dir": "",
            "volumes": {
            },
            "command": "", 
            "waitline": "BUILD SUCCESSFUL", 
        }
    elif repo == "kong":
        cmd = {}
    elif repo == "nextcloud":
        if tag.find("apache") >= 0:
            cmd = {
                "image": private_registry+"nextcloud:"+tag, 
                "environment": [
                ],
                "ports": {
                    "80/tcp":8080,
                }, 
                "working_dir": "",
                "volumes": {
                },
                "command": "", 
                "waitline": "apache2 -D FOREGROUND", 
            }
        else:
            cmd = {
                "image": private_registry+"nextcloud:"+tag, 
                "environment": [
                ],
                "ports": {
                    "80/tcp":8080,
                }, 
                "working_dir": "",
                "volumes": {
                },
                "command": "", 
                "waitline": "ready to handle connections", 
            }
    elif repo == "owncloud":
        if tag.find("apache") >= 0:
            cmd = {
                "image": private_registry+"owncloud:"+tag, 
                "environment": [
                ],
                "ports": {
                    "80/tcp":8080,
                }, 
                "working_dir": "",
                "volumes": {
                },
                "command": "", 
                "waitline": "apache2 -D FOREGROUND", 
            }
        else:
            cmd = {
                "image": private_registry+"owncloud:"+tag, 
                "environment": [
                ],
                "ports": {
                    "80/tcp":8080,
                }, 
                "working_dir": "",
                "volumes": {
                },
                "command": "", 
                "waitline": "ready to handle connections", 
            }
    elif repo == "eclipse-mosquitto":
        cmd = {
            "image": private_registry+"eclipse-mosquitto:"+tag, 
            "environment": [
            ],
            "ports": {
                "1883/tcp":1883,
                "9001/tcp":9001,
            }, 
            "working_dir": "",
            "volumes": {
            },
            "command": "", 
            "waitline": "listen socket on",
        }
    elif repo == "ghost":
        cmd = {
            "image": private_registry+"ghost:"+tag, 
            "environment": [
            ],
            "ports": {
                "3001/tcp":2368,
            }, 
            "working_dir": "",
            "volumes": {
            },
            "command": "", 
            "waitline": "localhost:2368",
        }
    elif repo == "nats":
        cmd = {
            "image": private_registry+"nats:"+tag, 
            "environment": [
            ],
            "ports": {
            }, 
            "working_dir": "",
            "volumes": {
            },
            "command": "", 
            "waitline": "Server is ready",
        }
    elif repo == "registry":
        cmd = {
            "image": private_registry+"registry:"+tag, 
            "environment": [
            ],
            "ports": {
            }, 
            "working_dir": "",
            "volumes": {
            },
            "command": "", 
            "waitline": "listening on",
        }
    elif repo == "solr":
        cmd = {
            "image": private_registry+"solr:"+tag, 
            "environment": [
            ],
            "ports": {
            }, 
            "working_dir": "",
            "volumes": {
            },
            "command": "", 
            "waitline": "Server Started",
        }
    elif repo == "vault":
        cmd = {
            "image": private_registry+"vault:"+tag, 
            "environment": [
            ],
            "ports": {
            }, 
            "working_dir": "",
            "volumes": {
            },
            "command": "", 
            "waitline": "Vault server started",
        }
    elif repo == "wordpress":
        if tag.find("apache") >= 0:
            cmd = {
                "image": private_registry+"wordpress:"+tag, 
                "environment": [
                ],
                "ports": {
                }, 
                "working_dir": "",
                "volumes": {
                },
                "command": "", 
                "waitline": "apache2 -D FOREGROUND",
            }
        elif tag.find("cli") >= 0:
            cmd = {}
        else:
            cmd = {
                "image": private_registry+"wordpress:"+tag, 
                "environment": [
                ],
                "ports": {
                }, 
                "working_dir": "",
                "volumes": {
                },
                "command": "", 
                "waitline": "ready to handle connections",
            }
    elif repo == "logstash":
        cmd = {
            "image": private_registry+"logstash:"+tag, 
            "environment": [
            ],
            "ports": {
            }, 
            "working_dir": "",
            "volumes": {
            },
            "command": "", 
            "waitline": "Successfully started Logstash",
        }
    elif repo == "rabbitmq":
        cmd = {
            "image": private_registry+"rabbitmq:"+tag, 
            "environment": [
            ],
            "ports": {
            }, 
            "working_dir": "",
            "volumes": {
            },
            "command": "", 
            "waitline": "Server startup complete",
        }
    # language images need to be optimized(OK)
    elif repo == "golang":
        cmd = {
            "image": private_registry+"golang:"+tag, 
            "environment": [
            ],
            "ports": {
            }, 
            "working_dir": "",
            "volumes": {
                apppath: {"bind": "/tmp", "mode": "rw"},
            },
            "command": "go run /tmp/golang/hello.go", 
            "waitline": "hello",
        }
    elif repo == "groovy":
        cmd = {
            "image": private_registry+"groovy:"+tag, 
            "environment": [
            ],
            "ports": {
            }, 
            "working_dir": "",
            "volumes": {
                apppath: {"bind": "/tmp", "mode": "rw"},
            },
            "command": "groovy /tmp/groovy/hello.groovy", 
            "waitline": "hello",
        }
    elif repo == "java":
        if tag.find("jre") >= 0:
            cmd = {
                "image": private_registry+"java:"+tag, 
                "environment": [
                ],
                "ports": {
                }, 
                "working_dir": "",
                "volumes": {
                    apppath: {"bind": "/tmp", "mode": "rw"},
                },
                "command": "java /tmp/java/hello.java", 
                "waitline": "hello",
            }
        else:
            cmd = {
                "image": private_registry+"java:"+tag, 
                "environment": [
                ],
                "ports": {
                }, 
                "working_dir": "",
                "volumes": {
                    apppath: {"bind": "/tmp", "mode": "rw"},
                },
                "command": "java /tmp/java/hello.java", 
                "waitline": "hello",
            }
    elif repo == "jruby":
        cmd = {
            "image": private_registry+"jruby:"+tag, 
            "environment": [
            ],
            "ports": {
            }, 
            "working_dir": "",
            "volumes": {
                apppath: {"bind": "/tmp", "mode": "rw"},
            },
            "command": "jruby /tmp/jruby/hello.rb", 
            "waitline": "hello",
        }
    elif repo == "openjdk":
        if tag.find("jre") >= 0:
            cmd = {
                "image": private_registry+"openjdk:"+tag, 
                "environment": [
                ],
                "ports": {
                }, 
                "working_dir": "",
                "volumes": {
                    apppath: {"bind": "/tmp", "mode": "rw"},
                },
                "command": "java /tmp/java/hello.java", 
                "waitline": "hello",
            }
        else:
            cmd = {
                "image": private_registry+"openjdk:"+tag, 
                "environment": [
                ],
                "ports": {
                }, 
                "working_dir": "",
                "volumes": {
                    apppath: {"bind": "/tmp", "mode": "rw"},
                },
                "command": "java /tmp/java/hello.java", 
                "waitline": "hello",
            }
    elif repo == "perl":
        cmd = {
            "image": private_registry+"perl:"+tag, 
            "environment": [
            ],
            "ports": {
            }, 
            "working_dir": "",
            "volumes": {
                apppath: {"bind": "/tmp", "mode": "rw"},
            },
            "command": "perl /tmp/perl/hello.pl", 
            "waitline": "hello",
        }
    elif repo == "php":
        cmd = {
            "image": private_registry+"php:"+tag, 
            "environment": [
            ],
            "ports": {
            }, 
            "working_dir": "",
            "volumes": {
                apppath: {"bind": "/tmp", "mode": "rw"},
            },
            "command": "php /tmp/php/hello.php", 
            "waitline": "hello",
        }
    elif repo == "python":
        cmd = {
            "image": private_registry+"python:"+tag, 
            "environment": [
            ],
            "ports": {
            }, 
            "working_dir": "",
            "volumes": {
                apppath: {"bind": "/tmp", "mode": "rw"},
            },
            "command": "python /tmp/python/hello.py", 
            "waitline": "hello",
        }
    elif repo == "ruby":
        cmd = {
            "image": private_registry+"ruby:"+tag, 
            "environment": [
            ],
            "ports": {
            }, 
            "working_dir": "",
            "volumes": {
                apppath: {"bind": "/tmp", "mode": "rw"},
            },
            "command": "ruby /tmp/ruby/hello.rb", 
            "waitline": "hello",
        }
    elif repo == "hello-world":
        cmd = {
            "image": private_registry+"hello-world:"+tag, 
            "environment": [
            ],
            "ports": {
            }, 
            "working_dir": "",
            "volumes": {
            },
            "command": "", 
            "waitline": "Hello from Docker!",
        }
    elif repo == "rocket.chat":
        cmd = {}
    elif repo == "haproxy":
        cmd = {}
    elif repo == "httpd":
        cmd = {
            "image": private_registry+"httpd:"+tag, 
            "environment": [
            ],
            "ports": {
            }, 
            "working_dir": "",
            "volumes": {
            },
            "command": "", 
            "waitline": "httpd -D FOREGROUND",
        }
    # need to be optimized (OK)
    elif repo == "nginx":
        cmd = {
            "image": private_registry+"nginx:"+tag, 
            "environment": [
            ],
            "ports": {
                "80/tcp": 8080,
            }, 
            "working_dir": "",
            "volumes": {
            },
            "command": "", 
            "waitline": "",
        }
    # need to be optimized(OK)
    elif repo == "node":
        cmd = {
            "image": private_registry+"node:"+tag, 
            "environment": [
            ],
            "ports": {
                "80/tcp": 8080,
            }, 
            "working_dir": "",
            "volumes": {
                apppath: {"bind": "/tmp", "mode": "rw"},
            },
            "command": "node /tmp/node/index.js", 
            "waitline": "",
        }
    elif repo == "percona":
        cmd = {
            "image": private_registry+"percona:"+tag, 
            "environment": [
                "MYSQL_ROOT_PASSWORD=my-secret-pw",
            ],
            "ports": {
            }, 
            "working_dir": "",
            "volumes": {
            },
            "command": "", 
            "waitline": "ready for connections",
        }
    elif repo == "swarm":
        cmd = {
            "image": private_registry+"swarm:"+tag, 
            "environment": [
            ],
            "ports": {
            }, 
            "working_dir": "",
            "volumes": {
            },
            "command": "create", 
            "waitline": "Token based discovery",
        }
    elif repo == "tomcat":
        cmd = {
            "image": private_registry+"tomcat:"+tag, 
            "environment": [
            ],
            "ports": {
            }, 
            "working_dir": "",
            "volumes": {
            },
            "command": "", 
            "waitline": "Server startup",
        }
    # need to be optimized (OK)
    elif repo == "traefik":
        cmd = {
            "image": private_registry+"traefik:"+tag, 
            "environment": [
            ],
            "ports": {
                "8080/tcp": 8080,
                "80/tcp": 80,
            }, 
            "working_dir": "",
            "volumes": {
                apppath+"/traefik.toml":{"bind": "/etc/traefik/traefik.toml", "mode": "rw"},
                "/var/run/docker.sock": {"bind": "/var/run/docker.sock", "mode": "rw"},
            },
            "command": "", 
            "waitline": "",
        }

    return cmd