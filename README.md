<div align="center">



# [![TinyTune][repo_logo_img]][repo_url] TinyTune

[![Go version][go_version_img]][go_dev_url]
[![License][repo_license_img]][repo_license_url]

<img alt="demo" src="./docs/demo.gif">

**TinyTune** is tiny **media server** with **web** interface.

It allows you to watch **videos**, **images** and also has a **search** feature.

</div>

## ⚡️ Install

The latest version of the TinyTune can be found on the GitHub [releases page](https://github.com/alxarno/tinytune/releases).

### Linux
```
# check that you have ffmpeg installed
ffmpeg -v

wget https://github.com/alxarno/tinytune/releases/download/v1.0.0/tinytune_linux_amd64

mv tinytune_linux_amd64 /usr/local/bin/tinytune

chmod +x /usr/local/bin/tinytune

tinytune /YOUR_MEDIA_FOLDER
```

## ⚙️ Commands & Options

```
NAME:
   TinyTune - the tiny media server

USAGE:
   tinytune [data folder path] [global options]

VERSION:
   n/a

AUTHOR:
   alxarno <alexarnowork@gmail.com>

COMMANDS:
   help, h  Shows a list of commands or help for one command

GLOBAL OPTIONS:
   --help, -h           show help
   --print-version, -v  print only the version (default: false)

   FFmpeg:

   --acceleration, -a  allows to utilize GPU computing power for ffmpeg (default: true)

   Processing:

   --image, --ai                            allows the server to process images (default: true)
   --max-new-image-items value, --ni value  limits the number of new image files to be processed (default: -1)
   --max-new-video-items value, --nv value  limits the number of new video files to be processed (default: -1)
   --parallel value, --pl value             simultaneous file processing (!large values increase RAM consumption!) (default: 16)
   --video, --av                            allows the server to process videos (default: true)

   Server:

   --port value, -p value  http server port (default: 8080)


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