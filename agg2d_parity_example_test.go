package agg

// ExampleAgg2D_cppParity shows the public Go equivalents for the main Agg2D
// call-shape differences that remain after the parity shims.
func ExampleAgg2D_cppParity() {
	a := NewAgg2D()
	buf := make([]uint8, 64*64*4)
	a.Attach(buf, 64, 64, 64*4)

	// C++: agg.clipBox(4, 5, 40, 50);
	a.ClipBox(4, 5, 40, 50)

	// C++: Agg2D::RectD r = agg.clipBox();
	r := a.GetClipBoxRect()
	_, _, _, _ = r.X1, r.Y1, r.X2, r.Y2

	// C++: agg.drawPath();                       // default FillAndStroke
	a.DrawPathDefault()

	// C++: agg.drawPathNoTransform();            // default FillAndStroke
	a.DrawPathNoTransformDefault()

	// C++: agg.viewport(...);                    // default XMidYMid
	a.ViewportDefault(0, 0, 10, 10, 0, 0, 100, 100)

	// C++: agg.font("font.ttf", 14.0);           // default style/cache/angle
	_ = a.FontDefault("font.ttf", 14.0)

	// C++: agg.text(10, 20, "hello");            // default roundOff/dx/dy
	a.TextDefault(10, 20, "hello")

	// C++: agg.blendImage(img, 5, 6);            // default alpha = 255
	img := Image{}
	img.Attach([]uint8{255, 0, 0, 255}, 1, 1, 4)
	_ = a.BlendImageSimpleDefaultAlpha(&img, 5, 6)

	// C++: img.premultiply(); img.demultiply();
	_ = img.Premultiply()
	_ = img.Demultiply()

	// C++: agg.parallelogram(x1, y1, x2, y2, para);
	// Go keeps the existing unit-square helper and exposes the rectangle-based
	// overload explicitly:
	a.ParallelogramFromRect(0, 0, 10, 20, []float64{0, 0, 10, 0, 0, 20})
}
