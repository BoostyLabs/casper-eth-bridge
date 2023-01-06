use std::convert::TryFrom;

use primitive_types::{U256, U512};

/// Predefined powers of 10.
#[rustfmt::skip]
const POWERS_OF_10: [[u64; 4]; 77] = [
    [0x1, 0x0, 0x0, 0x0],
    [0xa, 0x0, 0x0, 0x0],
    [0x64, 0x0, 0x0, 0x0],
    [0x3e8, 0x0, 0x0, 0x0],
    [0x2710, 0x0, 0x0, 0x0],
    [0x186a0, 0x0, 0x0, 0x0],
    [0xf4240, 0x0, 0x0, 0x0],
    [0x989680, 0x0, 0x0, 0x0],
    [0x5f5e100, 0x0, 0x0, 0x0],
    [0x3b9aca00, 0x0, 0x0, 0x0],
    [0x2540be400, 0x0, 0x0, 0x0],
    [0x174876e800, 0x0, 0x0, 0x0],
    [0xe8d4a51000, 0x0, 0x0, 0x0],
    [0x9184e72a000, 0x0, 0x0, 0x0],
    [0x5af3107a4000, 0x0, 0x0, 0x0],
    [0x38d7ea4c68000, 0x0, 0x0, 0x0],
    [0x2386f26fc10000, 0x0, 0x0, 0x0],
    [0x16345785d8a0000, 0x0, 0x0, 0x0],
    [0xde0b6b3a7640000, 0x0, 0x0, 0x0],
    [0x8ac7230489e80000, 0x0, 0x0, 0x0],
    [0x6bc75e2d63100000, 0x5, 0x0, 0x0],
    [0x35c9adc5dea00000, 0x36, 0x0, 0x0],
    [0x19e0c9bab2400000, 0x21e, 0x0, 0x0],
    [0x2c7e14af6800000, 0x152d, 0x0, 0x0],
    [0x1bcecceda1000000, 0xd3c2, 0x0, 0x0],
    [0x161401484a000000, 0x84595, 0x0, 0x0],
    [0xdcc80cd2e4000000, 0x52b7d2, 0x0, 0x0],
    [0x9fd0803ce8000000, 0x33b2e3c, 0x0, 0x0],
    [0x3e25026110000000, 0x204fce5e, 0x0, 0x0],
    [0x6d7217caa0000000, 0x1431e0fae, 0x0, 0x0],
    [0x4674edea40000000, 0xc9f2c9cd0, 0x0, 0x0],
    [0xc0914b2680000000, 0x7e37be2022, 0x0, 0x0],
    [0x85acef8100000000, 0x4ee2d6d415b, 0x0, 0x0],
    [0x38c15b0a00000000, 0x314dc6448d93, 0x0, 0x0],
    [0x378d8e6400000000, 0x1ed09bead87c0, 0x0, 0x0],
    [0x2b878fe800000000, 0x13426172c74d82, 0x0, 0x0],
    [0xb34b9f1000000000, 0xc097ce7bc90715, 0x0, 0x0],
    [0xf436a000000000, 0x785ee10d5da46d9, 0x0, 0x0],
    [0x98a224000000000, 0x4b3b4ca85a86c47a, 0x0, 0x0],
    [0x5f65568000000000, 0xf050fe938943acc4, 0x2, 0x0],
    [0xb9f5610000000000, 0x6329f1c35ca4bfab, 0x1d, 0x0],
    [0x4395ca0000000000, 0xdfa371a19e6f7cb5, 0x125, 0x0],
    [0xa3d9e40000000000, 0xbc627050305adf14, 0xb7a, 0x0],
    [0x6682e80000000000, 0x5bd86321e38cb6ce, 0x72cb, 0x0],
    [0x11d100000000000, 0x9673df52e37f2410, 0x47bf1, 0x0],
    [0xb22a00000000000, 0xe086b93ce2f768a0, 0x2cd76f, 0x0],
    [0x6f5a400000000000, 0xc5433c60ddaa1640, 0x1c06a5e, 0x0],
    [0x5986800000000000, 0xb4a05bc8a8a4de84, 0x118427b3, 0x0],
    [0x7f41000000000000, 0xe4395d69670b12b, 0xaf298d05, 0x0],
    [0xf88a000000000000, 0x8ea3da61e066ebb2, 0x6d79f8232, 0x0],
    [0xb564000000000000, 0x926687d2c40534fd, 0x446c3b15f9, 0x0],
    [0x15e8000000000000, 0xb8014e3ba83411e9, 0x2ac3a4edbbf, 0x0],
    [0xdb10000000000000, 0x300d0e549208b31a, 0x1aba4714957d, 0x0],
    [0x8ea0000000000000, 0xe0828f4db456ff0c, 0x10b46c6cdd6e3, 0x0],
    [0x9240000000000000, 0xc51999090b65f67d, 0xa70c3c40a64e6, 0x0],
    [0xb680000000000000, 0xb2fffa5a71fba0e7, 0x6867a5a867f103, 0x0],
    [0x2100000000000000, 0xfdffc78873d4490d, 0x4140c78940f6a24, 0x0],
    [0x4a00000000000000, 0xebfdcb54864ada83, 0x28c87cb5c89a2571, 0x0],
    [0xe400000000000000, 0x37e9f14d3eec8920, 0x97d4df19d6057673, 0x1],
    [0xe800000000000000, 0x2f236d04753d5b48, 0xee50b7025c36a080, 0xf],
    [0x1000000000000000, 0xd762422c946590d9, 0x4f2726179a224501, 0x9f],
    [0xa000000000000000, 0x69d695bdcbf7a87a, 0x17877cec0556b212, 0x639],
    [0x4000000000000000, 0x2261d969f7ac94ca, 0xeb4ae1383562f4b8, 0x3e3a],
    [0x8000000000000000, 0x57d27e23acbdcfe6, 0x30eccc3215dd8f31, 0x26e4d],
    [0x0, 0x6e38ed64bf6a1f01, 0xe93ff9f4daa797ed, 0x184f03],
    [0x0, 0x4e3945ef7a25360a, 0x1c7fc3908a8bef46, 0xf31627],
    [0x0, 0xe3cbb5ac5741c64, 0x1cfda3a5697758bf, 0x97edd87],
    [0x0, 0x8e5f518bb6891be8, 0x21e864761ea97776, 0x5ef4a747],
    [0x0, 0x8fb92f75215b1710, 0x5313ec9d329eaaa1, 0x3b58e88c7],
    [0x0, 0x9d3bda934d8ee6a0, 0x3ec73e23fa32aa4f, 0x25179157c9],
    [0x0, 0x245689c107950240, 0x73c86d67c5faa71c, 0x172ebad6ddc],
    [0x0, 0x6b61618a4bd21680, 0x85d4460dbbca8719, 0xe7d34c64a9c],
    [0x0, 0x31cdcf66f634e100, 0x3a4abc8955e946fe, 0x90e40fbeea1d],
    [0x0, 0xf20a1a059e10ca00, 0x46eb5d5d5b1cc5ed, 0x5a8e89d752524],
    [0x0, 0x746504382ca7e400, 0xc531a5a58f1fbb4b, 0x3899162693736a],
    [0x0, 0x8bf22a31be8ee800, 0xb3f07877973d50f2, 0x235fadd81c2822b],
    [0x0, 0x7775a5f171951000, 0x764b4abe8652979, 0x161bcca7119915b5],
];

/// Fixed-point Decimal type with 18 decimal places
pub struct Decimal(U256);

const fn one_u256() -> U256 {
    U256([1_000_000_000_000_000_000, 0, 0, 0])
}

const fn one_u512() -> U512 {
    U512([1_000_000_000_000_000_000, 0, 0, 0, 0, 0, 0, 0])
}

fn pow10(e: u8) -> Option<U256> {
    POWERS_OF_10.get(e as usize).copied().map(U256)
}

fn rescale(n: U256, d: i8) -> Option<U256> {
    match d.cmp(&0) {
        std::cmp::Ordering::Greater => n.checked_mul(pow10(d as u8)?),
        std::cmp::Ordering::Less => n.checked_div(pow10((-d) as u8)?),
        std::cmp::Ordering::Equal => Some(n),
    }
}

impl Decimal {
    const SCALE: u8 = 18;

    pub fn from_raw_with_scale<I: Into<U256>>(raw: I, scale: u8) -> Option<Self> {
        let raw = raw.into();

        let diff = (Self::SCALE as i8) - (scale as i8);
        let raw = rescale(raw, diff)?;

        Some(Self(raw))
    }

    pub fn to_raw_with_scale(&self, scale: u8) -> Option<U256> {
        let diff = (scale as i8) - (Self::SCALE as i8);
        let raw = rescale(self.0, diff)?;

        Some(raw)
    }

    pub const fn one() -> Self {
        Self(one_u256())
    }

    pub const fn zero() -> Self {
        Self(U256([0, 0, 0, 0]))
    }

    pub fn add(&self, other: &Self) -> Option<Self> {
        self.0.checked_add(other.0).map(Self)
    }

    pub fn sub(&self, other: &Self) -> Option<Self> {
        self.0.checked_sub(other.0).map(Self)
    }

    pub fn mul(&self, other: &Self) -> Option<Self> {
        U256::try_from(self.0.full_mul(other.0) / one_u512())
            .ok()
            .map(Self)
    }

    pub fn div(&self, other: &Self) -> Option<Self> {
        U256::try_from(self.0.full_mul(one_u256()).checked_div(other.0.into())?)
            .ok()
            .map(Self)
    }

    pub fn mul_div(&self, mul: &Self, div: &Self) -> Option<Self> {
        U256::try_from(self.0.full_mul(mul.0).checked_div(div.0.into())?)
            .ok()
            .map(Self)
    }
}

impl std::fmt::Display for Decimal {
    fn fmt(&self, f: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
        let (div, module) = self.0.div_mod(pow10(18).unwrap());
        write!(f, "{div}.{module:018}")
    }
}

#[cfg(test)]
mod test {
    use primitive_types::U256;

    use super::Decimal;

    #[test]
    fn rescale() {
        let n = Decimal::from_raw_with_scale(1_000_000_000, 9).unwrap();

        let e = n.to_raw_with_scale(9).unwrap();

        assert_eq!(U256::from(1_000_000_000), e);
    }
}
