import sys
# package need to be installed, pip install docker
import docker
import time
import yaml
import os
import hashlib

class Tester:

    def __init__(self, images):
        self.images_to_test = images

        # all files
        self.files = []
        self.files_num = 0
        # unique files in common files
        self.unique_files = []
        self.unique_files_num = 0

        self.unique_files_fingerprint = []
        self.unique_files_fingerprint_num = 0

        # all common files
        self.common_files_num = 0

        # all symlink files
        # self.symlink_files = []
        self.symlink_files_num = 0

        # size in file-level sharing
        self.size_in_file_level = 0

    def check(self):
        # detect whether the file exists, if true, delete it
        if os.path.exists("./images_tested.txt"):
            os.remove("./images_tested.txt")

    def test(self):
        self.check()

        total_size = 0

        client = docker.from_env()

        # if don't give a tag, then all image under this registry will be pulled
        repos = self.images_to_test[0]["repo"]

        for repo in repos:
            tags = self.images_to_test[1][repo]

            for tag in tags:

                # test images
                try:
                    image = client.images.get(name=repo+":"+str(tag))

                    # this is a array of upperdir in unicode 
                    upper_layers_in_unicode = image.attrs[u'GraphDriver'][u'Data'][u'UpperDir']

                    # only if layers_in_unicode only has ascii characters, use like this
                    upper_layers = str(upper_layers_in_unicode)
 
                    if image.attrs[u'GraphDriver'][u'Data'].has_key(u'LowerDir'):
                        lower_layers_in_unicode = image.attrs[u'GraphDriver'][u'Data'][u'LowerDir']
                        lower_layers = str(lower_layers_in_unicode)
                        # this is the array of path
                        layers = upper_layers.split(":") + lower_layers.split(":")
                    else:
                        layers = upper_layers.split(":")

                    image_files = []

                    for layer in layers:

                        for path, _, files in os.walk(layer):
                            for file in files:
                                image_files.append(os.path.join(path, file))

                    # print image_files
                    self.files += image_files

                except docker.errors.NotFound:
                    print repo + " not found...\n\n"

        # get unique files
        for file in self.files:
            if not os.path.isfile(file):
                pass
            else:
                if os.path.islink(file):
                    self.symlink_files_num += 1
                    # self.symlink_files.append(file)
                    # os.readlink(path) can return the content of symlink
                    print "hashing symlink", file, "...\n"
                    
                    hash = hashlib.sha256()
                    hash.update(os.readlink(file))
                    hash_value = hash.hexdigest()
                    if hash_value not in self.unique_files_fingerprint:
                        self.unique_files_fingerprint.append(hash_value)
                        self.unique_files.append(file)
                        total_size += os.lstat(file).st_size
                else:
                    self.common_files_num += 1
                    print "hashing", file, "...\n"
                    
                    with open(file, "rb") as f:
                        hash = hashlib.sha256()
                        while True:
                            data = f.read(200*1024*1024)
                            if not data:
                                break
                            hash.update(data)
                        hash_value = hash.hexdigest()
                        if hash_value not in self.unique_files_fingerprint:
                            self.unique_files_fingerprint.append(hash_value)
                            self.unique_files.append(file)
                            total_size += os.path.getsize(file)
        
        # get those unique files' size
        # total_size = 0
        # for file in self.unique_files:
        #     total_size += os.path.getsize(file)

        self.files_num = len(self.files)
        self.unique_files_num = len(self.unique_files)
        self.unique_files_fingerprint_num = len(self.unique_files_fingerprint)
        self.size_in_file_level = total_size

        with open("./images_tested.txt", "a") as f:
            # f.write("files:\n")
            # for file in self.files:
            #     f.write(file+"\n")
            #     f.write("unique_files:\n")
            for file in self.unique_files:
                f.write(file+"\n")
            f.write("files_num: "+str(self.files_num)+"\n")
            f.write("symlink_files_num: "+str(self.symlink_files_num)+"\n")
            f.write("common_files_num: "+str(self.common_files_num)+"\n")
            f.write("unique_files_num: "+str(self.unique_files_num)+"\n")
            f.write("total size: "+str(self.size_in_file_level/1024.0/1024.0))

class Generator:

    def __init__(self, profilePath=""):
        self.profilePath = profilePath

    def generateFromProfile(self):
        if self.profilePath == "":
            print "Error: profile path is null"

        with open(self.profilePath, 'r') as f:
            self.images_to_push = yaml.load(f)

        return self.images_to_push


if __name__ == "__main__":
    if len(sys.argv) >= 3 or len(sys.argv) <= 1:
        exit()

    generator = Generator(sys.argv[1])

    images_to_test = generator.generateFromProfile()

    tester = Tester(images_to_test)

    tester.test()

    print tester.size_in_file_level/1024.0/1024.0