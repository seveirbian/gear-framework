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

                    if cmds != {}:
                        container, createTime, appTime, startupTime = self.startup(client, repo, cmds)
                        # kill and remove all containers
                        container.remove(force=True)
                    else:
                        createTime = "none"
                        appTime = "none"
                        startupTime = "this image should not be tested! "

                    print "startup finished in ", startupTime, "s\n"

                    # record the image and its pulling time
                    self.record(repo, tag, image_pulled.attrs[u'Size']/1024.0/1024.0, pullTime, createTime, appTime, startupTime)

                    client.images.remove(self.private_registry+"/"+repo+":"+str(tag), force=True)

                except docker.errors.NotFound:
                    print repo + " not found...\n\n"
                except docker.errors.ImageNotFound:
                    print repo + " image not fount...\n\n"

    def startup(self, client, repo, cmds):
        if cmds == {}:
            return

        startTime = time.time()

        # start a container based on the image
        container = client.containers.create(image=cmds["image"], environment=cmds["environment"],
                                        ports=cmds["ports"], volumes=cmds["volumes"], working_dir=cmds["working_dir"],
                                        command=cmds["command"], detach=True)
        
        createTime = time.time() - startTime

        startTime = time.time()
        container.start()

        clock = time.time()
        isOverTime = False

        if cmds["waitline"] == "":
            if repo == "nginx":
                isFound = False
                while True:
                    if time.time() - clock > 60:
                        isOverTime = True
                        break

                    p = subprocess.Popen("curl localhost:8080", shell=True,
                                        bufsize=1, stderr=subprocess.STDOUT,
                                        stdout=subprocess.PIPE)
                    while True:
                        line = p.stdout.readline()
                        # print line
                        if line != '':
                            if line.find("Thank you for using nginx") >= 0:
                                isFound = True
                                break
                        else:
                            break
                    
                    if isFound:
                        break

            elif repo == "traefik":
                isFound = False
                while True:
                    if time.time() - clock > 60:
                        isOverTime = True
                        break

                    p = subprocess.Popen("curl localhost:8080", shell=True,
                                        bufsize=1, stderr=subprocess.STDOUT,
                                        stdout=subprocess.PIPE)
                    while True:
                        line = p.stdout.readline()
                        # print line
                        if line != '':
                            if line.find("dashboard") >= 0:
                                isFound = True
                                break
                        else:
                            break
                    
                    if isFound:
                        break

            elif repo == "node":
                isFound = False
                while True:
                    if time.time() - clock > 60:
                        isOverTime = True
                        break

                    p = subprocess.Popen("curl localhost:8080", shell=True,
                                        bufsize=1, stderr=subprocess.STDOUT,
                                        stdout=subprocess.PIPE)
                    while True:
                        line = p.stdout.readline()
                        # print line
                        if line != '':
                            if line.find("hello") >= 0:
                                isFound = True
                                break
                        else:
                            break
                    
                    if isFound:
                        break

            elif repo == "memcached":
                isFound = False
                while True:
                    if time.time() - clock > 60:
                        isOverTime = True
                        break

                    p = subprocess.Popen("curl localhost:8080", shell=True,
                                        bufsize=1, stderr=subprocess.STDOUT,
                                        stdout=subprocess.PIPE)
                    while True:
                        line = p.stdout.readline()
                        # print line
                        if line != '':
                            if line.find("Empty reply from server") >= 0:
                                isFound = True
                                break
                        else:
                            break
                    
                    if isFound:
                        break

        else:
            while True:
                if time.time() - clock > 60:
                    isOverTime = True
                    break
                if container.logs().find(cmds["waitline"]) >= 0:
                    break
        
        appTime = time.time() - startTime

        if isOverTime:
            startupTime = "this image needs to test again"
        else:
            startupTime = createTime + appTime

        return container, createTime, appTime, startupTime

    def record(self, repo, tag, size, pullTime, createTime, appTime, startupTime):
        self.result.append([repo, str(tag), size, pullTime, createTime, appTime, startupTime])

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