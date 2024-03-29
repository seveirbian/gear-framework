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
import xlwt
# package need to be installed, apt-get install python-pymongo
import pymongo

auto = False

private_registry = "202.114.10.146:9999/"
suffix = "-gearmd"

apppath = ""

# run paraments
hostPort = 27017
localVolume = "/var/lib/gear/volume"
pwd = os.getcwd()

runEnvironment = ["MONGO_INITDB_ROOT_USERNAME=bian", 
                  "MONGO_INITDB_ROOT_PASSWORD=1122", 
                  "MONGO_INITDB_DATABASE=games", ]
runPorts = {"27017/tcp": hostPort, }
runVolumes = {localVolume: {'bind': '/data/db', 'mode': 'rw'},}
runWorking_dir = ""
runCommand = ""
waitline = "waiting for connections"

# result
result = [["tag", "finishTime", "local data", "pull data"], ]

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
                        time.sleep(0.1)
                        pass

                while True:
                    if time.time() - startTime > 600:
                        break

                    try:
                        cli = pymongo.MongoClient("mongodb://%s:%s@127.0.0.1" % ("bian", "1122"))
                        db = cli["games"]
                        print "successfully open db!"
                        my_directory = {"ID": 1, "NAME": "Three kingdoms"}
                        posts = db.posts
                        posts.insert_one(my_directory)
                        print "successfully insert!"
                        posts.update_one({"NAME": "Three kingdoms"}, {"$set": {"NAME": "data2"}})
                        print "successfully update!"
                        print posts.find_one({"ID": 1})
                        posts.delete_one({"ID": 1})
                        print "successfully delete!"
                        break
                    except:
                        time.sleep(0.01) # wait 10ms
                        pass

                # print run time
                finishTime = time.time() - startTime

                print "finished in " , finishTime, "s"

                container_path = os.path.join("/var/lib/gear/private", private_repo)
                local_data = subprocess.check_output(['du','-sh', container_path]).split()[0].decode('utf-8')

                print "local data: ", local_data

                pull_data = get_net_data() - cnetdata

                print "pull data: ", pull_data

                print "\n"

                try: 
                    container.kill()
                except:
                    print "kill fail!"
                    pass
                    
                container.remove(force=True)

                # record the image and its Running time
                result.append([tag, finishTime, local_data, pull_data])

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

    workbook.save(os.path.split(os.path.realpath(__file__))[0]+"/second_run.xls")