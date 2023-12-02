use advent_of_code_2023_1::{*};

use std::fs::File;
use std::io::BufRead;
use std::io::BufReader;

fn get_calibration_value(line : String) -> Option<u32> {
    let first_numer = get_first_numer(&line);
    let last_numer = get_last_numer(&line);
    if first_numer.is_some() && last_numer.is_some() {
        return Some(first_numer.unwrap() * 10 + last_numer.unwrap())
    }
    None
}

fn main() {
    let file = File::open("problems/2023/1/answers/input.txt").unwrap();
    let reader = BufReader::new(file);
    
    let result : u32 = reader.lines()
        .map(|line| {
            if let Ok(line) = line {
                if let Some(calibration_value) = get_calibration_value(line) {
                    return calibration_value
                }
            }
            return 0
        })
        .sum();
    println!("solution1: {result}");
}
