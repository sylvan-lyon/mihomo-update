fn main() {
    // watch ./locales/
    println!("cargo:rerun-if-changed=locales/");
}
