// #[cfg(test)]
// mod tests {
//     #[test]
//     fn it_works() {
//         assert_eq!(2 + 2, 4);
//     }
// }

pub mod services;
pub mod execution;
pub mod testsuite;
pub mod core_api_bindings;

pub fn say_hello() {
    println!("Hello, world!");
}
