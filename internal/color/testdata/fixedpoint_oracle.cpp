// Oracle for the fixed-point colour helpers of AGG 2.6 (rgba8/gray8/rgba16/gray16).
//
// It evaluates the original C++ helpers over exhaustive (8-bit) or
// deterministic sampled (16-bit) inputs, folds every result into an FNV-1a 64
// hash and prints a Go source file with the hashes and a few hundred literal
// spot values. The Go tests in ../fixedpoint_parity_test.go recompute the same
// hashes with the Go helpers.
//
// Regenerate (from the repository root):
//
//   clang++ -O0 -w -I../agg-2.6/agg-src/include \
//       internal/color/testdata/fixedpoint_oracle.cpp -o /tmp/fp_oracle
//   /tmp/fp_oracle > internal/color/fixedpoint_golden_data_test.go
//   gofmt -w internal/color/fixedpoint_golden_data_test.go
//
// Note on rgba16::demultiply / gray16::demultiply (static): the expression
// `a * base_mask` is evaluated in `int` and overflows for a > 32768 (signed
// overflow, undefined behaviour). The result then depends on the optimisation
// level. The oracle therefore hashes the C++ call only where it is well
// defined, and hashes the unsigned (overflow-free) reading of the same formula
// separately ("Wide"). See docs/AGG_DELTAS.md.

#include <climits>
#include <cstdio>
#include <vector>

#include "agg_color_gray.h"
#include "agg_color_rgba.h"

using namespace agg;

typedef unsigned long long u64;

struct fnv {
    u64 h;
    fnv() : h(14695981039346656037ULL) {}
    void b(unsigned v) { h ^= (v & 0xFF); h *= 1099511628211ULL; }
    void u8(unsigned v) { b(v); }
    void u16(unsigned v) { b(v); b(v >> 8); }
};

struct lcg {
    u64 s;
    explicit lcg(u64 seed) : s(seed) {}
    unsigned next16() {
        s = s * 6364136223846793005ULL + 1442695040888963407ULL;
        return unsigned(s >> 48);
    }
};

static const unsigned E16[] = {0, 1, 2, 127, 128, 255, 256, 257, 32767,
                               32768, 32769, 65279, 65280, 65534, 65535};
static const int nE16 = sizeof(E16) / sizeof(E16[0]);

static const unsigned SPOT16[] = {0, 1, 255, 256, 32767, 32768, 65534, 65535};
static const int nSPOT16 = sizeof(SPOT16) / sizeof(SPOT16[0]);

static const int kLerpSamples16 = 1 << 20;
static const int kFromDoubleSteps = 1 << 20;
static const int kGradientSteps = 4096;

static std::vector<unsigned> sample16() {
    std::vector<unsigned> s(E16, E16 + nE16);
    lcg r(1);
    while (s.size() < 64) s.push_back(r.next16());
    return s;
}

// Triples used for the 16-bit lerp/prelerp hashes: E16^3 then LCG triples.
template <class F> static void triples16(F f) {
    for (int i = 0; i < nE16; ++i)
        for (int j = 0; j < nE16; ++j)
            for (int k = 0; k < nE16; ++k) f(E16[i], E16[j], E16[k]);
    lcg r(2);
    for (int n = 0; n < kLerpSamples16; ++n) {
        unsigned p = r.next16(), q = r.next16(), a = r.next16();
        f(p, q, a);
    }
}

struct gpair8 { unsigned v1, a1, v2, a2; };
static const gpair8 GP8[] = {{0, 0, 255, 255}, {255, 255, 0, 0}, {10, 20, 200, 250},
                             {200, 250, 10, 20}, {128, 64, 127, 192}, {0, 255, 255, 0},
                             {77, 77, 77, 77}, {1, 254, 254, 1}};
struct gpair16 { unsigned v1, a1, v2, a2; };
static const gpair16 GP16[] = {{0, 0, 65535, 65535}, {65535, 65535, 0, 0},
                               {1000, 2000, 60000, 65000}, {60000, 65000, 1000, 2000},
                               {32768, 16384, 32767, 49152}, {0, 65535, 65535, 0},
                               {12345, 12345, 12345, 12345}, {1, 65534, 65534, 1}};

int main() {
    std::vector<unsigned> s16 = sample16();

    // ---------------------------------------------------------------- 8-bit
    fnv m8, gm8, d8, gd8, sc8, gsc8, l8, gl8, pl8, gpl8, md8, gmd8, fd8, gfd8, gg8, lum8;
    for (unsigned a = 0; a < 256; ++a)
        for (unsigned b = 0; b < 256; ++b) {
            m8.u8(rgba8::multiply(a, b));
            gm8.u8(gray8::multiply(a, b));
            d8.u8(rgba8::demultiply(a, b));
            gd8.u8(gray8::demultiply(a, b));
            sc8.u8(rgba8::scale_cover(a, b));
            gsc8.u8(gray8::scale_cover(a, b));
            rgba8 c(a, a, a, b);
            c.demultiply();
            md8.u8(c.r);
            gray8 g(a, b);
            g.demultiply();
            gmd8.u8(g.v);
            for (unsigned x = 0; x < 256; ++x) {
                l8.u8(rgba8::lerp(a, b, x));
                gl8.u8(gray8::lerp(a, b, x));
                pl8.u8(rgba8::prelerp(a, b, x));
                gpl8.u8(gray8::prelerp(a, b, x));
                lum8.u8(gray8::luminance(rgba8(a, b, x)));
            }
        }
    for (int i = 0; i <= kFromDoubleSteps; ++i) {
        double d = double(i) / double(kFromDoubleSteps);
        fd8.u8(rgba8::from_double(d));
        gfd8.u8(gray8::from_double(d));
    }
    for (int p = 0; p < int(sizeof(GP8) / sizeof(GP8[0])); ++p) {
        gray8 c1(GP8[p].v1, GP8[p].a1), c2(GP8[p].v2, GP8[p].a2);
        for (int i = 0; i <= kGradientSteps; ++i) {
            gray8 r = c1.gradient(c2, double(i) / double(kGradientSteps));
            gg8.u8(r.v);
            gg8.u8(r.a);
        }
    }

    // --------------------------------------------------------------- 16-bit
    fnv m16, gm16, d16, gd16, dw16, mc16, gmc16, sc16, gsc16, l16, gl16, pl16, gpl16,
        md16, gmd16, fd16, gfd16, gg16, lum16;
    for (unsigned a = 0; a < 65536; ++a) {
        for (size_t j = 0; j < s16.size(); ++j) {
            unsigned b = s16[j];
            m16.u16(rgba16::multiply(a, b));
            gm16.u16(gray16::multiply(a, b));
            // Well-defined region of the static demultiply only.
            if (u64(a) * 65535ULL + (b >> 1) <= u64(INT_MAX)) {
                d16.u16(rgba16::demultiply(a, b));
                gd16.u16(gray16::demultiply(a, b));
            }
            // Overflow-free (unsigned) reading of the same formula.
            unsigned w;
            if (a * b == 0) w = 0;
            else if (a >= b) w = 65535;
            else w = (a * 65535u + (b >> 1)) / b;
            dw16.u16(w & 0xFFFF);
            rgba16 c(b, b, b, a);
            c.demultiply();
            md16.u16(c.r);
            gray16 g(b, a);
            g.demultiply();
            gmd16.u16(g.v);
        }
        for (unsigned cv = 0; cv < 256; ++cv) {
            mc16.u16(rgba16::mult_cover(a, cv));
            gmc16.u16(gray16::mult_cover(a, cv));
            sc16.u8(rgba16::scale_cover(cv, a));
            gsc16.u8(gray16::scale_cover(cv, a));
        }
    }
    triples16([&](unsigned p, unsigned q, unsigned a) {
        l16.u16(rgba16::lerp(p, q, a));
        gl16.u16(gray16::lerp(p, q, a));
        pl16.u16(rgba16::prelerp(p, q, a));
        gpl16.u16(gray16::prelerp(p, q, a));
    });
    for (size_t i = 0; i < s16.size(); ++i)
        for (size_t j = 0; j < s16.size(); ++j)
            for (size_t k = 0; k < s16.size(); ++k)
                lum16.u16(gray16::luminance(rgba16(s16[i], s16[j], s16[k])));
    for (int i = 0; i <= kFromDoubleSteps; ++i) {
        double d = double(i) / double(kFromDoubleSteps);
        fd16.u16(rgba16::from_double(d));
        gfd16.u16(gray16::from_double(d));
    }
    for (int p = 0; p < int(sizeof(GP16) / sizeof(GP16[0])); ++p) {
        gray16 c1(GP16[p].v1, GP16[p].a1), c2(GP16[p].v2, GP16[p].a2);
        for (int i = 0; i <= kGradientSteps; ++i) {
            gray16 r = c1.gradient(c2, double(i) / double(kGradientSteps));
            gg16.u16(r.v);
            gg16.u16(r.a);
        }
    }

    // ------------------------------------------------------------- add()
    fnv add8, gadd8, add16, gadd16;
    for (unsigned x = 0; x < 256; x += 5)
        for (unsigned y = 0; y < 256; y += 5)
            for (int m = 0; m < 2; ++m) {
                unsigned ca = m ? 255 : y;
                for (unsigned cv = 0; cv < 256; ++cv) {
                    rgba8 d(x, x, x, x);
                    d.add(rgba8(y, y, y, ca), cv);
                    add8.u8(d.r);
                    add8.u8(d.a);
                    gray8 g(x, x);
                    g.add(gray8(y, ca), cv);
                    gadd8.u8(g.v);
                    gadd8.u8(g.a);
                }
            }
    for (size_t i = 0; i < s16.size(); ++i)
        for (size_t j = 0; j < s16.size(); ++j)
            for (int m = 0; m < 2; ++m) {
                unsigned x = s16[i], y = s16[j];
                unsigned ca = m ? 65535 : y;
                for (unsigned cv = 0; cv < 256; ++cv) {
                    rgba16 d(x, x, x, x);
                    d.add(rgba16(y, y, y, ca), cv);
                    add16.u16(d.r);
                    add16.u16(d.a);
                    gray16 g(x, x);
                    g.add(gray16(y, ca), cv);
                    gadd16.u16(g.v);
                    gadd16.u16(g.a);
                }
            }

    // ---------------------------------------------------------------- output
    printf("// Code generated by testdata/fixedpoint_oracle.cpp from AGG 2.6 C++. DO NOT EDIT.\n\n");
    printf("package color\n\n");
    printf("// FNV-1a 64 hashes of the C++ helper outputs (see fixedpoint_parity_test.go).\n");
    printf("const (\n");
#define H(name, f) printf("\t%s = 0x%016llx\n", name, f.h)
    H("goldenHashRGBA8Multiply", m8);
    H("goldenHashGray8Multiply", gm8);
    H("goldenHashRGBA8Demultiply", d8);
    H("goldenHashGray8Demultiply", gd8);
    H("goldenHashRGBA8ScaleCover", sc8);
    H("goldenHashGray8ScaleCover", gsc8);
    H("goldenHashRGBA8Lerp", l8);
    H("goldenHashGray8Lerp", gl8);
    H("goldenHashRGBA8Prelerp", pl8);
    H("goldenHashGray8Prelerp", gpl8);
    H("goldenHashRGBA8MemberDemultiply", md8);
    H("goldenHashGray8MemberDemultiply", gmd8);
    H("goldenHashRGBA8FromDouble", fd8);
    H("goldenHashGray8FromDouble", gfd8);
    H("goldenHashGray8Gradient", gg8);
    H("goldenHashGray8LuminanceRGBA8", lum8);
    H("goldenHashRGBA16Multiply", m16);
    H("goldenHashGray16Multiply", gm16);
    H("goldenHashRGBA16DemultiplyDefined", d16);
    H("goldenHashGray16DemultiplyDefined", gd16);
    H("goldenHashRGBA16DemultiplyWide", dw16);
    H("goldenHashRGBA16MultCover", mc16);
    H("goldenHashGray16MultCover", gmc16);
    H("goldenHashRGBA16ScaleCover", sc16);
    H("goldenHashGray16ScaleCover", gsc16);
    H("goldenHashRGBA16Lerp", l16);
    H("goldenHashGray16Lerp", gl16);
    H("goldenHashRGBA16Prelerp", pl16);
    H("goldenHashGray16Prelerp", gpl16);
    H("goldenHashRGBA16MemberDemultiply", md16);
    H("goldenHashGray16MemberDemultiply", gmd16);
    H("goldenHashRGBA16FromDouble", fd16);
    H("goldenHashGray16FromDouble", gfd16);
    H("goldenHashGray16Gradient", gg16);
    H("goldenHashGray16LuminanceRGBA16", lum16);
    H("goldenHashRGBA8Add", add8);
    H("goldenHashGray8Add", gadd8);
    H("goldenHashRGBA16Add", add16);
    H("goldenHashGray16Add", gadd16);
#undef H
    printf(")\n\n");

    // Spot values: 16-bit lerp/prelerp over SPOT16^3 and the first 128 LCG triples.
    printf("// goldenRGBA16LerpSpots holds {p, q, a, lerp(p,q,a), prelerp(p,q,a)}.\n");
    printf("var goldenRGBA16LerpSpots = [][5]uint16{\n");
    for (int i = 0; i < nSPOT16; ++i)
        for (int j = 0; j < nSPOT16; ++j)
            for (int k = 0; k < nSPOT16; ++k) {
                unsigned p = SPOT16[i], q = SPOT16[j], a = SPOT16[k];
                printf("\t{%u, %u, %u, %u, %u},\n", p, q, a, rgba16::lerp(p, q, a),
                       rgba16::prelerp(p, q, a));
            }
    {
        lcg r(3);
        for (int n = 0; n < 128; ++n) {
            unsigned p = r.next16(), q = r.next16(), a = r.next16();
            printf("\t{%u, %u, %u, %u, %u},\n", p, q, a, rgba16::lerp(p, q, a),
                   rgba16::prelerp(p, q, a));
        }
    }
    printf("}\n\n");

    // 16-bit multiply / mult_cover / scale_cover / demultiply spot values.
    printf("// goldenRGBA16MulSpots holds {a, b, multiply(a,b), demultiply(a,b)}\n");
    printf("// (demultiply in its overflow-free reading).\n");
    printf("var goldenRGBA16MulSpots = [][4]uint16{\n");
    for (int i = 0; i < nE16; ++i)
        for (int j = 0; j < nE16; ++j) {
            unsigned a = E16[i], b = E16[j];
            unsigned w;
            if (a * b == 0) w = 0;
            else if (a >= b) w = 65535;
            else w = (a * 65535u + (b >> 1)) / b;
            printf("\t{%u, %u, %u, %u},\n", a, b, rgba16::multiply(a, b), w);
        }
    printf("}\n\n");

    printf("// goldenRGBA16CoverSpots holds {v, cover, mult_cover(v,cover), scale_cover(cover,v)}.\n");
    printf("var goldenRGBA16CoverSpots = [][4]uint16{\n");
    {
        static const unsigned C8[] = {0, 1, 2, 127, 128, 254, 255};
        for (int i = 0; i < nE16; ++i)
            for (int j = 0; j < 7; ++j) {
                unsigned v = E16[i], cv = C8[j];
                printf("\t{%u, %u, %u, %u},\n", v, cv, rgba16::mult_cover(v, cv),
                       unsigned(rgba16::scale_cover(cv, v)));
            }
    }
    printf("}\n\n");

    // 8-bit lerp/prelerp spot values.
    printf("// goldenRGBA8LerpSpots holds {p, q, a, lerp(p,q,a), prelerp(p,q,a)}.\n");
    printf("var goldenRGBA8LerpSpots = [][5]uint8{\n");
    {
        static const unsigned S8[] = {0, 1, 127, 128, 254, 255};
        for (int i = 0; i < 6; ++i)
            for (int j = 0; j < 6; ++j)
                for (int k = 0; k < 6; ++k) {
                    unsigned p = S8[i], q = S8[j], a = S8[k];
                    printf("\t{%u, %u, %u, %u, %u},\n", p, q, a, rgba8::lerp(p, q, a),
                           rgba8::prelerp(p, q, a));
                }
    }
    printf("}\n\n");

    // Gradient spot values for gray8/gray16, including k near 1 where the
    // C++ code truncates uround(k * base_scale) to value_type.
    printf("// goldenGrayGradientSpots holds {k*1e6, v8, a8, v16, a16} for the pair\n");
    printf("// gray8(10,20)->gray8(200,250) and gray16(1000,2000)->gray16(60000,65000).\n");
    printf("var goldenGrayGradientSpots = [][5]uint32{\n");
    {
        static const double K[] = {0.0, 0.25, 0.5, 0.75, 0.996, 0.998, 0.999, 1.0,
                                   0.99999, 0.999992, 0.999993};
        for (int i = 0; i < int(sizeof(K) / sizeof(K[0])); ++i) {
            gray8 a8(10, 20), b8(200, 250);
            gray16 a16(1000, 2000), b16(60000, 65000);
            gray8 r8 = a8.gradient(b8, K[i]);
            gray16 r16 = a16.gradient(b16, K[i]);
            printf("\t{%u, %u, %u, %u, %u},\n", unsigned(K[i] * 1e6 + 0.5), r8.v, r8.a,
                   r16.v, r16.a);
        }
    }
    printf("}\n");
    return 0;
}
