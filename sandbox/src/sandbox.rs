use cgroups_rs::*;
use chroot::{apply_chroot, make_mount, Mount, mount_procfs, read_only_copy_mount};
use libc_bindings::{
    close_nonstd_fds, drop_groups, exec, fclose, FileAccessMode, fork, ForkProcess, gid_t,
    kill, make_closing_pipes, privatize_mounts, repoint_stream, set_kill_on_parent_death, set_res_uid_and_gid,
    stderr, stdin, stdout, uid_t, wait_any_nohang, wait_for_nohang,
};
use std::{
    fs::File,
    io::{Read, Write},
    os::unix::io::FromRawFd,
    path::PathBuf,
    process,
};

#[derive(Clone)]
pub struct Context {
    pub sandbox_id: u32,
    pub sandbox_uid: uid_t,
    pub sandbox_gid: gid_t,
    pub container_path: PathBuf,
    pub stdin: String,
    pub stdout: String,
    pub stderr: String,
    pub readable: Vec<String>,
    pub writable: Vec<String>,
    pub working_directory: PathBuf,
    pub mem_limit_bytes: i64,
    pub time_lim: std::time::Duration,
    pub wall_time_lim: std::time::Duration,
    pub pid_limit: i64,
    pub default_mounts: bool,
}

fn setup_container_fs(ctx: &Context) {
    std::fs::create_dir_all(&ctx.container_path).unwrap();
    setup_mounts(ctx);
}

fn setup_mounts(ctx: &Context) {
    privatize_mounts().unwrap();

    mount_procfs(&ctx.container_path);

    if ctx.default_mounts {
        for path in vec!["/bin", "/usr/bin", "/usr/lib", "/lib"] {
            make_mount(
                &ctx.container_path,
                &read_only_copy_mount(PathBuf::from(path)),
            )
                .unwrap();
        }
    }
    for path in &ctx.readable {
        let paths = parse_mount_path(path);
        make_mount(
            &ctx.container_path,
            &Mount {
                writable: false,
                outside: paths.0,
                inside: paths.1,
            },
        )
            .unwrap();
    }
    for path in &ctx.writable {
        let paths = parse_mount_path(path);
        make_mount(
            &ctx.container_path,
            &Mount {
                writable: true,
                outside: paths.0,
                inside: paths.1,
            },
        )
            .unwrap();
    }
    if ctx.default_mounts {
        let usrlib32 = PathBuf::from("/usr/lib32");
        if usrlib32.exists() && usrlib32.is_dir() {
            make_mount(&ctx.container_path, &read_only_copy_mount(usrlib32)).unwrap();
        }
        let lib64 = PathBuf::from("/lib64");
        if lib64.exists() && lib64.is_dir() {
            make_mount(&ctx.container_path, &read_only_copy_mount(lib64)).unwrap();
        }
        let lib32 = PathBuf::from("/lib32");
        if lib32.exists() && lib32.is_dir() {
            make_mount(&ctx.container_path, &read_only_copy_mount(lib32)).unwrap();
        }
    }
}

fn parse_mount_path(path: &String) -> (PathBuf, PathBuf) {
    let parts = path.find(':');
    return match parts {
        None => (PathBuf::from(path), PathBuf::from(path.to_string())),
        Some(idx) => (
            PathBuf::from(path[..idx].to_string()),
            PathBuf::from(path[idx + 1..].to_string()),
        ),
    };
}

fn read_command() -> (String, Vec<String>) {
    let mut cmds = [0; 1];
    let n = std::io::stdin().read(&mut cmds[..]).unwrap();
    if n == 0 {
        return (String::new(), vec![]);
    }
    let mut exec = String::new();
    let mut args = vec![String::new(); (cmds[0] as usize) - 1];

    let mut buffer = [0; 1024];
    let mut at = 0;
    // Note: we only support getting one command written at a time.
    while at < cmds[0] {
        let mut append = if at == 0 {
            &mut exec
        } else {
            &mut args[(at as usize) - 1]
        };
        let n = std::io::stdin().read(&mut buffer[..]).unwrap();
        if n == 0 {
            eprintln!("command ended unexpectedly");
            process::exit(1);
        }
        for k in 0..n {
            if buffer[k] == 0 {
                at += 1;
                if at == cmds[0] {
                    break;
                }
                append = &mut args[(at as usize) - 1];
            } else {
                append.push(buffer[k] as char);
            }
        }
    }
    (exec, args)
}

fn setup_cgroups(ctx: &Context) -> Cgroup {
    let hier = cgroups_rs::hierarchies::auto();
    let cgroup_name = format!("omogen-{}", ctx.sandbox_id);
    cgroups_rs::cgroup_builder::CgroupBuilder::new(&cgroup_name).build(hier)
}

pub fn sandbox_main(ctx: Context) -> isize {
    set_kill_on_parent_death().unwrap();
    let cg = setup_cgroups(&ctx);
    let cg_mem: &cgroups_rs::memory::MemController = cg.controller_of().unwrap();
    let cg_cpu: &cgroups_rs::cpu::CpuController = cg.controller_of().unwrap();
    let cg_pid: &cgroups_rs::pid::PidController = cg.controller_of().unwrap();
    cg_mem.set_limit(ctx.mem_limit_bytes).unwrap();
    setup_container_fs(&ctx);
    loop {
        let cmd = read_command();
        if cmd.0.len() == 0 {
            eprintln!("no more commands, exiting");
            break;
        }
        close_nonstd_fds().unwrap();
        let pipes = make_closing_pipes().unwrap();
        eprintln!("cmd: {:?} {:?}", cmd.0, cmd.1);
        match fork().unwrap() {
            ForkProcess::Child => {
                apply_chroot(&ctx.container_path, &ctx.working_directory);
                drop_groups().unwrap();
                set_res_uid_and_gid(ctx.sandbox_uid, ctx.sandbox_gid).unwrap();
                // Close the read end
                unsafe { File::from_raw_fd(pipes.0) };
                setup_and_run(pipes.1, cmd.0, cmd.1, &ctx);
                process::exit(1);
            }
            ForkProcess::Parent(child) => {
                // Close the write end
                unsafe { File::from_raw_fd(pipes.1) };

                let err_pipe = pipes.0;
                let mut err_file = unsafe { File::from_raw_fd(err_pipe) };
                let mut s = String::new();
                cg_pid.set_pid_max(MaxValue::Value(ctx.pid_limit)).unwrap();
                cg_mem.add_task(&CgroupPid::from(child as u64)).unwrap();
                cg_cpu.add_task(&CgroupPid::from(child as u64)).unwrap();
                cg_pid.add_task(&CgroupPid::from(child as u64)).unwrap();
                // The program will start exec'ing immediately after we read this write, since the
                // write of pipes block on the corresponding read.
                let now = std::time::SystemTime::now();
                err_file.read_to_string(&mut s).unwrap();
                if s != "ok" {
                    println!("killed setup");
                    println!("done");
                    continue;
                }
                let mut sleep = 5;
                let cpu_before = cpu_stat_nanos(cg_cpu.cpu().stat);
                loop {
                    let maybe_exit = wait_for_nohang(child).unwrap();
                    match maybe_exit {
                        None => {
                            let cpu_nanos = cpu_stat_nanos(cg_cpu.cpu().stat) - cpu_before;
                            let cpu_time = std::time::Duration::new(
                                cpu_nanos / 1_000_000_000,
                                (cpu_nanos % 1_000_000_000) as u32,
                            );
                            let wall_time = now.elapsed().unwrap();
                            // Wait 1 extra second of CPU time for displaying close calls to judges
                            if wall_time > ctx.wall_time_lim || cpu_time > ctx.time_lim + std::time::Duration::from_secs(1) {
                                break;
                            }
                            std::thread::sleep(std::time::Duration::from_millis(sleep));
                            sleep = if 2 * sleep > 100 { 100 } else { sleep * 2 }
                        }
                        Some(exit) => {
                            if s == "err" {
                                eprintln!("failed to redirect streams in the sandbox");
                                process::exit(1);
                            } else if s == "okexec" {
                                eprintln!("failed to exec in the sandbox (see stderr)");
                                process::exit(1);
                            } else {
                                if libc::WIFEXITED(exit) {
                                    println!("code {:?}", libc::WEXITSTATUS(exit));
                                } else if libc::WIFSIGNALED(exit) {
                                    println!("signal {:?}", libc::WTERMSIG(exit));
                                } else if libc::WIFSTOPPED(exit) {
                                    println!("signal {:?}", libc::WSTOPSIG(exit));
                                }
                            }
                            break;
                        }
                    }
                }
                eprintln!("finished command, freezing cgroup");
                // Make sure no new pids can be created to kill fork bombs
                cg_pid.set_pid_max(MaxValue::Value(0)).unwrap();
                eprintln!("cgroup frozen, killing processes");
                // Kill all the processes in the cgroup.
                loop {
                    let tasks = cg_pid.tasks();
                    if tasks.len() == 0 {
                        break;
                    }
                    for pid in tasks {
                        kill(pid.pid as i32).unwrap();
                    }
                    std::thread::sleep(std::time::Duration::from_millis(50));
                }
                loop {
                    match wait_any_nohang() {
                        Ok(res) => {
                            eprintln!("done: {:?}", res);
                            continue;
                        }
                        Err(libc::ECHILD) => break,
                        Err(_) => panic!(
                            "unexpected wait error {:?}",
                            std::io::Error::last_os_error()
                        ),
                    }
                }
                let cpu_nanos = cpu_stat_nanos(cg_cpu.cpu().stat) - cpu_before;
                let cpu_time = std::time::Duration::new(
                    cpu_nanos / 1_000_000_000,
                    (cpu_nanos % 1_000_000_000) as u32,
                );
                if cpu_time > ctx.time_lim {
                    println!("killed tle");
                }
                // Nanos -> Millis
                println!("cpu {:?}", cpu_nanos / 1_000_000);
                println!("done");
            }
        }
        eprintln!("done with cmd: {:?} {:?}", cmd.0, cmd.1);
    }
    cg.delete().unwrap();
    0
}

fn cpu_stat_nanos(stat: String) -> u64 {
    for line in stat.split("\n") {
        let fields: Vec<&str> = line.split(' ').collect();
        if fields[0] == "usage_usec" {
            match fields[1].parse::<u64>() {
                Ok(res) => { return res * 1000; }
                Err(_) => panic!(
                    "failed parsing usage_usec!?"
                )
            }
        }
    }
    panic!("cpu.stat doesn't have usage_usec!?")
}

fn set_streams(ctx: &Context) -> Result<(), String> {
    unsafe {
        if ctx.stdin.len() == 0 {
            fclose(stdin)?;
        } else {
            repoint_stream(ctx.stdin.to_string(), stdin, FileAccessMode::Readable)?;
        }
        if ctx.stdout.len() == 0 {
            fclose(stdout)?;
        } else {
            repoint_stream(ctx.stdout.to_string(), stdout, FileAccessMode::Writable)?;
        }
        if ctx.stderr.len() == 0 {
            fclose(stderr)?;
        } else {
            repoint_stream(ctx.stderr.to_string(), stderr, FileAccessMode::Writable)?;
        }
    }
    Ok(())
}

fn setup_and_run(err_pipe: i32, cmd: String, args: Vec<String>, ctx: &Context) {
    let mut err_file = unsafe { File::from_raw_fd(err_pipe) };
    set_streams(ctx).unwrap_or_else(|err| {
        eprintln!("setup error: {:?}", err);
        write!(&mut err_file, "err").unwrap();
        process::exit(1);
    });
    write!(&mut err_file, "ok").unwrap();
    exec(cmd.clone(), args.clone()).unwrap_or_else(|err| {
        eprintln!("{:?}", err);
        write!(&mut err_file, "exec").unwrap();
        process::exit(1);
    });
}
