pub mod transport {
    pub mod v1 {
        include!(concat!(env!("OUT_DIR"), "/transport.v1.rs"));
    }
}

pub mod syncspace {
    pub mod v1 {
        include!(concat!(env!("OUT_DIR"), "/syncspace.v1.rs"));
    }
}
