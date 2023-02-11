use libc::{dqblk, Q_SETQUOTA, QCMD, QIF_LIMITS, quotactl, SYS_close_range, USRQUOTA};
pub use libc::dev_t;
pub use libc::FILE;
pub use libc::gid_t;
pub use libc::pid_t;
pub use libc::stat;
pub use libc::uid_t;
use std::ffi::CString;
use std::io::Error;
use std::os::unix::ffi::OsStrExt;
use std::path::Path;
use std::ptr;
use std::result::Result;

extern "C" {
    pub static mut stdin: *mut FILE;
    pub static mut stdout: *mut FILE;
    pub static mut stderr: *mut FILE;
}

pub fn set_kill_on_parent_death() -> Result<(), String> {
    let ret = unsafe { libc::prctl(libc::PR_SET_PDEATHSIG, libc::SIGKILL) };
    if ret == 0 {
        Ok(())
    } else {
        Err(format!("prctl: {:?}", Error::last_os_error()))
    }
}

pub enum ForkProcess {
    Parent(libc::pid_t),
    Child,
}

pub fn fork() -> Result<ForkProcess, String> {
    let ret = unsafe { libc::fork() };
    if ret == -1 {
        Err(format!("fork: {:?}", Error::last_os_error()))
    } else if ret == 0 {
        Ok(ForkProcess::Child)
    } else {
        Ok(ForkProcess::Parent(ret))
    }
}

pub fn wait_for(pid: libc::pid_t) -> Result<libc::c_int, String> {
    let mut status = 0;
    let ret = unsafe { libc::waitpid(pid, &mut status as *mut i32, 0) };
    if ret == -1 {
        Err(format!("wait: {:?}", Error::last_os_error()))
    } else {
        Ok(status)
    }
}

pub fn wait_for_nohang(pid: libc::pid_t) -> Result<Option<libc::c_int>, String> {
    let mut status = 0;
    let ret = unsafe { libc::waitpid(pid, &mut status as *mut i32, libc::WNOHANG) };
    if ret == -1 {
        Err(format!("wait: {:?}", Error::last_os_error()))
    } else if ret == 0 {
        Ok(None)
    } else {
        Ok(Some(status))
    }
}

pub fn wait_any_nohang() -> Result<Option<libc::c_int>, i32> {
    let mut status = 0;
    let ret = unsafe { libc::waitpid(-1, &mut status as *mut i32, libc::WNOHANG) };
    if ret == -1 {
        if let Some(err) = Error::last_os_error().raw_os_error() {
            Err(err)
        } else {
            panic!("got error that was not OS error");
        }
    } else if ret == 0 {
        Ok(None)
    } else {
        Ok(Some(status))
    }
}

pub fn kill(pid: libc::pid_t) -> Result<(), String> {
    let ret = unsafe { libc::kill(pid, libc::SIGKILL) };
    if ret == -1 {
        Err(format!("kill: {:?}", Error::last_os_error()))
    } else {
        Ok(())
    }
}

pub fn chroot(path: &Path) -> Result<(), String> {
    let ret = unsafe {
        #[allow(temporary_cstring_as_ptr)]
        libc::chroot(CString::new(path.as_os_str().as_bytes()).unwrap().as_ptr())
    };
    if ret == 0 {
        Ok(())
    } else {
        Err(format!("chroot: {:?}", Error::last_os_error()))
    }
}

pub fn chdir(path: &Path) -> Result<(), String> {
    let ret = unsafe {
        #[allow(temporary_cstring_as_ptr)]
        libc::chdir(CString::new(path.as_os_str().as_bytes()).unwrap().as_ptr())
    };
    if ret == 0 {
        Ok(())
    } else {
        Err(format!("chdir: {:?}", Error::last_os_error()))
    }
}

pub fn exec(executable: String, mut args: Vec<String>, mut env: Vec<String>) -> Result<(), String> {
    let mut args_cstrs = Vec::with_capacity(args.len());
    let mut args_pointers = Vec::with_capacity(args.len());
    args_cstrs.push(CString::new(executable.clone()).unwrap());
    args_pointers.push(args_cstrs.last().unwrap().as_ptr());
    for arg in args.drain(..) {
        args_cstrs.push(CString::new(arg).unwrap());
        args_pointers.push(args_cstrs.last().unwrap().as_ptr());
    }
    args_pointers.push(ptr::null());

    let mut env_cstrs = Vec::with_capacity(env.len());
    let mut env_pointers = Vec::with_capacity(env.len());
    env_cstrs.push(CString::new(executable.clone()).unwrap());
    env_pointers.push(env_cstrs.last().unwrap().as_ptr());
    for e in env.drain(..) {
        env_cstrs.push(CString::new(e).unwrap());
        env_pointers.push(env_cstrs.last().unwrap().as_ptr());
    }
    env_pointers.push(ptr::null());
    unsafe {
        #[allow(temporary_cstring_as_ptr)]
        libc::execvpe(
            CString::new(executable).unwrap().as_ptr(),
            args_pointers.as_ptr(),
            env_pointers.as_ptr(),
        )
    };
    Err(format!("execvp: {:?}", Error::last_os_error()))
}

bitmask! {
    #[derive(Debug)]
    pub mask MountOptions: u64 where flags MountOption {
        Bind = libc::MS_BIND,
        NoSuid = libc::MS_NOSUID,
        NoDev = libc::MS_NODEV,
        ReadOnly = libc::MS_RDONLY,
        NoExec = libc::MS_NOEXEC,
        Private = libc::MS_PRIVATE,
        Rec = libc::MS_REC,
        Remount = libc::MS_REMOUNT,
    }
}

pub fn mount(source: &Path, target: &Path, opts: MountOptions) -> Result<(), String> {
    let ret = unsafe {
        #[allow(temporary_cstring_as_ptr)]
        libc::mount(
            CString::new(source.as_os_str().as_bytes())
                .unwrap()
                .as_ptr(),
            CString::new(target.as_os_str().as_bytes())
                .unwrap()
                .as_ptr(),
            ptr::null(),
            *opts,
            ptr::null(),
        )
    };
    if ret == 0 {
        Ok(())
    } else {
        Err(format!("mount: {:?}", Error::last_os_error()))
    }
}

pub fn mount_proc(source: &Path, target: &Path, opts: MountOptions) -> Result<(), String> {
    let ret = unsafe {
        #[allow(temporary_cstring_as_ptr)]
        libc::mount(
            CString::new(source.as_os_str().as_bytes())
                .unwrap()
                .as_ptr(),
            CString::new(target.as_os_str().as_bytes())
                .unwrap()
                .as_ptr(),
            CString::new("proc").unwrap().as_ptr(),
            *opts,
            ptr::null(),
        )
    };
    if ret == 0 {
        Ok(())
    } else {
        Err(format!("mount: {:?}", Error::last_os_error()))
    }
}

pub fn privatize_mounts() -> Result<(), String> {
    let ret = unsafe {
        #[allow(temporary_cstring_as_ptr)]
        libc::mount(
            ptr::null(),
            CString::new("/").unwrap().as_ptr(),
            ptr::null(),
            *(MountOption::Rec | MountOption::Private),
            ptr::null(),
        )
    };
    if ret == 0 {
        Ok(())
    } else {
        Err(format!("mount: {:?}", Error::last_os_error()))
    }
}

pub struct Passwd {
    pub uid: uid_t,
}

pub fn find_user(username: String) -> Result<Passwd, String> {
    reset_errno();
    #[allow(temporary_cstring_as_ptr)]
        let ret = unsafe { libc::getpwnam(CString::new(username).unwrap().as_ptr()) };
    if ret == ptr::null_mut() {
        Err(format!("getpwnam: {:?}", Error::last_os_error()))
    } else {
        unsafe { Ok(Passwd { uid: (*ret).pw_uid }) }
    }
}

pub struct Group {
    pub gid: gid_t,
}

pub fn find_group(group_name: String) -> Result<Group, String> {
    reset_errno();
    #[allow(temporary_cstring_as_ptr)]
        let ret = unsafe { libc::getgrnam(CString::new(group_name).unwrap().as_ptr()) };
    if ret == ptr::null_mut() {
        Err(format!("getgrnam: {:?}", Error::last_os_error()))
    } else {
        unsafe { Ok(Group { gid: (*ret).gr_gid }) }
    }
}

pub fn set_res_uid_and_gid(uid: uid_t, gid: gid_t) -> Result<(), String> {
    let gid_ret = unsafe { libc::setresgid(gid, gid, gid) };
    if gid_ret != 0 {
        return Err(format!("setresgid: {:?}", Error::last_os_error()));
    }
    let uid_ret = unsafe { libc::setresuid(uid, uid, uid) };
    if uid_ret != 0 {
        return Err(format!("setresuid: {:?}", Error::last_os_error()));
    }
    Ok(())
}

pub fn drop_groups() -> Result<(), String> {
    let ret = unsafe { libc::setgroups(0, ptr::null_mut()) };
    if ret == 0 {
        Ok(())
    } else {
        Err(format!("setgroups: {:?}", Error::last_os_error()))
    }
}

pub enum FileAccessMode {
    Readable,
    Writable,
}

pub fn repoint_stream(path: String, stream: *mut FILE, mode: FileAccessMode) -> Result<(), String> {
    let access = match mode {
        FileAccessMode::Readable => "r",
        FileAccessMode::Writable => "w",
    };
    #[allow(temporary_cstring_as_ptr)]
        let ret = unsafe {
        libc::freopen(
            CString::new(path).unwrap().as_ptr(),
            CString::new(access).unwrap().as_ptr(),
            stream,
        )
    };
    if ret == ptr::null_mut() {
        Err(format!("freopen: {:?}", Error::last_os_error()))
    } else {
        Ok(())
    }
}

pub fn fclose(stream: *mut FILE) -> Result<(), String> {
    let ret = unsafe { libc::fclose(stream) };
    if ret != 0 {
        Err(format!("fclose: {:?}", Error::last_os_error()))
    } else {
        Ok(())
    }
}

pub fn get_device(path: &Path) -> Result<dev_t, String> {
    let mut path_stat = std::mem::MaybeUninit::<stat>::uninit();
    let ret = unsafe {
        #[allow(temporary_cstring_as_ptr)]
        stat(
            CString::new(path.as_os_str().as_bytes()).unwrap().as_ptr(),
            path_stat.as_mut_ptr(),
        )
    };
    if ret != 0 {
        Err(format!("stat: {:?}", Error::last_os_error()))
    } else {
        unsafe { Ok(path_stat.assume_init().st_dev) }
    }
}

pub fn set_user_quota(
    block_quota: u64,
    inode_quota: u64,
    mount_dev: &Path,
    uid: uid_t,
) -> Result<(), String> {
    unsafe {
        let mut quota = std::mem::MaybeUninit::<dqblk>::uninit().assume_init();
        quota.dqb_bhardlimit = block_quota;
        quota.dqb_bsoftlimit = block_quota;
        quota.dqb_ihardlimit = inode_quota;
        quota.dqb_isoftlimit = inode_quota;
        quota.dqb_valid = QIF_LIMITS;
        #[allow(temporary_cstring_as_ptr)]
            let ret = quotactl(
            QCMD(Q_SETQUOTA, USRQUOTA),
            CString::new(mount_dev.as_os_str().as_bytes())
                .unwrap()
                .as_ptr(),
            uid as i32,
            (&mut quota as *mut dqblk) as *mut i8,
        );
        if ret == 0 {
            Ok(())
        } else {
            Err(format!("quotactl: {:?}", Error::last_os_error()))
        }
    }
}

pub fn make_closing_pipes() -> Result<(i32, i32), String> {
    let mut fds = [0; 2];
    let ret = unsafe { libc::pipe2(fds.as_mut_ptr(), libc::O_CLOEXEC) };
    if ret != 0 {
        Err(format!("pipe2: {:?}", Error::last_os_error()))
    } else {
        Ok((fds[0], fds[1]))
    }
}

pub fn close_nonstd_fds() -> Result<(), String> {
    let ret = unsafe { libc::syscall(SYS_close_range, 3, u32::max as i64, 0) };
    if ret == 0 {
        Ok(())
    } else {
        Err(format!("close_range: {:?}", Error::last_os_error()))
    }
}

pub fn set_rlimit(resource: libc::__rlimit_resource_t, soft: u64, hard: u64) -> Result<(), String>  {
    let ret = unsafe {
        let mut rlimit = std::mem::MaybeUninit::<libc::rlimit64>::uninit().assume_init();
        rlimit.rlim_cur = soft;
        rlimit.rlim_max = hard;
        libc::setrlimit64(resource, &rlimit)
    };
    if ret == 0 {
        Ok(())
    } else {
        Err(format!("setrlimit64: {:?}", Error::last_os_error()))
    }
}

fn reset_errno() {
    unsafe {
        *libc::__errno_location() = 0;
    }
}