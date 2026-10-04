// Agg2D C++ oracle: renders small scenes through the original AGG 2.6
// agg2d.cpp and dumps the raw pixel buffers (RGBA order, top row first) plus
// the image-filter LUTs selected by Agg2D::imageFilter(). The Go test
// cpp_oracle_test.go replays the same scenes and compares byte-for-byte.
//
// Agg2D::Color is srgba8 and is converted to the linear rgba8 ColorType on
// use; the Go port treats Color as linear. Every colour channel below is
// therefore 0 or 255 (fixed points of the sRGB->linear table; alpha is not
// converted) so the scenes isolate rendering semantics from that delta.
//
// Build (macOS/Homebrew FreeType; FreeType is only needed because agg2d.h
// otherwise pulls in the Win32 font engine -- no text is rendered). Recent
// FreeType declares FT_Outline::tags as unsigned char*, so agg_font_freetype.cpp
// needs a one-cast patch to compile:
//
//   AGG=/path/to/agg-2.6/agg-src
//   sed 's/tags  = outline.tags  + first;/tags  = (char*)outline.tags  + first;/' \
//     $AGG/font_freetype/agg_font_freetype.cpp > ft_patched.cpp
//   clang++ -std=c++11 -O0 -ffp-contract=off -DAGG2D_USE_FREETYPE \
//     -I$AGG/include -I$AGG/agg2d -I$AGG/font_freetype \
//     $(pkg-config --cflags freetype2) \
//     agg2d_oracle.cpp $AGG/agg2d/agg2d.cpp ft_patched.cpp \
//     $AGG/src/agg_trans_affine.cpp $AGG/src/agg_bezier_arc.cpp $AGG/src/agg_arc.cpp \
//     $AGG/src/agg_rounded_rect.cpp $AGG/src/agg_curves.cpp $AGG/src/agg_vcgen_stroke.cpp \
//     $AGG/src/agg_image_filters.cpp $AGG/src/agg_sqrt_tables.cpp \
//     $(pkg-config --libs freetype2) -o agg2d_oracle
//   ./agg2d_oracle <outdir>
#include <cstdio>
#include <cstring>
#include <string>
#include <vector>

#include "agg2d.h"

static const int W = 40;
static const int H = 40;

typedef void (*SceneFn)(Agg2D&);

static void triangle(Agg2D& g, Agg2D::DrawPathFlag flag)
{
    g.resetPath();
    g.moveTo(4.3, 5.7);
    g.lineTo(35.6, 12.2);
    g.lineTo(13.1, 36.4);
    g.closePolygon();
    g.drawPath(flag);
}

static void curveShape(Agg2D& g, double s)
{
    // Coordinates are in "unit" space multiplied by s.
    g.resetPath();
    g.moveTo(0.5 * s, 5.5 * s);
    g.cubicCurveTo(0.5 * s, 0.2 * s, 6.2 * s, 0.4 * s, 5.8 * s, 3.1 * s);
    g.cubicCurveTo(5.4 * s, 6.0 * s, 2.0 * s, 4.0 * s, 0.5 * s, 5.5 * s);
    g.closePolygon();
    g.drawPath(Agg2D::FillAndStroke);
}

// --- master alpha -----------------------------------------------------------

static void sceneMasterAlphaFill(Agg2D& g)
{
    g.masterAlpha(0.5);
    g.noLine();
    g.fillColor(255, 0, 0);
    triangle(g, Agg2D::FillOnly);
    g.fillColor(0, 0, 255, 128);
    g.resetPath();
    g.moveTo(2.5, 22.25);
    g.lineTo(37.75, 22.25);
    g.lineTo(37.75, 30.5);
    g.lineTo(2.5, 30.5);
    g.closePolygon();
    g.drawPath(Agg2D::FillOnly);
}

static void sceneMasterAlphaStroke(Agg2D& g)
{
    g.masterAlpha(0.5);
    g.noFill();
    g.lineColor(0, 0, 255);
    g.lineWidth(3.0);
    triangle(g, Agg2D::StrokeOnly);
}

static void sceneMasterAlphaOver1(Agg2D& g)
{
    g.masterAlpha(1.5);
    g.lineColor(0, 0, 0, 200);
    g.fillColor(0, 255, 0, 160);
    g.lineWidth(1.5);
    triangle(g, Agg2D::FillAndStroke);
}

static void sceneMasterAlphaFillWithLineColor(Agg2D& g)
{
    g.masterAlpha(0.3);
    g.lineColor(255, 0, 255);
    triangle(g, Agg2D::FillWithLineColor);
}

// --- anti-alias gamma ---------------------------------------------------------

static void sceneGamma22(Agg2D& g)
{
    g.antiAliasGamma(2.2);
    g.fillColor(255, 0, 0);
    g.lineColor(0, 0, 0);
    g.lineWidth(1.25);
    triangle(g, Agg2D::FillAndStroke);
}

static void sceneGammaAlpha(Agg2D& g)
{
    g.antiAliasGamma(0.6);
    g.masterAlpha(0.7);
    g.fillColor(0, 0, 255);
    g.lineColor(0, 255, 255);
    g.lineWidth(2.0);
    triangle(g, Agg2D::FillAndStroke);
}

static void sceneGammaLarge(Agg2D& g)
{
    // Outside the old Go clamp range [0.1, 3].
    g.antiAliasGamma(4.5);
    g.fillColor(0, 0, 0);
    triangle(g, Agg2D::FillOnly);
}

static void sceneGradientMasterAlpha(Agg2D& g)
{
    g.masterAlpha(0.5);
    g.noLine();
    g.fillLinearGradient(3, 3, 37, 30, Agg2D::Color(0, 0, 255, 255), Agg2D::Color(0, 0, 255, 40), 1.0);
    triangle(g, Agg2D::FillOnly);
}

// --- approximation scale ------------------------------------------------------

static void sceneApproxDefault(Agg2D& g)
{
    // No scale call: conv_curve/conv_stroke keep their default scale of 1.0.
    g.fillColor(0, 255, 255);
    g.lineColor(0, 0, 0);
    g.lineWidth(3.0);
    curveShape(g, 6.0);
}

static void sceneApproxScaled(Agg2D& g)
{
    // scale() sets approximation_scale(worldToScreen(1.0) * g_approxScale).
    g.scale(6.0, 6.0);
    g.fillColor(0, 255, 255);
    g.lineColor(0, 0, 0);
    g.lineWidth(0.5);
    curveShape(g, 1.0);
}

static void sceneApproxAfterReset(Agg2D& g)
{
    // resetTransformations() does not touch the approximation scale.
    g.scale(0.25, 0.25);
    g.resetTransformations();
    g.fillColor(0, 255, 255);
    g.lineColor(0, 0, 0);
    g.lineWidth(3.0);
    curveShape(g, 6.0);
}

static void sceneApproxSkew(Agg2D& g)
{
    // skew()/rotate()/translate() do not update the approximation scale.
    g.scale(0.5, 0.5);
    g.skew(0.4, 0.1);
    g.rotate(0.2);
    g.translate(4.0, 2.0);
    g.fillColor(0, 255, 255);
    g.lineColor(0, 0, 0);
    g.lineWidth(4.0);
    curveShape(g, 10.0);
}

// --- drawPath guards ----------------------------------------------------------

static void sceneGuardNoFillClear(Agg2D& g)
{
    // A fully transparent fill colour skips the fill entirely, so even the
    // clear operator leaves the destination untouched.
    g.blendMode(Agg2D::BlendClear);
    g.noFill();
    g.noLine();
    triangle(g, Agg2D::FillAndStroke);
    triangle(g, Agg2D::FillOnly);
    triangle(g, Agg2D::StrokeOnly);
    triangle(g, Agg2D::FillWithLineColor);
}

static void sceneGuardLineWidth(Agg2D& g)
{
    g.fillColor(255, 0, 0);
    g.lineColor(0, 0, 0);
    g.lineWidth(-3.0);
    triangle(g, Agg2D::FillAndStroke);
    g.lineWidth(0.0);
    g.resetPath();
    g.moveTo(3, 36);
    g.lineTo(37, 30);
    g.drawPath(Agg2D::StrokeOnly);
}

struct Scene
{
    const char* name;
    SceneFn fn;
};

static const Scene scenes[] = {
    {"master_alpha_fill", sceneMasterAlphaFill},
    {"master_alpha_stroke", sceneMasterAlphaStroke},
    {"master_alpha_over1", sceneMasterAlphaOver1},
    {"master_alpha_fill_with_line_color", sceneMasterAlphaFillWithLineColor},
    {"gamma_2_2", sceneGamma22},
    {"gamma_alpha", sceneGammaAlpha},
    {"gamma_large", sceneGammaLarge},
    {"gradient_master_alpha", sceneGradientMasterAlpha},
    {"approx_default", sceneApproxDefault},
    {"approx_scaled", sceneApproxScaled},
    {"approx_after_reset", sceneApproxAfterReset},
    {"approx_skew", sceneApproxSkew},
    {"guard_nofill_clear", sceneGuardNoFillClear},
    {"guard_line_width", sceneGuardLineWidth},
};

static unsigned long long fnv1a(const unsigned char* p, size_t n)
{
    unsigned long long h = 14695981039346656037ULL;
    for (size_t i = 0; i < n; ++i)
    {
        h ^= p[i];
        h *= 1099511628211ULL;
    }
    return h;
}

template<class Filter>
static void dumpFilter(FILE* f, const char* name, const Filter& filter)
{
    agg::image_filter_lut lut(filter, true);
    unsigned n = lut.diameter() * agg::image_subpixel_scale;
    std::vector<unsigned char> bytes(n * 2);
    for (unsigned i = 0; i < n; ++i)
    {
        agg::int16 w = lut.weight_array()[i];
        bytes[i * 2] = (unsigned char)(w & 0xFF);
        bytes[i * 2 + 1] = (unsigned char)((w >> 8) & 0xFF);
    }
    fprintf(f, "%s %.17g %u %d %016llx\n", name, lut.radius(), lut.diameter(), lut.start(),
            fnv1a(bytes.data(), bytes.size()));
}

int main(int argc, char** argv)
{
    if (argc < 2)
    {
        fprintf(stderr, "usage: %s <outdir>\n", argv[0]);
        return 1;
    }
    std::string out = argv[1];

    for (size_t s = 0; s < sizeof(scenes) / sizeof(scenes[0]); ++s)
    {
        std::vector<unsigned char> buf(W * H * 4);
        Agg2D g;
        g.attach(buf.data(), W, H, W * 4);
        g.clearAll(255, 255, 0);
        scenes[s].fn(g);

        // Agg2D renders BGRA (order_bgra); store as RGBA.
        std::vector<unsigned char> rgba(buf.size());
        for (int i = 0; i < W * H; ++i)
        {
            rgba[i * 4 + 0] = buf[i * 4 + 2];
            rgba[i * 4 + 1] = buf[i * 4 + 1];
            rgba[i * 4 + 2] = buf[i * 4 + 0];
            rgba[i * 4 + 3] = buf[i * 4 + 3];
        }
        std::string path = out + "/" + scenes[s].name + ".rgba";
        FILE* f = fopen(path.c_str(), "wb");
        if (!f)
        {
            perror(path.c_str());
            return 1;
        }
        fwrite(rgba.data(), 1, rgba.size(), f);
        fclose(f);
    }

    // The filter set and mapping of Agg2D::imageFilter() (agg2d.cpp:1241).
    std::string path = out + "/image_filters.txt";
    FILE* f = fopen(path.c_str(), "w");
    if (!f)
    {
        perror(path.c_str());
        return 1;
    }
    dumpFilter(f, "Bilinear", agg::image_filter_bilinear());
    dumpFilter(f, "Hanning", agg::image_filter_hanning());
    dumpFilter(f, "Hermite", agg::image_filter_hermite());
    dumpFilter(f, "Quadric", agg::image_filter_quadric());
    dumpFilter(f, "Bicubic", agg::image_filter_bicubic());
    dumpFilter(f, "Catrom", agg::image_filter_catrom());
    dumpFilter(f, "Spline16", agg::image_filter_spline16());
    dumpFilter(f, "Spline36", agg::image_filter_spline36());
    dumpFilter(f, "Blackman144", agg::image_filter_blackman144());
    fclose(f);
    return 0;
}
