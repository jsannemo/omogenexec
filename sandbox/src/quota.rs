extern crate proc_mounts;

use libc_bindings;
use std::path::PathBuf;
use CONTAINER_ROOT_PATH;

pub fn set_quota(block_quota: u64, inode_quota: u64, uid: libc_bindings::uid_t) {
    let sandbox_dev = libc_bindings::get_device(&PathBuf::from(CONTAINER_ROOT_PATH));
    for mount in proc_mounts::MountIter::new().unwrap() {
        let mount_ok = mount.unwrap();
        let mount_dev = libc_bindings::get_device(&mount_ok.dest);
        if mount_dev == sandbox_dev {
            libc_bindings::set_user_quota(block_quota, inode_quota, &mount_ok.source, uid).unwrap();
            return;
        }
    }
    panic!("Did not find mounted device")
}
