// #[cfg(test)]
// mod tests {
//     #[test]
//     fn it_works() {
//         assert_eq!(2 + 2, 4);
//     }
// }

use services::service;

pub mod services;


pub fn say_hello() {
    println!("Hello, world!");
}
