### Test

#### Build
``CGO_ENABLED=0 go build -o news ./cmd/main.go``

#### Run
##### Local
```
mkdir -p `pwd`/coll_dir
./news coll -t mobile -s daum -d ./coll_dir -e -l 3 -b
```
##### Docker
```
mkdir -p `pwd`/coll_dir
docker run -it --name coll-news -v `pwd`/coll_dir:/home/coll/coll_dir coll-news:v0.1.11 news coll -t mobile -s daum -d ./coll_dir -e -l 3 -b /usr/bin/chromium-browser
```
##### Synology
[synology/coll-news.json](synology/coll-news.json)
should modify volume_bindings configuration
```
   "volume_bindings" : [
      {
         "host_volume_file" : "/apps/coll_news/coll_dir",
         "mount_point" : "/home/coll/coll_dir",
         "type" : "rw"
      }
```