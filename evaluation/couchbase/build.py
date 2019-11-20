import sys
# package need to be installed, pip install docker
import docker 
import time
import yaml
import os
import xlwt
import subprocess

# result
result = [["tag", "build"], ]

cmd = "/home/vagrant/go/src/github.com/seveirbian/gear/gear"

class Puller:

    def __init__(self, images):  
        self.images_to_pull = images
    
    def pull(self):

        repos = self.images_to_pull[0]["repo"]

        for repo in repos:
            tags = self.images_to_pull[1][repo]
            for tag in tags:

                image_to_build = "202.114.10.146:10000/" + repo + ":" + tag
                print "start pulling: ", image_to_build

                # get present time
                startTime = time.time()

                subprocess.call(cmd + " build " + image_to_build, shell=True)

                # record the image and its pulling time
                result.append([tag, buildTime])

class Generator:
    
    def __init__(self, profilePath=""):
        self.profilePath = profilePath
    
    def generateFromProfile(self):
        if self.profilePath == "":
            print "Error: profile path is null"
        
        with open(self.profilePath, 'r') as f:
            self.images = yaml.load(f)

        return self.images

if __name__ == "__main__":

    if len(sys.argv) == 2:
        auto = True

    generator = Generator(os.path.split(os.path.realpath(__file__))[0]+"/image_versions.yaml")

    images = generator.generateFromProfile()

    puller = Puller(images)

    puller.pull()

    # create a workbook sheet
    workbook = xlwt.Workbook()
    sheet = workbook.add_sheet("build_time")

    for row in range(len(result)):
        for column in range(len(result[row])):
            sheet.write(row, column, result[row][column])

    workbook.save(os.path.split(os.path.realpath(__file__))[0]+"/buildtime.xls")