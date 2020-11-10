# gd-suite

## touch
gd-suite supports touching **folders** in Google Drive. For example:

```bash
gds touch gd:/folder1
gds touch gd:/folder2-with-padding-slash/
```
It also supports touching **parents**. Just use:

```bash
gds touch gd:/parent/child/file
```

It will touch all file/folders in sequence below:

1. gd:/parent/child/file
2. gd:/parent/child/
3. gd:/parent/

If you don't like this feature, you can use `-F`(`--first-file`) to disable it.