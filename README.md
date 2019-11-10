# redundantFinder
Find and delete redundant files in specific directory. Support multiple directories.

> Note: By default, only several file extensions are selected. You can change it by using `-ext` option. But remember only select what you need. For some extensions you should NOT delete same files even if they are the same, because these files might be used by OS in special way.

## Highlights
- `Fast`: Very fast when being used with large files and only a small part of files are redundant. It will firstly compare files using file size, and then compare file's content using `sha512`. Normally every file has different size (especially for videos and images) so we could reduce times we need to scan file's content.
- `Reliable`: Use `sha512` (not `md5` or `sha1`) to reduce the chance that two files are different but have same hash. But remember there are still very very low possibility. Please take care.
- `Multiple Directories`: You could use it with multiple directories.
- `Cross Platform`: All major platforms are supported.

## Usage
redundantFinder [-r] [-ext=extensions] <target_directory> ...
  -ext string
        specify file extension for scanning, any file without these extension will be ignored. it should be split by '|'. (default "jpg|png|arw|raw|nec|jpeg")
  -r    delete extra copies after scan
