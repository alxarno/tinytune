<div align="center">



# [![TinyTune][repo_logo_img]][repo_url] TinyTune

[![Go version][go_version_img]][go_dev_url]
[![License][repo_license_img]][repo_license_url]

<img alt="demo" src="./docs/demo.gif">

**TinyTune** is tiny **media server** with **web** interface.

It allows you to watch **videos**, **images** and also has a **search** feature.

</div>

## 🎯 Features

 - [x] **one** executable file
 - [x] searching
 - [x] **animated** previews for video
 - [x] streaming **.flv**, **.avi**, etc
 - [x] rich settings for **optimize** indexing big media folders


## ⚡️ Install

The latest version of the TinyTune can be found on the GitHub [releases page](https://github.com/alxarno/tinytune/releases).

> This program uses **FFmpeg**, make sure it is available for calling. Instructions for installing it can be found [here](https://www.ffmpeg.org/download.html).

### Linux
```
# check that you have ffmpeg installed
ffmpeg -v

wget https://github.com/alxarno/tinytune/releases/download/v1.2.2/tinytune_linux_amd64

mv tinytune_linux_amd64 /usr/local/bin/tinytune

chmod +x /usr/local/bin/tinytune

tinytune /YOUR_MEDIA_FOLDER
```

## 🚀 Performance

The first start of the program takes some time, because the initial processing of files takes place to create previews and extract meta information. The following are the statistics of the test folder with data, hardware, and the result.

```
Hardware:
   CPU: AMD Ryzen 7 2700X (16) @ 3.700GHz
   RAM: 32036MiB
   Disk: Samsung SSD 970 EVO Plus 250GB

Folder:
   4 677 items, totalling 20,6 GB

Folder files statistic:
     size       type  count   
----------------------------
    16GiB        mp4    199
   1,9GiB        avi      4
   1,5GiB        jpg   4249
   145MiB        flv      2
    18MiB       jpeg     54
   4,3MiB        gif      6
   2,7MiB        png      3
   1,9MiB        JPG     20
----------------------------

Results:
   index.tinytune file - 33,7 MB
   indexing time - 00:01:09
```


## ⚙️ Commands & Options

```
NAME:
   TinyTune - the tiny media server

USAGE:
   tinytune [data folder path] [global options]

VERSION:
   1.2.0

AUTHOR:
   alxarno <alexarnowork@gmail.com>

COMMANDS:
   help, h  Shows a list of commands or help for one command

GLOBAL OPTIONS:
   --help, -h           show help
   --print-version, -v  print only the version (default: false)

   Common:

   --index-save, --is  the program creates a special file in the working directory “index.tinytune”. This file stores all necessary data obtained during indexing of the working directory.
                You can turn off its saving, but at the next startup, the application will start processing again (default: true)

   Processing:
    In order for the web interface to be able to view thumbnails of media files, as well as play them, the program needs to process them and get meta information.
    This process can be long, so here are the options that will help limit the number of files to process.

   --excludes value  if you want to more finely restrict the files to be processed, use this option. You can specify multiple regular expressions, separated by commas.
                Files that fall under one of these expressions will not be processed (but you will still see them in the interface).
                Example: '\\.(mp4|avi)$' -> turn off processing for all files with .mp4 and .avi extensions
   --image           allows the server to process images, to show thumbnails (default: true)
   --includes value  this parameter will help to include back into processing files that were disabled by the '--exclude' parameter. Regular expressions are also used here, separated by commas.
                Example: 'video/sample[.]mp4$' -> will return the sample.mp4 file, which is located in the video folder (no matter at what level the folder is located) to processing
   --max-file-size value  this option restricts files from being processed if their size exceeds a certain value. Values can be specified as follows: 25KB, 10mb, 1GB, 2gb (default: "-1B")
   --max-images value     limits the number of image files to be processed (thumbnails producing) (default: -1)
   --max-videos value     limits the number of video files to be processed (thumbnails producing) (default: -1)
   --parallel value       simultaneous image/video processing (!large values increase RAM consumption!) (default: 16)
   --timeout value        sometimes some files take too long to process, here you can specify a time limit in which they should be processed. Examples of values: 5m, 120s (default: "2m")
   --video                allows the server to process videos, for playing them in browser and show thumbnails (default: true)

   Server:

   --port value, -p value  http server port (default: 8080)
   --streaming value       some files cannot be played in the browser, such as flv and avi. Therefore, such files need to be transcoded.
                Specify here, using regular expressions, which files you would like to transcode on the fly for browser viewing (default: "\\.(flv|f4v|avi)$")


COPYRIGHT:
   (c) github.com/alxarno/tinytune

```
## 🖥️ Development

```
#!/usr/bin/env bash

git clone https://github.com/alxarno/tinytune
cd tinytune

# Install dependencies
make ubuntu
# Start hot-reload way
make watch
```

## 🧾 License

Usage is provided under the [GPLv3 License](./LICENSE). See LICENSE for the full details.

<!-- Go -->

[go_version_img]: https://img.shields.io/badge/Go-1.22+-00ADD8?style=for-the-badge&logo=go
[go_report_img]: https://img.shields.io/badge/Go_report-A+-success?style=for-the-badge&logo=none
[go_dev_url]: https://pkg.go.dev/github.com/create-go-app/cli/v4

<!-- Repository -->

[repo_url]: https://github.com/alxarno/tinytune
[repo_logo_img]: ./docs/icon.jpg
[repo_license_url]: https://github.com/alxarno/tinytune/blob/main/LICENSE
[repo_license_img]: https://img.shields.io/github/license/alxarno/tinytune?style=for-the-badge&logo=none