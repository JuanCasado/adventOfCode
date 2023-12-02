
pub fn get_first_numer(line : &String) -> Option<u32> {
    for character in line.chars() {
        if character.is_numeric() {
            return character.to_digit(10)
        }
    }
    None
}

pub fn get_last_numer(line : &String) -> Option<u32> {
    for character in line.chars().rev() {
        if character.is_numeric() {
            return character.to_digit(10)
        }
    }
    None
}
