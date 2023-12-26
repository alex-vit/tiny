import kotlinx.coroutines.delay
import okhttp3.MediaType.Companion.toMediaType
import okhttp3.OkHttpClient
import okhttp3.Request
import okhttp3.RequestBody.Companion.asRequestBody
import okio.buffer
import okio.sink
import java.io.IOException
import java.nio.file.InvalidPathException
import java.nio.file.Path
import java.time.Duration
import kotlin.io.path.*
import kotlin.math.roundToInt
import kotlin.random.Random
import kotlin.random.nextLong

suspend fun main(args: Array<String>) {
    if (args.isEmpty()) {
        println(
            """
            Usage: 
              tiny cat.png dog.jpg # shrink supplied files
              tiny . # shrink images in folder
        """.trimIndent()
        )
        return
    }

    val firstPath = args[0].toPathOrNull()
    if (firstPath == null) {
        println("Invalid path: ${args[0]}")
        return
    }

    val imagePaths: List<Path>
    if (firstPath.isDirectory()) {
        if (args.size > 1) {
            println("First argument is a folder, ignoring the rest of args.")
        }
        imagePaths = firstPath.listDirectoryEntries().filter(Path::isImagePath)
    } else {
        imagePaths = args.map(Path::of).filter(Path::isImagePath)
    }

    for (path in imagePaths) {
        print("Shrinking $path... ")
        delay(Random.nextLong(500L..1000))

        val shrinkResp = postShrink(path)
        if (shrinkResp == null) {
            println("Failed upload")
            continue
        }
        if (!backUpFile(path)) {
            println("Failed backup")
            continue
        }
        if (!download(shrinkResp.output.url, path)) {
            println("Failed download")
            continue
        }

        val saved = (100 * (1 - shrinkResp.output.ratio)).roundToInt()
        println("OK! Saved $saved%")
    }
}

fun Path.isImagePath() = isRegularFile() && extension in listOf("jpg", "jpeg", "png")

fun String.toPathOrNull() =
    try {
        Path.of(this)
    } catch (_: InvalidPathException) {
        null
    }

private val client = OkHttpClient.Builder()
    .callTimeout(Duration.ofSeconds(10))
    .build()
private val postReqBuilder = Request.Builder()
    .url("https://tinyjpg.com/backend/opt/shrink")

private fun postShrink(path: Path): ShrinkResp? {
    val req = postReqBuilder
        .post(path.toFile().asRequestBody(mediaType(path)))
        .build()
    try {
        val response = client.newCall(req).execute()
        val respStr = response.body?.string() ?: return null
        return DefaultJson.decodeFromString<ShrinkResp>(respStr)
    } catch (_: IOException) {
        return null
    }
}

private fun mediaType(path: Path) = when (path.extension) {
    "jpg", "jpeg" -> "image/jpeg"
    // "png"
    else -> "image/png"
}.toMediaType()

private fun backUpFile(path: Path): Boolean {
    val backupPath = path.withNameSuffix("original")
    try {
        path.toFile().copyTo(backupPath.toFile(), overwrite = true)
        return true
    } catch (_: IOException) {
        return false
    }
}

/**
 * Return path with a suffix appended to file name. For example, given a "backup" suffix:
 * `~/dir/file.txt` --> `~/dir/file_backup.txt
 */
@Suppress("NAME_SHADOWING")
private fun Path.withNameSuffix(suffix: String): Path {
    val suffix = if (suffix.startsWith('_')) suffix else "_$suffix"
    val nameWithSuffix = fileName.nameWithoutExtension + suffix + "." + extension
    return Path.of(parent.pathString, nameWithSuffix)
}

private fun download(url: String, existingPath: Path): Boolean {
    val file = existingPath.toFile().sink().buffer()
    val request = Request.Builder().url(url).build()
    try {
        val response = client.newCall(request).execute()
        val body = response.body?.source() ?: return false
        file.writeAll(body)
        return true
    } catch (_: IOException) {
        return false
    } finally {
        file.close()
    }
}