use cfg_aliases::cfg_aliases;

fn main() {
    cfg_aliases! {
        wasm: { target_arch = "wasm32" },
        std: { feature = "std" },
        onchain: { feature = "onchain" }
    }
}
