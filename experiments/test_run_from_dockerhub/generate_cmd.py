private_registry = ""

def cmdGenerate(repo, tag):
    if repo == "alpine":
        cmd = ["docker run --name hello "+private_registry+"alpine:"+tag+" echo hello", "hello"]
    elif repo == "amazonlinux":
        cmd = ["docker run --name hello "+private_registry+"amazonlinux:"+tag+" echo hello", "hello"]
    elif repo == "buildpack-deps":
        cmd = ["docker run --name hello "+private_registry+"buildpack-deps:"+tag+" echo hello", "hello"]
    elif repo == "busybox":
        cmd = ["docker run --name hello "+private_registry+"busybox:"+tag+" echo hello", "hello"]
    elif repo == "centos":
        cmd = ["docker run --name hello "+private_registry+"centos:"+tag+" echo hello", "hello"]
    elif repo == "debian":
        cmd = ["docker run --name hello "+private_registry+"debian:"+tag+" echo hello", "hello"]
    elif repo == "fedora":
        cmd = ["docker run --name hello "+private_registry+"fedora:"+tag+" echo hello", "hello"]
    elif repo == "ubuntu":
        cmd = ["docker run --name hello "+private_registry+"ubuntu:"+tag+" echo hello", "hello"]
    elif repo == "cassandra":
        cmd = ["docker run --name hello "+private_registry+"cassandra:"+tag, "Listening for thrift clients"]
    elif repo == "elasticsearch":
        cmd = ["docker run --name hello -p 9200:9200 -p 9300:9300 -e \"discovery.type=single-node\" "+private_registry+"elasticsearch:"+tag, "o.e.l.LicenseService"]
    elif repo == "influxdb":
        cmd = ["docker run --name hello -p 8086:8086 "+private_registry+"influxdb:"+tag, "Sending usage statistics to usage."]
    elif repo == "mariadb":
        cmd = ["docker run --name hello -e MYSQL_ROOT_PASSWORD=my-secret-pw "+private_registry+"mariadb:"+tag, "ready for connections"]
    elif repo == "memcached":
        cmd = ["docker run --name hello "+private_registry+"memcached:"+tag+" echo hello", "hello"]
    elif repo == "mongo":
        cmd = ["docker run --name hello "+private_registry+"mongo:"+tag, "waiting for connections"]
    elif repo == "mysql":
        cmd = ["docker run --name hello -e MYSQL_ROOT_PASSWORD=my-secret-pw "+private_registry+"mysql:"+tag, "ready for connections"]
    elif repo == "neo4j":
        cmd = ["docker run --name hello "+private_registry+"neo4j:"+tag, "Started"]
    elif repo == "postgres":
        cmd = ["docker run --name hello -e POSTGRES_PASSWORD=mysecretpassword "+private_registry+"postgres:"+tag, "database system is ready to accept connections"]
    elif repo == "redis":
        cmd = ["docker run --name hello "+private_registry+"redis:"+tag, "Ready to accept connections"]
    elif repo == "rethinkdb":
        cmd = ["docker run --name hello "+private_registry+"rethinkdb:"+tag, "Server ready"]
    elif repo == "consul":
        cmd = []
    elif repo == "jenkins":
        cmd = []
    elif repo == "maven":
        cmd = []
    elif repo == "kibana":
        cmd = []
    elif repo == "sonarqube":
        cmd = []
    elif repo == "telegraf":
        cmd = []
    elif repo == "sentry":
        cmd = []
    elif repo == "docker":
        cmd = ["docker run --name hello "+private_registry+"docker:"+tag+" echo hello", "hello"]
    elif repo == "drupal":
        cmd = ["docker run --name hello -p 8080:80 "+private_registry+"drupal:"+tag, "apache2 -D FOREGROUND"]
    elif repo == "gradle":
        cmd = ["docker run --name hello "+private_registry+"gradle:"+tag, "BUILD SUCCESSFUL"]
    elif repo == "kong":
        cmd = []
    elif repo == "nextcloud":
        cmd = ["docker run --name hello -p 8080:80 "+private_registry+"nextcloud:"+tag, "apache2 -D FOREGROUND"]
    elif repo == "owncloud":
        cmd = ["docker run --name hello -p 80:80 "+private_registry+"owncloud:"+tag, "apache2 -D FOREGROUND"]
    elif repo == "eclipse-mosquitto":
        cmd = ["docker run --name hello -p 1883:1883 -p 9001:9001 "+private_registry+"eclipse-mosquitto:"+tag, "listen socket on"]
    elif repo == "ghost":
        cmd = ["docker run --name hello -p 3001:2368 "+private_registry+"ghost:"+tag, "Ghost boot"]
    elif repo == "nats":
        cmd = ["docker run --name hello "+private_registry+"nats:"+tag, "Server is ready"]
    elif repo == "registry":
        cmd = ["docker run --name hello "+private_registry+"registry:"+tag, "listening on"]
    elif repo == "solr":
        cmd = ["docker run --name hello "+private_registry+"solr:"+tag, "Server Started"]
    elif repo == "vault":
        cmd = ["docker run --name hello "+private_registry+"vault:"+tag, "Vault server started"]
    elif repo == "wordpress":
        cmd = ["docker run --name hello "+private_registry+"wordpress:"+tag, "apache2 -D FOREGROUND"]
    elif repo == "logstash":
        cmd = ["docker run --name hello "+private_registry+"logstash:"+tag, "Successfully started Logstash"]
    elif repo == "rabbitmq":
        cmd = ["docker run --name hello "+private_registry+"rabbitmq:"+tag, "Server startup complete"]
    elif repo == "golang":
        cmd = ["docker run --name hello "+private_registry+"golang:"+tag+" echo hello", "hello"]
    elif repo == "groovy":
        cmd = ["docker run --name hello "+private_registry+"groovy:"+tag+" echo hello", "hello"]
    elif repo == "java":
        cmd = ["docker run --name hello "+private_registry+"java:"+tag+" echo hello", "hello"]
    elif repo == "jruby":
        cmd = ["docker run --name hello "+private_registry+"jruby:"+tag+" echo hello", "hello"]
    elif repo == "openjdk":
        cmd = ["docker run --name hello "+private_registry+"openjdk:"+tag+" echo hello" "hello"]
    elif repo == "perl":
        cmd = ["docker run --name hello "+private_registry+"perl:"+tag+" echo hello", "hello"]
    elif repo == "php":
        cmd = ["docker run --name hello "+private_registry+"php:"+tag+" echo hello", "hello"]
    elif repo == "python":
        cmd = ["docker run --name hello "+private_registry+"python:"+tag+" echo hello", "hello"]
    elif repo == "ruby":
        cmd = ["docker run --name hello "+private_registry+"ruby:"+tag+" echo hello", "hello"]
    elif repo == "hello-world":
        cmd = ["docker run --name hello "+private_registry+"hello-world:"+tag, "Hello from Docker!"]
    elif repo == "rocket.chat":
        cmd = []
    elif repo == "haproxy":
        cmd = ["docker run --name hello "+private_registry+"haproxy:"+tag+" echo hello", "hello"]
    elif repo == "httpd":
        cmd = ["docker run --name hello "+private_registry+"httpd:"+tag, "httpd -D FOREGROUND"]
    elif repo == "nginx":
        cmd = ["docker run --name hello -p 8080:80 "+private_registry+"nginx:"+tag, ""]
    elif repo == "node":
        cmd = ["docker run --name hello "+private_registry+"node:"+tag+" echo hello", "hello"]
    elif repo == "percona":
        cmd = ["docker run --name hello -e MYSQL_ROOT_PASSWORD=my-secret-pw "+private_registry+"percona:"+tag, "ready for connections"]
    elif repo == "swarm":
        cmd = ["docker run --name hello "+private_registry+"swarm:"+tag+"create", "Token based discovery"]
    elif repo == "tomcat":
        cmd = ["docker run --name hello "+private_registry+"tomcat:"+tag, "Server startup"]
    elif repo == "traefik":
        cmd = ["docker run --name hello -p 8080:8080 -p 80:80 -v $PWD/traefik.toml:/etc/traefik/traefik.toml -v /var/run/docker.sock:/var/run/docker.sock "+private_registry+"traefik:"+tag, ""]

    return cmd