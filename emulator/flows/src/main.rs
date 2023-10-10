// use questdb::{
//     Result,
//     ingress::{
//         Buffer,
//         SenderBuilder, TableName, ColumnName, TimestampNanos
//     }
// };

// use rand_distr::{Normal, Distribution};
// use std::convert::TryInto;
// use std::{thread, time};

use nell::ffi::tcp_info;
use sysinfo::{ProcessExt, System, SystemExt};
// use std::fs::File;
use std::{os::unix::io::FromRawFd};
use std::net::TcpStream;

use std::io::Error;

// use pidfd_getfd::get_file_from_pidfd;
use pidfd_getfd::pidfd_getfd;
use libc::{TCP_INFO, SOL_TCP, c_void};
use std::mem;
use std::os::unix::io::AsRawFd;
// fn main() -> Result<()> {
fn main() {
    let s = System::new_all();
    let flags = pidfd_getfd::GetFdFlags::empty();
    
    for process in s.processes_by_name("nc") {
        println!("process name: {} {}", process.name(), process.pid());
        // let fd = get_file_from_pidfd(6397, 3, flags);
        
        let res = unsafe {pidfd_getfd(6397, 3, flags.bits())};
        if res == -1 {
            println!("get_file_from_pidfd failed {:?}", res);
            println!("error: {:?}", Error::last_os_error())

        } else {
            let socket = unsafe { TcpStream::from_raw_fd(res)};
            let mut tcpinfo = nell::ffi::tcp_info::default();
            // let ptr = tcpinfo.as_ptr();
            // let mut ptr: *c_void = &tcpinfo;
            let tcp_size = mem::size_of::<nell::ffi::tcp_info>();
            
            // nell::sys::socket::getsockopt(&socket, nell::sys::socket::Level::SOCKET, nell::sys::socket::Name::);
            
            let result = unsafe {libc::getsockopt(socket.as_raw_fd(), libc::SOL_TCP, libc::TCP_INFO, &tcpinfo as *const nell::ffi::tcp_info as *mut libc::c_void, tcp_size as *mut u32)};

            println!("getsockopt: {:?}", result);
        }

        // letraw_fd = fd.into_raw_fd();
        
    }
    // let mut sender = SenderBuilder::new("100.113.13.30", 9009).connect()?;
    // let mut buffer = Buffer::new();
    // let table_name = TableName::new("sensors")?;
    // let column_temperature = ColumnName::new("temperature")?;
    // let column_humidity = ColumnName::new("humidity")?;

    // let humidity_normal = Normal::new(50.0, 10.0).unwrap();
    // let temperature_normal = Normal::new(25.0, 3.0).unwrap();

    // for i in 1..10001 {
    //     // let mut time = time::Instant::now();
    //     let ts: TimestampNanos = std::time::SystemTime::now().try_into()?;
    //     buffer.table(table_name)?.symbol("id", "karl")?.column_f64(column_temperature, temperature_normal.sample(&mut rand::thread_rng()))?.column_i64(column_humidity, humidity_normal.sample(&mut rand::thread_rng()) as i64)?.at(ts)?;
    //     if i % 50 == 0 {
    //         sender.flush(&mut buffer)?;
    //     }
    //     thread::sleep(time::Duration::from_millis(10));
    // }
    // Ok(())

}

