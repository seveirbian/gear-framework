import sys
# package need to be installed, pip install docker
import docker 
import time
import yaml
import os
import xlwt

auto = False

private_registry = "202.114.10.146:10000/"

# result
result = [["tag", "finishTime", "size", "data"], ]

class Puller:

    def __init__(self, images):  
        self.images_to_pull = images

    def check(self):
        # detect whether the file exists, if true, delete it
        if os.path.exists("./images_pulled.txt"):
            os.remove("./images_pulled.txt")
    
    def pull(self):
        self.check()

        client = docker.from_env()
        # if don't give a tag, then all image under this registry will be pulled
        repos = self.images_to_pull[0]["repo"]

        for repo in repos:
            tags = self.images_to_pull[1][repo]
            for tag in tags:
                print "start pulling: ", private_registry+repo, ":", tag

                # get present time
                startTime = time.time()

                # get present net data
                cnetdata = get_net_data()

                # pull images
                try:
                    image_pulled = client.images.pull(repository=private_registry+repo, tag=str(tag))

                    # print pull time
                    finishTime = time.time() - startTime

                    print "finished in " , finishTime, "s"

                    # get image's size
                    size = image_pulled.attrs[u'Size'] / 1000000.0
                    print "image size: ", size

                    data = get_net_data() - cnetdata

                    print "pull data: ", data

                    print "\n"

                    # record the image and its pulling time
                    result.append([tag, finishTime, size, data])

                except docker.errors.NotFound:
                    print private_registry+repo + " not found...\n\n"
                except docker.errors.ImageNotFound:
                    print private_registry+repo + " image not fount...\n\n"

                if auto != True:  
                    raw_input("Next?")

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

    data = 0
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

    puller = Puller(images)

    puller.pull()

    # create a workbook sheet
    workbook = xlwt.Workbook()
    sheet = workbook.add_sheet("run_time")

    for row in range(len(result)):
        for column in range(len(result[row])):
            sheet.write(row, column, result[row][column])

    workbook.save(os.path.split(os.path.realpath(__file__))[0]+"/pull.xls")