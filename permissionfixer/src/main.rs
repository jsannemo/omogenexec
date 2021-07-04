extern crate clap;

use clap::Clap;

use std::{
    process::Command,
    fs::canonicalize,
};

const OMOGEN_PATH: &str = "/var/lib/omogen/";

#[derive(Clap)]
#[clap(name = "omogenexec-fixpermissions", about = "Resets permissions of files created by a sandbox user")]
struct Opt {
    #[clap(long)]
    path: String,
}

fn main() {
    let opt = Opt::parse();
    let path = canonicalize(format!("{}{}", OMOGEN_PATH, opt.path)).unwrap();
    if !path.starts_with(OMOGEN_PATH) {
        panic!("Can only take ownership of files under /var/lib/omogen/")
    }
    Command::new("/usr/bin/chattr")
        .arg("-i")
        .arg("-R")
        .arg(&path)
        .status()
        .expect("Failed to unmark immutability");
    Command::new("/bin/chown")
        .arg("-R")
        .arg("omogenexec-user:omogenexec-users")
        .arg(&path)
        .status()
        .expect("Failed to unmark immutability");
    Command::new("/bin/chmod")
        .arg("-R")
        .arg("gu+wrx")
        .arg(&path)
        .status()
        .expect("Failed to unmark immutability");
}
