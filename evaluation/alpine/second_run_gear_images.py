import sys
# package need to be installed, pip install docker
import docker 
import time
import random
import yaml
import os
import subprocess
import signal
import urllib2
import psycopg2
import shutil
import pymongo
import xlwt
# package need to be installed, apt-get install python-mysqldb
import MySQLdb

auto = False

private_registry = "202.114.10.146:9999/"
suffix = "-gearmd"

apppath = ""

# run paraments
hostPort = 5000
localVolume = ""
pwd = os.getcwd()

runEnvironment = []
runPorts = {}
runVolumes = {}
runWorking_dir = ""
runCommand = "echo hello"
waitline = "hello"

# result
result = [["tag", "finishTime"], ]

class Runner:

    def __init__(self, images):  
        self.images_to_pull = images

    def check(self):
        # detect whether the file exists, if true, delete it
        if os.path.exists("./images_run.txt"):
            os.remove("./images_run.txt")
    
    def run(self):
        self.check()

        client = docker.from_env()
        # if don't give a tag, then all image under this registry will be pulled
        repos = self.images_to_pull[0]["repo"]

        for repo in repos:
            tags = self.images_to_pull[1][repo]
            for tag in tags:
                private_repo = private_registry + repo + suffix + ":" + tag

                if localVolume != "":
                    if os.path.exists(localVolume) == False:
                        os.makedirs(localVolume)

                print "start running: ", private_repo

                # create a random name
                runName = '%d' % (random.randint(1,100000000))

                # get present time
                startTime = time.time()

                # get present net data
                cnetdata = get_net_data()

                # run images
                container = client.containers.create(image=private_repo, environment=runEnvironment,
                                    ports=runPorts, volumes=runVolumes, working_dir=runWorking_dir,
                                    command=runCommand, name=runName, detach=True)

                container.start()

                while True:
                    if waitline == "":
                        break
                    elif container.logs().find(waitline) >= 0:
                        break
                    else:
                        time.sleep(0.01)
                        pass

                # print run time
                finishTime = time.time() - startTime

                print "finished in " , finishTime, "s"

                data = get_net_data() - cnetdata

                print "pull data: ", data

                print "\n"

                try: 
                    container.kill()
                except:
                    print "kill fail!"
                    pass
                    
                container.remove(force=True)

                # record the image and its Running time
                result.append([tag, finishTime, data])

                if auto != True: 
                    raw_input("Next?")
                else:
                    time.sleep(5)

                if localVolume != "":
                    shutil.rmtree(localVolume)

class Generator:
    
    def __init__(self, profilePath=""):
        self.profilePath = profilePath
    
    def generateFromProfile(self):
        if self.profilePath == "":
            print "Error: profile path is null"
        
        with open(self.profilePath, 'r') as f:
            self.images = yaml.load(f)

        return self.images

def get_net_data():
    netCard = "/proc/net/dev"
    fd = open(netCard, "r")

    for line in fd.readlines():
        if line.find("enp0s3") >= 0:
            field = line.split()
            data = float(field[1]) / 1024.0 / 1024.0

    fd.close()
    return data

if __name__ == "__main__":

    if len(sys.argv) == 2:
        auto = True

    generator = Generator(os.path.split(os.path.realpath(__file__))[0]+"/image_versions.yaml")

    images = generator.generateFromProfile()

    runner = Runner(images)

    runner.run()

    # create a workbook sheet
    workbook = xlwt.Workbook()
    sheet = workbook.add_sheet("run_time")

    for row in range(len(result)):
        for column in range(len(result[row])):
            sheet.write(row, column, result[row][column])

    workbook.save("./second_run.xls")