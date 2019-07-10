import sys
# package need to be installed, pip install docker
import docker 
import time
import yaml
import os
import xlwt
import subprocess
import generate_cmd

class Puller:

    def __init__(self, images):  
        self.images_to_pull = images
        self.private_registry = "202.114.10.146:9999"
        self.result = []


    def check(self):
        # detect whether the file exists, if true, delete it
        if os.path.exists("./images_run.txt"):
            os.remove("./images_run.txt")
    
    def pull(self):
        self.check()

        client = docker.from_env()
        # if don't give a tag, then all image under this registry will be pulled
        repos = self.images_to_pull[0]["repo"]

        for repo in repos:
            tags = self.images_to_pull[1][repo]
            for tag in tags:
                print "start pulling: ", repo, ":", tag

                # get present time
                startTime = time.time()

                # pull images
                try:
                    image_pulled = client.images.pull(repository=self.private_registry+"/"+repo, tag=str(tag))

                    # print pull time
                    pullTime = time.time() - startTime

                    print "finished in " , pullTime, "s\n"

                    cmds = generate_cmd.cmdGenerate(repo, str(tag))

                    # start up the image
                    print "start startup containers..."
                    startTime = time.time()

                    self.startup(cmds)

                    startupTime = time.time() - startTime
                    print "startup finished in ", startupTime, "s\n"

                    # kill and remove all containers
                    p = subprocess.Popen("docker rm -f hello", shell=True,
                                bufsize=1, stderr=subprocess.STDOUT,
                                stdout=subprocess.PIPE)
                    while True:
                        line = p.stdout.readline()

                        if line.find("hello") >= 0:
                            break

                    # record the image and its pulling time
                    self.record(repo, tag, image_pulled.attrs[u'Size'], pullTime, startupTime)

                    client.images.remove(self.private_registry+"/"+repo+":"+str(tag))

                except docker.errors.NotFound:
                    print repo + " not found...\n\n"
                except docker.errors.ImageNotFound:
                    print repo + " image not fount...\n\n"

    def startup(self, cmds):
        if cmds == []:
            return
        
        if cmds[1] == "":
            if cmds[0].find("nginx") >= 0:
                subprocess.Popen(cmds[0], shell=True,
                                    bufsize=1, stderr=subprocess.STDOUT,
                                    stdout=subprocess.PIPE)
                while True:
                    p = subprocess.Popen("curl localhost:8080", shell=True,
                                    bufsize=1, stderr=subprocess.STDOUT,
                                    stdout=subprocess.PIPE)
                    line = p.stdout.readline()
                    if line.find("Thank you for using nginx") >= 0:
                        break
            elif cmds[0].find("traefik") >= 0:
                subprocess.Popen(cmds[0], shell=True,
                                    bufsize=1, stderr=subprocess.STDOUT,
                                    stdout=subprocess.PIPE)
                while True:
                    p = subprocess.Popen("curl localhost:8080", shell=True,
                                    bufsize=1, stderr=subprocess.STDOUT,
                                    stdout=subprocess.PIPE)
                    line = p.stdout.readline()
                    if line.find("dashboard") >= 0:
                        break
        else:
            p = subprocess.Popen(cmds[0], shell=True,
                                bufsize=1, stderr=subprocess.STDOUT,
                                stdout=subprocess.PIPE)
            while True:
                line = p.stdout.readline()
                if line == "":
                    continue
        
                # print "out: " + line.strip()

                if line.find(cmds[1]) >= 0:
                    break

    def record(self, repo, tag, size, pullTime, startupTime):
        self.result.append([repo, tag, size, pullTime, startupTime])

class Generator:
    
    def __init__(self, profilePath=""):
        self.profilePath = profilePath
    
    def generateFromProfile(self):
        if self.profilePath == "":
            print "Error: profile path is null"
        
        with open(self.profilePath, 'r') as f:
            self.images_to_pull = yaml.load(f)

        return self.images_to_pull


if __name__ == "__main__":
    if len(sys.argv) >= 3 or len(sys.argv) <= 1:
        exit()

    generator = Generator(sys.argv[1])

    images_to_pull = generator.generateFromProfile()

    puller = Puller(images_to_pull)

    puller.pull()

    # create a workbook sheet
    workbook = xlwt.Workbook()
    sheet = workbook.add_sheet("run_time")

    for row in range(len(puller.result)):
        for column in range(len(puller.result[row])):
            sheet.write(row, column, puller.result[row][column])

    workbook.save("./image_run.xls")