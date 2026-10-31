package datasynth

import (
	"crypto/rand"
	"encoding/binary"
	"math"
)

// RandomFloat returns a pseudo-random float64 in [0.0, 1.0) using cryptographically secure random bytes.
func RandomFloat() float64 {
	var b [8]byte
	_, _ = rand.Read(b[:])
	val := binary.LittleEndian.Uint64(b[:])
	return float64(val&(1<<53-1)) / float64(1<<53)
}

// SampleLaplace draws a random sample from a zero-mean Laplace distribution with scale parameter b.
// Laplace(0, b) where b = Sensitivity / Epsilon.
func SampleLaplace(scale float64) float64 {
	if scale <= 0 {
		return 0.0
	}
	u := RandomFloat() - 0.5
	if u == 0 {
		return 0.0
	}
	sgn := 1.0
	if u < 0 {
		sgn = -1.0
	}
	return -scale * sgn * math.Log(1.0-2.0*math.Abs(u))
}

// SampleNormalBoxMuller draws a random variable from Normal(mean, stdDev) using the Box-Muller transform.
func SampleNormalBoxMuller(mean, stdDev float64) float64 {
	u1 := RandomFloat()
	u2 := RandomFloat()
	if u1 < 1e-10 {
		u1 = 1e-10
	}
	z0 := math.Sqrt(-2.0*math.Log(u1)) * math.Cos(2.0*math.Pi*u2)
	return mean + z0*stdDev
}

// GetEpsilonForPrivacyLevel resolves default differential privacy budget epsilon based on tier.
func GetEpsilonForPrivacyLevel(level PrivacyLevel, configuredEpsilon float64) float64 {
	if configuredEpsilon > 0.0 {
		return configuredEpsilon
	}
	switch level {
	case PrivacyRestricted:
		return 0.1 // Strictest noise
	case PrivacyConfidential:
		return 0.5
	case PrivacyInternal:
		return 1.0
	case PrivacyPublic:
		fallthrough
	default:
		return 2.0 // Lower noise, higher statistical accuracy
	}
}

// CalculateReidentificationRisk calculates empirical re-identification vulnerability (0.0 to 100.0%).
func CalculateReidentificationRisk(epsilon float64, rowCount int, fieldCount int) float64 {
	// Re-identification risk increases with lower noise (higher epsilon) and higher dimensional field count
	baseRisk := (1.0 - math.Exp(-epsilon*0.75)) * 100.0
	dimensionFactor := math.Min(1.5, 1.0+float64(fieldCount)*0.03)
	risk := baseRisk * dimensionFactor

	if risk > 99.0 {
		risk = 99.0
	}
	if risk < 1.0 {
		risk = 1.0
	}
	return math.Round(risk*10) / 10
}

