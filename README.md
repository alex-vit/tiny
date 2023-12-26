# tiny

A CLI to shrink images.

## Usage

```shell
tiny cat.png dog.jpg
```

Will shrink files in-place, saving originals in `cat_original.png` and similar.

## Building

### Just run something

```shell
./gradlew installDist
./build/install/tiny/bin/tiny # run!
```

### "Distributions"

```shell
./gradlew build # creates a bunch of stuff in ./build/distributions
```

Then publish these to `brew` or something.

### Standalone (graalvm)

```shell
./gradlew shadowJar # creates ./build/libs/tiny-$version-all.jar
native-image -jar build/libs/tiny-$version-all.jar tiny
upx tiny # to make it significantly lighter, like 34MB down to 13MB lighter.
```

You'll need
- `native-image` that is part of [GraalVM](https://www.graalvm.org/latest/docs/getting-started/) and lives in `$GraalVMPath/bin/native-image`
- [upx](https://upx.github.io/)

## Todo

- [x] folder support: `tiny .`
- [ ] _"Are you sure?_ for folders
- [ ] set up github actions building shadow jar and binaries for all three platforms that I can cargo-cult later
