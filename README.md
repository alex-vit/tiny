# tiny

A CLI to shrink images.

## Usage

```shell
tiny cat.png dog.jpg
```

Will shrink files in-place, saving originals in `cat_original.png` and similar.

## Building

### "Distributions"

```shell
./gradlew build # creates a bunch of stuff in ./build/distributions
```

### Standalone (graalvm)

```shell
./gradlew shadowJar # creates ./build/libs/tiny-$version-all.jar
native-image -jar build/libs/tiny-$version-all.jar tiny
```

## Todo

- [ ] folder support with _"Are you sure?_; `tiny .`
