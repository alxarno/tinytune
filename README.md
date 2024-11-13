<div align="center">

[![TinyTune][repo_logo_img]][repo_url]

# TinyTune

[![Go version][go_version_img]][go_dev_url]
[![License][repo_license_img]][repo_license_url]

**TinyTune** is tiny **media server** with **web** interface.

</div>

## ⚡️ Install

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
   --video, --av                            allows the server to process videos (default: true)

   Server:

   --port value, -p value  http server port (default: 8080)


COPYRIGHT:
   (c) github.com/alxarno/tinytune

```
## 🖥️ Development

## 🧾 License

<!-- Go -->

[go_version_img]: https://img.shields.io/badge/Go-1.22+-00ADD8?style=for-the-badge&logo=go
[go_report_img]: https://img.shields.io/badge/Go_report-A+-success?style=for-the-badge&logo=none
[go_dev_url]: https://pkg.go.dev/github.com/create-go-app/cli/v4

<!-- Repository -->

[repo_url]: https://github.com/alxarno/tinytune
[repo_logo_img]: ./web/assets/icon.jpg
[repo_license_url]: https://github.com/alxarno/tinytune/blob/main/LICENSE
[repo_license_img]: https://img.shields.io/github/license/alxarno/tinytune?style=for-the-badge&logo=none
