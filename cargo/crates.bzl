"""
@generated
cargo-raze generated Bazel file.

DO NOT EDIT! Replaced on runs of cargo-raze
"""

load("@bazel_tools//tools/build_defs/repo:git.bzl", "new_git_repository")  # buildifier: disable=load
load("@bazel_tools//tools/build_defs/repo:http.bzl", "http_archive")  # buildifier: disable=load
load("@bazel_tools//tools/build_defs/repo:utils.bzl", "maybe")  # buildifier: disable=load

def raze_fetch_remote_crates():
    """This function defines a collection of repos and should be called in a WORKSPACE file"""
    maybe(
        http_archive,
        name = "raze__aho_corasick__0_7_18",
        url = "https://crates.io/api/v1/crates/aho-corasick/0.7.18/download",
        type = "tar.gz",
        sha256 = "1e37cfd5e7657ada45f742d6e99ca5788580b5c529dc78faf11ece6dc702656f",
        strip_prefix = "aho-corasick-0.7.18",
        build_file = Label("//cargo/remote:BUILD.aho-corasick-0.7.18.bazel"),
    )

    maybe(
        http_archive,
        name = "raze__atty__0_2_14",
        url = "https://crates.io/api/v1/crates/atty/0.2.14/download",
        type = "tar.gz",
        sha256 = "d9b39be18770d11421cdb1b9947a45dd3f37e93092cbf377614828a319d5fee8",
        strip_prefix = "atty-0.2.14",
        build_file = Label("//cargo/remote:BUILD.atty-0.2.14.bazel"),
    )

    maybe(
        http_archive,
        name = "raze__autocfg__1_0_1",
        url = "https://crates.io/api/v1/crates/autocfg/1.0.1/download",
        type = "tar.gz",
        sha256 = "cdb031dd78e28731d87d56cc8ffef4a8f36ca26c38fe2de700543e627f8a464a",
        strip_prefix = "autocfg-1.0.1",
        build_file = Label("//cargo/remote:BUILD.autocfg-1.0.1.bazel"),
    )

    maybe(
        http_archive,
        name = "raze__bitflags__0_4_0",
        url = "https://crates.io/api/v1/crates/bitflags/0.4.0/download",
        type = "tar.gz",
        sha256 = "8dead7461c1127cf637931a1e50934eb6eee8bff2f74433ac7909e9afcee04a3",
        strip_prefix = "bitflags-0.4.0",
        build_file = Label("//cargo/remote:BUILD.bitflags-0.4.0.bazel"),
    )

    maybe(
        http_archive,
        name = "raze__bitflags__1_2_1",
        url = "https://crates.io/api/v1/crates/bitflags/1.2.1/download",
        type = "tar.gz",
        sha256 = "cf1de2fe8c75bc145a2f577add951f8134889b4795d47466a54a5c846d691693",
        strip_prefix = "bitflags-1.2.1",
        build_file = Label("//cargo/remote:BUILD.bitflags-1.2.1.bazel"),
    )

    maybe(
        http_archive,
        name = "raze__bitmask__0_5_0",
        url = "https://crates.io/api/v1/crates/bitmask/0.5.0/download",
        type = "tar.gz",
        sha256 = "5da9b3d9f6f585199287a473f4f8dfab6566cf827d15c00c219f53c645687ead",
        strip_prefix = "bitmask-0.5.0",
        build_file = Label("//cargo/remote:BUILD.bitmask-0.5.0.bazel"),
    )

    maybe(
        http_archive,
        name = "raze__cc__1_0_68",
        url = "https://crates.io/api/v1/crates/cc/1.0.68/download",
        type = "tar.gz",
        sha256 = "4a72c244c1ff497a746a7e1fb3d14bd08420ecda70c8f25c7112f2781652d787",
        strip_prefix = "cc-1.0.68",
        build_file = Label("//cargo/remote:BUILD.cc-1.0.68.bazel"),
    )

    maybe(
        http_archive,
        name = "raze__cfg_if__0_1_10",
        url = "https://crates.io/api/v1/crates/cfg-if/0.1.10/download",
        type = "tar.gz",
        sha256 = "4785bdd1c96b2a846b2bd7cc02e86b6b3dbf14e7e53446c4f54c92a361040822",
        strip_prefix = "cfg-if-0.1.10",
        build_file = Label("//cargo/remote:BUILD.cfg-if-0.1.10.bazel"),
    )

    maybe(
        http_archive,
        name = "raze__cfg_if__1_0_0",
        url = "https://crates.io/api/v1/crates/cfg-if/1.0.0/download",
        type = "tar.gz",
        sha256 = "baf1de4339761588bc0619e3cbc0120ee582ebb74b53b4efbf79117bd2da40fd",
        strip_prefix = "cfg-if-1.0.0",
        build_file = Label("//cargo/remote:BUILD.cfg-if-1.0.0.bazel"),
    )

    maybe(
        http_archive,
        name = "raze__cgroups_rs__0_2_6",
        url = "https://crates.io/api/v1/crates/cgroups-rs/0.2.6/download",
        type = "tar.gz",
        sha256 = "5c5c9f6e5c72958dc962baa5f8bb37fb611017854b0d774b8adab4d7416ab445",
        strip_prefix = "cgroups-rs-0.2.6",
        build_file = Label("//cargo/remote:BUILD.cgroups-rs-0.2.6.bazel"),
    )

    maybe(
        http_archive,
        name = "raze__clap__3_0_0_beta_2",
        url = "https://crates.io/api/v1/crates/clap/3.0.0-beta.2/download",
        type = "tar.gz",
        sha256 = "4bd1061998a501ee7d4b6d449020df3266ca3124b941ec56cf2005c3779ca142",
        strip_prefix = "clap-3.0.0-beta.2",
        build_file = Label("//cargo/remote:BUILD.clap-3.0.0-beta.2.bazel"),
    )

    maybe(
        http_archive,
        name = "raze__clap_derive__3_0_0_beta_2",
        url = "https://crates.io/api/v1/crates/clap_derive/3.0.0-beta.2/download",
        type = "tar.gz",
        sha256 = "370f715b81112975b1b69db93e0b56ea4cd4e5002ac43b2da8474106a54096a1",
        strip_prefix = "clap_derive-3.0.0-beta.2",
        build_file = Label("//cargo/remote:BUILD.clap_derive-3.0.0-beta.2.bazel"),
    )

    maybe(
        http_archive,
        name = "raze__err_derive__0_2_4",
        url = "https://crates.io/api/v1/crates/err-derive/0.2.4/download",
        type = "tar.gz",
        sha256 = "22deed3a8124cff5fa835713fa105621e43bbdc46690c3a6b68328a012d350d4",
        strip_prefix = "err-derive-0.2.4",
        build_file = Label("//cargo/remote:BUILD.err-derive-0.2.4.bazel"),
    )

    maybe(
        http_archive,
        name = "raze__hashbrown__0_11_2",
        url = "https://crates.io/api/v1/crates/hashbrown/0.11.2/download",
        type = "tar.gz",
        sha256 = "ab5ef0d4909ef3724cc8cce6ccc8572c5c817592e9285f5464f8e86f8bd3726e",
        strip_prefix = "hashbrown-0.11.2",
        build_file = Label("//cargo/remote:BUILD.hashbrown-0.11.2.bazel"),
    )

    maybe(
        http_archive,
        name = "raze__heck__0_3_3",
        url = "https://crates.io/api/v1/crates/heck/0.3.3/download",
        type = "tar.gz",
        sha256 = "6d621efb26863f0e9924c6ac577e8275e5e6b77455db64ffa6c65c904e9e132c",
        strip_prefix = "heck-0.3.3",
        build_file = Label("//cargo/remote:BUILD.heck-0.3.3.bazel"),
    )

    maybe(
        http_archive,
        name = "raze__hermit_abi__0_1_19",
        url = "https://crates.io/api/v1/crates/hermit-abi/0.1.19/download",
        type = "tar.gz",
        sha256 = "62b467343b94ba476dcb2500d242dadbb39557df889310ac77c5d99100aaac33",
        strip_prefix = "hermit-abi-0.1.19",
        build_file = Label("//cargo/remote:BUILD.hermit-abi-0.1.19.bazel"),
    )

    maybe(
        http_archive,
        name = "raze__indexmap__1_7_0",
        url = "https://crates.io/api/v1/crates/indexmap/1.7.0/download",
        type = "tar.gz",
        sha256 = "bc633605454125dec4b66843673f01c7df2b89479b32e0ed634e43a91cff62a5",
        strip_prefix = "indexmap-1.7.0",
        build_file = Label("//cargo/remote:BUILD.indexmap-1.7.0.bazel"),
    )

    maybe(
        http_archive,
        name = "raze__lazy_static__1_4_0",
        url = "https://crates.io/api/v1/crates/lazy_static/1.4.0/download",
        type = "tar.gz",
        sha256 = "e2abad23fbc42b3700f2f279844dc832adb2b2eb069b2df918f455c4e18cc646",
        strip_prefix = "lazy_static-1.4.0",
        build_file = Label("//cargo/remote:BUILD.lazy_static-1.4.0.bazel"),
    )

    maybe(
        http_archive,
        name = "raze__libc__0_2_97",
        url = "https://crates.io/api/v1/crates/libc/0.2.97/download",
        type = "tar.gz",
        sha256 = "12b8adadd720df158f4d70dfe7ccc6adb0472d7c55ca83445f6a5ab3e36f8fb6",
        strip_prefix = "libc-0.2.97",
        build_file = Label("//cargo/remote:BUILD.libc-0.2.97.bazel"),
    )

    maybe(
        http_archive,
        name = "raze__log__0_4_14",
        url = "https://crates.io/api/v1/crates/log/0.4.14/download",
        type = "tar.gz",
        sha256 = "51b9bbe6c47d51fc3e1a9b945965946b4c44142ab8792c50835a980d362c2710",
        strip_prefix = "log-0.4.14",
        build_file = Label("//cargo/remote:BUILD.log-0.4.14.bazel"),
    )

    maybe(
        http_archive,
        name = "raze__memchr__2_4_0",
        url = "https://crates.io/api/v1/crates/memchr/2.4.0/download",
        type = "tar.gz",
        sha256 = "b16bd47d9e329435e309c58469fe0791c2d0d1ba96ec0954152a5ae2b04387dc",
        strip_prefix = "memchr-2.4.0",
        build_file = Label("//cargo/remote:BUILD.memchr-2.4.0.bazel"),
    )

    maybe(
        http_archive,
        name = "raze__nix__0_20_0",
        url = "https://crates.io/api/v1/crates/nix/0.20.0/download",
        type = "tar.gz",
        sha256 = "fa9b4819da1bc61c0ea48b63b7bc8604064dd43013e7cc325df098d49cd7c18a",
        strip_prefix = "nix-0.20.0",
        build_file = Label("//cargo/remote:BUILD.nix-0.20.0.bazel"),
    )

    maybe(
        http_archive,
        name = "raze__nix__0_7_0",
        url = "https://crates.io/api/v1/crates/nix/0.7.0/download",
        type = "tar.gz",
        sha256 = "a0d95c5fa8b641c10ad0b8887454ebaafa3c92b5cd5350f8fc693adafd178e7b",
        strip_prefix = "nix-0.7.0",
        build_file = Label("//cargo/remote:BUILD.nix-0.7.0.bazel"),
    )

    maybe(
        http_archive,
        name = "raze__os_str_bytes__2_4_0",
        url = "https://crates.io/api/v1/crates/os_str_bytes/2.4.0/download",
        type = "tar.gz",
        sha256 = "afb2e1c3ee07430c2cf76151675e583e0f19985fa6efae47d6848a3e2c824f85",
        strip_prefix = "os_str_bytes-2.4.0",
        build_file = Label("//cargo/remote:BUILD.os_str_bytes-2.4.0.bazel"),
    )

    maybe(
        http_archive,
        name = "raze__partition_identity__0_2_8",
        url = "https://crates.io/api/v1/crates/partition-identity/0.2.8/download",
        type = "tar.gz",
        sha256 = "ec13ba9a0eec5c10a89f6ec1b6e9e2ef7d29b810d771355abbd1c43cae003ed6",
        strip_prefix = "partition-identity-0.2.8",
        build_file = Label("//cargo/remote:BUILD.partition-identity-0.2.8.bazel"),
    )

    maybe(
        http_archive,
        name = "raze__proc_macro_error__1_0_4",
        url = "https://crates.io/api/v1/crates/proc-macro-error/1.0.4/download",
        type = "tar.gz",
        sha256 = "da25490ff9892aab3fcf7c36f08cfb902dd3e71ca0f9f9517bea02a73a5ce38c",
        strip_prefix = "proc-macro-error-1.0.4",
        build_file = Label("//cargo/remote:BUILD.proc-macro-error-1.0.4.bazel"),
    )

    maybe(
        http_archive,
        name = "raze__proc_macro_error_attr__1_0_4",
        url = "https://crates.io/api/v1/crates/proc-macro-error-attr/1.0.4/download",
        type = "tar.gz",
        sha256 = "a1be40180e52ecc98ad80b184934baf3d0d29f979574e439af5a55274b35f869",
        strip_prefix = "proc-macro-error-attr-1.0.4",
        build_file = Label("//cargo/remote:BUILD.proc-macro-error-attr-1.0.4.bazel"),
    )

    maybe(
        http_archive,
        name = "raze__proc_macro2__1_0_27",
        url = "https://crates.io/api/v1/crates/proc-macro2/1.0.27/download",
        type = "tar.gz",
        sha256 = "f0d8caf72986c1a598726adc988bb5984792ef84f5ee5aa50209145ee8077038",
        strip_prefix = "proc-macro2-1.0.27",
        build_file = Label("//cargo/remote:BUILD.proc-macro2-1.0.27.bazel"),
    )

    maybe(
        http_archive,
        name = "raze__proc_mounts__0_2_4",
        url = "https://crates.io/api/v1/crates/proc-mounts/0.2.4/download",
        type = "tar.gz",
        sha256 = "2ad7e9c8d1b8c20f16a84d61d7c4c0325a5837c1307a2491b509cd92fb4e4442",
        strip_prefix = "proc-mounts-0.2.4",
        build_file = Label("//cargo/remote:BUILD.proc-mounts-0.2.4.bazel"),
    )

    maybe(
        http_archive,
        name = "raze__quote__1_0_9",
        url = "https://crates.io/api/v1/crates/quote/1.0.9/download",
        type = "tar.gz",
        sha256 = "c3d0b9745dc2debf507c8422de05d7226cc1f0644216dfdfead988f9b1ab32a7",
        strip_prefix = "quote-1.0.9",
        build_file = Label("//cargo/remote:BUILD.quote-1.0.9.bazel"),
    )

    maybe(
        http_archive,
        name = "raze__regex__1_5_4",
        url = "https://crates.io/api/v1/crates/regex/1.5.4/download",
        type = "tar.gz",
        sha256 = "d07a8629359eb56f1e2fb1652bb04212c072a87ba68546a04065d525673ac461",
        strip_prefix = "regex-1.5.4",
        build_file = Label("//cargo/remote:BUILD.regex-1.5.4.bazel"),
    )

    maybe(
        http_archive,
        name = "raze__regex_syntax__0_6_25",
        url = "https://crates.io/api/v1/crates/regex-syntax/0.6.25/download",
        type = "tar.gz",
        sha256 = "f497285884f3fcff424ffc933e56d7cbca511def0c9831a7f9b5f6153e3cc89b",
        strip_prefix = "regex-syntax-0.6.25",
        build_file = Label("//cargo/remote:BUILD.regex-syntax-0.6.25.bazel"),
    )

    maybe(
        http_archive,
        name = "raze__rustc_version__0_1_7",
        url = "https://crates.io/api/v1/crates/rustc_version/0.1.7/download",
        type = "tar.gz",
        sha256 = "c5f5376ea5e30ce23c03eb77cbe4962b988deead10910c372b226388b594c084",
        strip_prefix = "rustc_version-0.1.7",
        build_file = Label("//cargo/remote:BUILD.rustc_version-0.1.7.bazel"),
    )

    maybe(
        http_archive,
        name = "raze__rustversion__1_0_5",
        url = "https://crates.io/api/v1/crates/rustversion/1.0.5/download",
        type = "tar.gz",
        sha256 = "61b3909d758bb75c79f23d4736fac9433868679d3ad2ea7a61e3c25cfda9a088",
        strip_prefix = "rustversion-1.0.5",
        build_file = Label("//cargo/remote:BUILD.rustversion-1.0.5.bazel"),
    )

    maybe(
        http_archive,
        name = "raze__semver__0_1_20",
        url = "https://crates.io/api/v1/crates/semver/0.1.20/download",
        type = "tar.gz",
        sha256 = "d4f410fedcf71af0345d7607d246e7ad15faaadd49d240ee3b24e5dc21a820ac",
        strip_prefix = "semver-0.1.20",
        build_file = Label("//cargo/remote:BUILD.semver-0.1.20.bazel"),
    )

    maybe(
        http_archive,
        name = "raze__strsim__0_10_0",
        url = "https://crates.io/api/v1/crates/strsim/0.10.0/download",
        type = "tar.gz",
        sha256 = "73473c0e59e6d5812c5dfe2a064a6444949f089e20eec9a2e5506596494e4623",
        strip_prefix = "strsim-0.10.0",
        build_file = Label("//cargo/remote:BUILD.strsim-0.10.0.bazel"),
    )

    maybe(
        http_archive,
        name = "raze__syn__1_0_73",
        url = "https://crates.io/api/v1/crates/syn/1.0.73/download",
        type = "tar.gz",
        sha256 = "f71489ff30030d2ae598524f61326b902466f72a0fb1a8564c001cc63425bcc7",
        strip_prefix = "syn-1.0.73",
        build_file = Label("//cargo/remote:BUILD.syn-1.0.73.bazel"),
    )

    maybe(
        http_archive,
        name = "raze__synstructure__0_12_4",
        url = "https://crates.io/api/v1/crates/synstructure/0.12.4/download",
        type = "tar.gz",
        sha256 = "b834f2d66f734cb897113e34aaff2f1ab4719ca946f9a7358dba8f8064148701",
        strip_prefix = "synstructure-0.12.4",
        build_file = Label("//cargo/remote:BUILD.synstructure-0.12.4.bazel"),
    )

    maybe(
        http_archive,
        name = "raze__syscalls__0_3_3",
        url = "https://crates.io/api/v1/crates/syscalls/0.3.3/download",
        type = "tar.gz",
        sha256 = "9ad4126c98e506c5c2ada914b40c17b57a99f7d7044e35ff631a3504effb28e4",
        strip_prefix = "syscalls-0.3.3",
        build_file = Label("//cargo/remote:BUILD.syscalls-0.3.3.bazel"),
    )

    maybe(
        http_archive,
        name = "raze__termcolor__1_1_2",
        url = "https://crates.io/api/v1/crates/termcolor/1.1.2/download",
        type = "tar.gz",
        sha256 = "2dfed899f0eb03f32ee8c6a0aabdb8a7949659e3466561fc0adf54e26d88c5f4",
        strip_prefix = "termcolor-1.1.2",
        build_file = Label("//cargo/remote:BUILD.termcolor-1.1.2.bazel"),
    )

    maybe(
        http_archive,
        name = "raze__textwrap__0_12_1",
        url = "https://crates.io/api/v1/crates/textwrap/0.12.1/download",
        type = "tar.gz",
        sha256 = "203008d98caf094106cfaba70acfed15e18ed3ddb7d94e49baec153a2b462789",
        strip_prefix = "textwrap-0.12.1",
        build_file = Label("//cargo/remote:BUILD.textwrap-0.12.1.bazel"),
    )

    maybe(
        http_archive,
        name = "raze__unicode_segmentation__1_7_1",
        url = "https://crates.io/api/v1/crates/unicode-segmentation/1.7.1/download",
        type = "tar.gz",
        sha256 = "bb0d2e7be6ae3a5fa87eed5fb451aff96f2573d2694942e40543ae0bbe19c796",
        strip_prefix = "unicode-segmentation-1.7.1",
        build_file = Label("//cargo/remote:BUILD.unicode-segmentation-1.7.1.bazel"),
    )

    maybe(
        http_archive,
        name = "raze__unicode_width__0_1_8",
        url = "https://crates.io/api/v1/crates/unicode-width/0.1.8/download",
        type = "tar.gz",
        sha256 = "9337591893a19b88d8d87f2cec1e73fad5cdfd10e5a6f349f498ad6ea2ffb1e3",
        strip_prefix = "unicode-width-0.1.8",
        build_file = Label("//cargo/remote:BUILD.unicode-width-0.1.8.bazel"),
    )

    maybe(
        http_archive,
        name = "raze__unicode_xid__0_2_2",
        url = "https://crates.io/api/v1/crates/unicode-xid/0.2.2/download",
        type = "tar.gz",
        sha256 = "8ccb82d61f80a663efe1f787a51b16b5a51e3314d6ac365b08639f52387b33f3",
        strip_prefix = "unicode-xid-0.2.2",
        build_file = Label("//cargo/remote:BUILD.unicode-xid-0.2.2.bazel"),
    )

    maybe(
        http_archive,
        name = "raze__vec_map__0_8_2",
        url = "https://crates.io/api/v1/crates/vec_map/0.8.2/download",
        type = "tar.gz",
        sha256 = "f1bddf1187be692e79c5ffeab891132dfb0f236ed36a43c7ed39f1165ee20191",
        strip_prefix = "vec_map-0.8.2",
        build_file = Label("//cargo/remote:BUILD.vec_map-0.8.2.bazel"),
    )

    maybe(
        http_archive,
        name = "raze__version_check__0_9_3",
        url = "https://crates.io/api/v1/crates/version_check/0.9.3/download",
        type = "tar.gz",
        sha256 = "5fecdca9a5291cc2b8dcf7dc02453fee791a280f3743cb0905f8822ae463b3fe",
        strip_prefix = "version_check-0.9.3",
        build_file = Label("//cargo/remote:BUILD.version_check-0.9.3.bazel"),
    )

    maybe(
        http_archive,
        name = "raze__void__1_0_2",
        url = "https://crates.io/api/v1/crates/void/1.0.2/download",
        type = "tar.gz",
        sha256 = "6a02e4885ed3bc0f2de90ea6dd45ebcbb66dacffe03547fadbb0eeae2770887d",
        strip_prefix = "void-1.0.2",
        build_file = Label("//cargo/remote:BUILD.void-1.0.2.bazel"),
    )

    maybe(
        http_archive,
        name = "raze__winapi__0_3_9",
        url = "https://crates.io/api/v1/crates/winapi/0.3.9/download",
        type = "tar.gz",
        sha256 = "5c839a674fcd7a98952e593242ea400abe93992746761e38641405d28b00f419",
        strip_prefix = "winapi-0.3.9",
        build_file = Label("//cargo/remote:BUILD.winapi-0.3.9.bazel"),
    )

    maybe(
        http_archive,
        name = "raze__winapi_i686_pc_windows_gnu__0_4_0",
        url = "https://crates.io/api/v1/crates/winapi-i686-pc-windows-gnu/0.4.0/download",
        type = "tar.gz",
        sha256 = "ac3b87c63620426dd9b991e5ce0329eff545bccbbb34f3be09ff6fb6ab51b7b6",
        strip_prefix = "winapi-i686-pc-windows-gnu-0.4.0",
        build_file = Label("//cargo/remote:BUILD.winapi-i686-pc-windows-gnu-0.4.0.bazel"),
    )

    maybe(
        http_archive,
        name = "raze__winapi_util__0_1_5",
        url = "https://crates.io/api/v1/crates/winapi-util/0.1.5/download",
        type = "tar.gz",
        sha256 = "70ec6ce85bb158151cae5e5c87f95a8e97d2c0c4b001223f33a334e3ce5de178",
        strip_prefix = "winapi-util-0.1.5",
        build_file = Label("//cargo/remote:BUILD.winapi-util-0.1.5.bazel"),
    )

    maybe(
        http_archive,
        name = "raze__winapi_x86_64_pc_windows_gnu__0_4_0",
        url = "https://crates.io/api/v1/crates/winapi-x86_64-pc-windows-gnu/0.4.0/download",
        type = "tar.gz",
        sha256 = "712e227841d057c1ee1cd2fb22fa7e5a5461ae8e48fa2ca79ec42cfc1931183f",
        strip_prefix = "winapi-x86_64-pc-windows-gnu-0.4.0",
        build_file = Label("//cargo/remote:BUILD.winapi-x86_64-pc-windows-gnu-0.4.0.bazel"),
    )
