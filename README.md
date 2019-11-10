# redundantFinder
Find and delete redundant files in specific directory. Support multiple directories.

You could download executable files directly in [release page](https://github.com/ceeji/redundantFinder/releases).

> Note: By default, only several file extensions are selected. You can change it by using `-ext` option. But remember only select what you need. For some extensions you should NOT delete same files even if they are the same, because these files might be used by OS in special way.

## Highlights
- `Fast`: Very fast when being used with large files and only a small part of files are redundant. It will firstly compare files using file size, and then compare file's content using `sha512`. Normally every file has different size (especially for videos and images) so we could reduce times we need to scan file's content.
- `Reliable`: Use `sha512` (not `md5` or `sha1`) to reduce the chance that two files are different but have same hash. But remember there are still very very low possibility. Please take care. Also, unless used with `-r`, no action will be actually taken.
- `Multiple Directories`: You could use it with multiple directories.
- `Ordered`: The program will sort file paths in a redundant group so you could infer which file will be kept to avoid you looking for files in different directories.
- `Cross Platform`: All major platforms are supported: FreeBSD / Windows / Linux / Arm and so on.

## Usage

```shell
redundantFinder [-r] [-ext=extensions] <target_directory> ...
Copyright(C) 2019 Ceeji Cheng <hi.ceeji#gmail.com> and contributors

  -ext string
        specify file extensions for scanning, any file without these extension will be ignored. Multiple values should be splited by '|'. If empty, any file will be included. (default "jpg|png|arw|raw|nec|jpeg|mp4|mp3|json|m4a|avi|mpeg|mpg|dat|doc|docx|ppt|pptx|db|txt|zip|gz|bz|7z|tar|rar|bzip|iso|pkg|wav")
  -r    delete redundant copies after scan
  -v    show version and exit
```
