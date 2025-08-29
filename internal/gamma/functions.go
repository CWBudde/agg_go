package gamma

import "math"

// GammaNone ≈ agg::gamma_none
type GammaNone struct{}

func (GammaNone) Apply(x float64) float64 { return x }

// GammaPower ≈ agg::gamma_power
type GammaPower struct{ g float64 }

func NewGammaPower(g float64) GammaPower  { return GammaPower{g: g} }
func (gp *GammaPower) SetGamma(g float64) { gp.g = g }
func (gp GammaPower) Gamma() float64      { return gp.g }
func (gp GammaPower) Apply(x float64) float64 {
	return math.Pow(x, gp.g)
}

// GammaThreshold ≈ agg::gamma_threshold
type GammaThreshold struct{ t float64 }

func NewGammaThreshold(t float64) GammaThreshold  { return GammaThreshold{t: t} }
func (gt *GammaThreshold) SetThreshold(t float64) { gt.t = t }
func (gt GammaThreshold) Threshold() float64      { return gt.t }
func (gt GammaThreshold) Apply(x float64) float64 {
	if x < gt.t {
		return 0.0
	}
	return 1.0
}

// GammaLinear ≈ agg::gamma_linear
type GammaLinear struct{ s, e float64 }

func NewGammaLinear(s, e float64) GammaLinear { return GammaLinear{s: s, e: e} }
func (gl *GammaLinear) Set(s, e float64)      { gl.s, gl.e = s, e }
func (gl *GammaLinear) SetStart(s float64)    { gl.s = s }
func (gl *GammaLinear) SetEnd(e float64)      { gl.e = e }
func (gl GammaLinear) Start() float64         { return gl.s }
func (gl GammaLinear) End() float64           { return gl.e }
func (gl GammaLinear) Apply(x float64) float64 {
	if x < gl.s {
		return 0.0
	}
	if x > gl.e {
		return 1.0
	}
	return (x - gl.s) / (gl.e - gl.s)
}

// GammaMultiply ≈ agg::gamma_multiply
type GammaMultiply struct{ mul float64 }

func NewGammaMultiply(v float64) GammaMultiply { return GammaMultiply{mul: v} }
func (gm *GammaMultiply) SetValue(v float64)   { gm.mul = v }
func (gm GammaMultiply) Value() float64        { return gm.mul }
func (gm GammaMultiply) Apply(x float64) float64 {
	y := x * gm.mul
	if y > 1.0 {
		y = 1.0
	}
	return y
}

// sRGB/linear helpers (AGG inline functions).
func SRGBToLinear(x float64) float64 {
	if x <= 0.04045 {
		return x / 12.92
	}
	return math.Pow((x+0.055)/1.055, 2.4)
}

func LinearToSRGB(x float64) float64 {
	if x <= 0.0031308 {
		return x * 12.92
	}
	return 1.055*math.Pow(x, 1.0/2.4) - 0.055
}
