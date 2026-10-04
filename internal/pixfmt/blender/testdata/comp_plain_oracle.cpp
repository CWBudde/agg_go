// Composite-on-straight-alpha C++ oracle for CompositeBlenderPlain.
//
// AGG 2.6 only composites into premultiplied buffers. The Go port's Agg2D
// framebuffer stores straight alpha, so CompositeBlenderPlain premultiplies
// the destination on read, runs the AGG operator, and demultiplies on write.
// This program builds that bridge from stock AGG pieces -- the same three
// calls as comp_op_adaptor_rgba_plain in engine/cpp_native_stub.cpp:
// multiplier_rgba::premultiply, comp_op_adaptor_rgba::blend_pix (source
// premultiplied with rgba8::multiply), multiplier_rgba::demultiply -- and
// dumps the result for a grid of straight dst pixels x sources x covers.
// comp_plain_oracle_test.go replays the grid through BlendPix.
//
// Build and run (header-only):
//
//   AGG=/path/to/agg-2.6/agg-src
//   clang++ -std=c++11 -O0 -ffp-contract=off -I$AGG/include \
//     comp_plain_oracle.cpp -o comp_plain_oracle
//   ./comp_plain_oracle <outdir>
//
// Output: <outdir>/comp_plain_<op>.bin, 4 bytes (RGBA) per grid entry in the
// loop order of main().
#include <cstdio>
#include <string>
#include <vector>

#include "agg_pixfmt_rgba.h"

typedef agg::rgba8 color_type;
typedef agg::order_rgba order_type;
typedef agg::multiplier_rgba<color_type, order_type> multiplier;
typedef agg::comp_op_adaptor_rgba<color_type, order_type> adaptor;

static const unsigned kColor[] = {0, 1, 77, 128, 200, 255};
static const unsigned kDstAlpha[] = {0, 1, 64, 128, 200, 254, 255};
static const unsigned kSrcAlpha[] = {0, 1, 77, 128, 179, 255};
static const unsigned kCover[] = {1, 64, 128, 255};

#define COUNT(a) (sizeof(a) / sizeof((a)[0]))

struct Op
{
    const char* name;
    unsigned op;
};

static const Op kOps[] = {
    {"src_over", agg::comp_op_src_over},
    {"xor", agg::comp_op_xor},
};

int main(int argc, char** argv)
{
    if (argc < 2)
    {
        fprintf(stderr, "usage: %s <outdir>\n", argv[0]);
        return 1;
    }
    const unsigned nc = COUNT(kColor);
    for (unsigned o = 0; o < COUNT(kOps); ++o)
    {
        std::vector<unsigned char> out;
        for (unsigned da = 0; da < COUNT(kDstAlpha); ++da)
        for (unsigned sa = 0; sa < COUNT(kSrcAlpha); ++sa)
        for (unsigned cv = 0; cv < COUNT(kCover); ++cv)
        for (unsigned i = 0; i < nc; ++i)
        for (unsigned j = 0; j < nc; ++j)
        {
            color_type::value_type p[4];
            p[order_type::R] = kColor[i];
            p[order_type::G] = kColor[nc - 1 - i];
            p[order_type::B] = kColor[(i + 2) % nc];
            p[order_type::A] = kDstAlpha[da];
            unsigned r = kColor[j];
            unsigned g = kColor[(j + 3) % nc];
            unsigned b = kColor[nc - 1 - j];
            unsigned a = kSrcAlpha[sa];

            multiplier::premultiply(p);
            adaptor::blend_pix(kOps[o].op, p, r, g, b, a, kCover[cv]);
            multiplier::demultiply(p);

            out.push_back(p[order_type::R]);
            out.push_back(p[order_type::G]);
            out.push_back(p[order_type::B]);
            out.push_back(p[order_type::A]);
        }
        std::string path = std::string(argv[1]) + "/comp_plain_" + kOps[o].name + ".bin";
        FILE* f = fopen(path.c_str(), "wb");
        if (!f)
        {
            perror(path.c_str());
            return 1;
        }
        fwrite(out.data(), 1, out.size(), f);
        fclose(f);
    }
    return 0;
}
