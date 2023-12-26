import kotlinx.serialization.SerialName
import kotlinx.serialization.Serializable

@Serializable
data class ShrinkResp(
    @SerialName("output") val output: Output,
) {
    @Serializable
    data class Output(
        @SerialName("url") val url: String,
        @SerialName("ratio") val ratio: Double,
    )
}