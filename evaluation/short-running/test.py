import sys
# package need to be installed, pip install docker
import docker 
import time
import yaml
import os
import random
import subprocess
import signal
import urllib2
import shutil

auto = False

private_registry = "202.114.10.146:9999/"

apppath = ""

# run paraments
hostPort = 80
localVolume = ""
pwd = os.path.split(os.path.realpath(__file__))[0]

runEnvironment = []
runPorts = {"80/tcp": hostPort,}
runVolumes = {}
runWorking_dir = ""
runCommand = ""
waitline = ""

def run():
    client = docker.from_env()

    private_repo = private_registry + "httpd" + ":" + "2.4.41"

    # create a random name
    runName = '%d' % (random.randint(1,100000000))

    # get present time
    startTime = time.time()

    # run images
    container = client.containers.create(image=private_repo, environment=runEnvironment,
                        ports=runPorts, volumes=runVolumes, working_dir=runWorking_dir,
                        command=runCommand, name=runName, detach=True)

    times = 100

    while times > 0:

        container.start()

        while True:
            if time.time() - startTime > 600:
                break

            try:
                req = urllib2.urlopen('http://localhost:%d'%hostPort)
                if req.read().find("It works!") >= 0:
                    print "OK!"
                req.close()
                break
            except:
                time.sleep(0.1) # wait 100ms
                pass

        try: 
            container.kill()
        except:
            print "kill fail!"
            pass

        times = times - 1

    # print run time
    finishTime = time.time() - startTime

    print "finished in " , finishTime, "s\n"
                    
    container.remove(force=True)


if __name__ == "__main__":

    run()