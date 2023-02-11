#[macro_use]
extern crate bitmask;
extern crate cgroups_rs;
extern crate clap;
extern crate libc;
extern crate nix;
extern crate syscalls;
mod chroot;
mod libc_bindings;
mod quota;
mod sandbox;

use clap::Clap;
use libc_bindings::{find_group, find_user, pid_t, set_kill_on_parent_death, wait_for, set_rlimit};
use quota::set_quota;
use sandbox::{sandbox_main, Context};
use std::{
    path::{Path, PathBuf},
    process,
};

const SANDBOX_USER_PREFIX: &str = "omogenexec-user";
const SANDBOX_GROUP: &str = "omogenexec-users";
const CONTAINER_ROOT_PATH: &str = "/var/lib/omogen/sandbox";
const MAX_CONTAINERS: u32 = 100;

#[derive(Clap)]
#[clap(name = "omogenexec", about = "A low-level sandbox.")]
struct Opt {
    /// Sandbox ID: between 0 and 100
    #[clap(long)]
    sandbox_id: u32,

    /// The path inside the sandbox that stdin should be redirected to
    #[clap(long)]
    stdin: Option<String>,
    /// The path inside the sandbox that stdout should be redirected to
    #[clap(long)]
    stdout: Option<String>,
    /// The path inside the sandbox that stderr should be redirected to
    #[clap(long)]
    stderr: Option<String>,

    /// List of paths that should be read-only in the sandbox
    #[clap(long)]
    readable: Vec<String>,
    /// List of paths that should be read-write in the sandbox
    #[clap(long)]
    writable: Vec<String>,
    /// Working directory of the sandbox process
    #[clap(long, default_value = "/")]
    working_dir: String,
    #[clap(long)]
    no_default_mounts: bool,
    /// Environment variables for the sandbox process
    #[clap(long)]
    env: Vec<String>,

    /// The file system quota in blocks
    #[clap(long)]
    blocks: u64,
    /// The file system inode quota
    #[clap(long)]
    inodes: u64,
    /// The memory limit in megabytes.
    #[clap(long)]
    memory_mb: u64,
    /// The CPU time limit in milliseconds.
    #[clap(long)]
    time_lim_ms: u64,
    /// The wall-clock time limit in milliseconds.
    #[clap(long)]
    wall_time_lim_ms: u64,
    /// The maximum number of concurrent processes the sandboxed process may create
    #[clap(long)]
    pid_limit: i64,
}

// This is used for the clone'd inner sandbox process, which requires
// 16-byte alignment on some platforms.
#[repr(align(16))]
struct Stack([u8; 2 * 1024 * 1024]);

fn main() {
    // We always want to die if our parent dies to avoid possible races and left-over container
    // processes.
    set_kill_on_parent_death().unwrap();

    let opt = validated_opts().unwrap_or_else(|err| {
        eprintln!("Failed validating arguments: {}", err);
        process::exit(1);
    });
    let ctx = Context {
        sandbox_uid: find_user(format!("{}{}", SANDBOX_USER_PREFIX, opt.sandbox_id))
            .unwrap()
            .uid,
        sandbox_gid: find_group(SANDBOX_GROUP.to_string()).unwrap().gid,
        container_path: Path::new(CONTAINER_ROOT_PATH).join(opt.sandbox_id.to_string()),
        sandbox_id: opt.sandbox_id,
        stdin: opt.stdin.unwrap_or("".to_string()),
        stdout: opt.stdout.unwrap_or("".to_string()),
        stderr: opt.stderr.unwrap_or("".to_string()),
        env: opt.env,
        readable: opt.readable,
        writable: opt.writable,
        working_directory: PathBuf::from(opt.working_dir),
        mem_limit_bytes: opt.memory_mb as i64 * 1024 * 1024,
        time_lim: std::time::Duration::from_millis(opt.time_lim_ms),
        wall_time_lim: std::time::Duration::from_millis(opt.wall_time_lim_ms),
        pid_limit: opt.pid_limit,
        default_mounts: !opt.no_default_mounts,
    };

    set_quota(opt.blocks, opt.inodes, ctx.sandbox_uid);
    set_rlimit(libc::RLIMIT_STACK, libc::RLIM_INFINITY, libc::RLIM_INFINITY).unwrap();
    let ref mut stack = Stack([0; 2 * 1024 * 1024]);
    let mut cloned = ctx.clone();
    let sandbox_pid = unsafe {
        libc::clone(
            clone_main,
            stack.0.as_mut_ptr() as *mut libc::c_void,
            libc::CLONE_NEWNS
                | libc::CLONE_NEWPID
                | libc::CLONE_NEWNET
                | libc::CLONE_NEWIPC
                | libc::SIGCHLD,
            (&mut cloned as *mut Context) as *mut libc::c_void,
        )
    };
    // Note: no printlns from here on to avoid sync issues
    monitor_main(&ctx, sandbox_pid);
}

pub extern "C" fn clone_main(arg: *mut libc::c_void) -> libc::c_int {
    let ctx = unsafe { (&*(arg as *mut Context)).clone() };
    sandbox_main(ctx);
    0
}

fn validated_opts() -> Result<Opt, String> {
    let opt = Opt::parse();
    if opt.sandbox_id >= MAX_CONTAINERS {
        return Err(format!(
            "Sandbox ID was {}, must be at most {}",
            opt.sandbox_id,
            MAX_CONTAINERS - 1
        ));
    }
    Ok(opt)
}

fn monitor_main(ctx: &Context, sandbox_pid: pid_t) {
    let wait_res = wait_for(sandbox_pid);
    let cleanup_res = std::fs::remove_dir_all(&ctx.container_path);
    wait_res.unwrap();
    cleanup_res.unwrap();
}
