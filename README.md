# tiny

CLI to shrink JPEG / PNG images.

## Installation

```shell
go install github.com/alex-vit/tiny@latest
```

## Usage

```
# tiny cat.jpg dog.png ...

# tiny .
```

`tiny .` expects a literal dot, not any folder, which I realized just now and migth fix in the future.

> [!WARNING]
> tiny replaces the image with the compressed one. If you'd like tiny to preserve the original, flip the `preserveOriginals` constant and recompile.
