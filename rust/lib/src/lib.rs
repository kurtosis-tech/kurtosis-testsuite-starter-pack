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

pub fn say_hello() {
    println!("Hello, world!");
}
