import sys
# package need to be installed, pip install docker
import docker
import time
import yaml
import os
import hashlib

class Tester:

    def __init__(self, path):
        self.path = path
        self.files = []
        self.unique_files_fingerprint = []

    def test(self):
        total_size = 0

        for path, _, files in os.walk(self.path):
            for file in files:
                image_files.append(os.path.join(path, file))

        # get unique files
        for file in self.files:
            if not os.path.isfile(file):
                pass
            else:
                if os.path.islink(file):
                    print "hashing symlink", file, "...\n"
                    
                    hash = hashlib.sha256()
                    hash.update(os.readlink(file))
                    hash_value = hash.hexdigest()
                    if hash_value not in self.unique_files_fingerprint:
                        self.unique_files_fingerprint.append(hash_value)
                        total_size += os.lstat(file).st_size
                else:
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
                            total_size += os.path.getsize(file)

        with open("./images_tested.txt", "a") as f:
            f.write("unique_files_num: "+str(len(self.unique_files_num))+"\n")
            f.write("total size: "+str(self.size_in_file_level/1024.0/1024.0))

if __name__ == "__main__":
    if len(sys.argv) >= 2 or len(sys.argv) <= 1:
        exit()

    tester = Tester(sys.argv[1])

    tester.test()