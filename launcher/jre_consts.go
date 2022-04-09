package launcher

import "time"

// javaRuntimesURL is the URL to the Java runtimes manifest.
const javaRuntimesURL = "https://launchermeta.mojang.com/v1/products/java-runtime/2ec0cc96c44e5a76b9c8b7c39df7210883d12871/all.json"

// javaRuntimesManifestName is the name of the Java runtimes manifest file.
const javaRuntimesManifestName = "runtimes.json"

// javaRuntimesManifestTTL is the time after which the Java runtimes manifest should be re-fetched.
const javaRuntimesManifestTTL = time.Hour * 1
