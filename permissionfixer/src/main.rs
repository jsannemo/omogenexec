extern crate clap;

use clap::Clap;

use std::{
    process::Command,
};

const SUBMISSION_PATH: &str = "/var/lib/omogen/submissions";

#[derive(Clap)]
#[clap(name = "omogenexec-fixpermissions", about = "Resets permissions of files created by a sandbox user")]
struct Opt {
    #[clap(long)]
    submission_id: u32,
}

fn main() {
    let opt = Opt::parse();
    let submission_path = format!("{}{}", SUBMISSION_PATH, opt.submission_id);
    Command::new("/usr/bin/chattr")
        .arg("-i")
        .arg("-R")
        .arg(&submission_path)
        .status()
        .expect("Failed to unmark immutability");
    Command::new("/bin/chown")
        .arg("-R")
        .arg("omogenexec-user:omogenexec-users")
        .arg(&submission_path)
        .status()
        .expect("Failed to unmark immutability");
    Command::new("/bin/chmod")
        .arg("-R")
        .arg("gu+wrx")
        .arg(&submission_path)
        .status()
        .expect("Failed to unmark immutability");
}
