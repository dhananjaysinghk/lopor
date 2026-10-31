package datasynth

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"math"
	"strings"
	"time"
)

var (
	firstNames = []string{"Alex", "Jordan", "Taylor", "Morgan", "Sam", "Chris", "Casey", "Riley", "Avery", "Jamie"}
	lastNames  = []string{"Chen", "Smith", "Johnson", "Patel", "Garcia", "Miller", "Davis", "Rodriguez", "Martinez", "Wilson"}
	domains    = []string{"synthetic-corp.test", "anonymized.example", "mock-enterprise.internal", "privacy-cloud.local"}
)

// SynthesizeRow generates a single synthetic record adhering to column specs and differential privacy noise.
func SynthesizeRow(fields []SyntheticFieldSchema, epsilon float64) map[string]interface{} {
	row := make(map[string]interface{})

	for _, f := range fields {
		if f.IsNullable && f.NullProbability > 0 && RandomFloat() < f.NullProbability {
			row[f.Name] = nil
			continue
		}

		switch f.Type {
		case FieldTypeUUID:
			row[f.Name] = generateSyntheticUUID()

		case FieldTypeName:
			fn := firstNames[int(RandomFloat()*float64(len(firstNames)))]
			ln := lastNames[int(RandomFloat()*float64(len(lastNames)))]
			row[f.Name] = fmt.Sprintf("%s %s", fn, ln)

		case FieldTypeEmail:
			fn := strings.ToLower(firstNames[int(RandomFloat()*float64(len(firstNames)))])
			ln := strings.ToLower(lastNames[int(RandomFloat()*float64(len(lastNames)))])
			dom := domains[int(RandomFloat()*float64(len(domains)))]
			randTag := int(RandomFloat() * 899) + 100
			row[f.Name] = fmt.Sprintf("%s.%s%d@%s", fn, ln, randTag, dom)

		case FieldTypePhone:
			area := int(RandomFloat()*800) + 200
			prefix := int(RandomFloat()*899) + 100
			line := int(RandomFloat()*8999) + 1000
			row[f.Name] = fmt.Sprintf("+1-%03d-%03d-%04d", area, prefix, line)

		case FieldTypeIntegerRange:
			min := f.MinValue
			max := f.MaxValue
			if max <= min {
				max = min + 100
			}
			baseVal := min + RandomFloat()*(max-min)
			// Apply differential Laplace noise inversely proportional to epsilon
			noiseScale := (max - min) * 0.05 / epsilon
			noisyVal := baseVal + SampleLaplace(noiseScale)
			if noisyVal < min {
				noisyVal = min
			}
			if noisyVal > max {
				noisyVal = max
			}
			row[f.Name] = int(math.Round(noisyVal))

		case FieldTypeFloatNormal:
			mean := f.Mean
			stdDev := f.StdDev
			if stdDev <= 0 {
				stdDev = 10.0
			}
			baseVal := SampleNormalBoxMuller(mean, stdDev)
			noiseScale := stdDev * 0.1 / epsilon
			noisyVal := baseVal + SampleLaplace(noiseScale)
			row[f.Name] = math.Round(noisyVal*100) / 100

		case FieldTypeCategorical:
			if len(f.Categories) == 0 {
				row[f.Name] = "DEFAULT"
			} else {
				idx := sampleCategoricalIndex(f.Categories, f.Weights)
				row[f.Name] = f.Categories[idx]
			}

		case FieldTypeTimestamp:
			// Random timestamp within the past 180 days
			daysAgo := RandomFloat() * 180.0
			secAgo := int64(daysAgo * 86400)
			ts := time.Now().UTC().Add(-time.Duration(secAgo) * time.Second)
			row[f.Name] = ts.Format(time.RFC3339)

		case FieldTypeFreeText:
			fallthrough
		default:
			row[f.Name] = fmt.Sprintf("Synthetic transaction record ref #%s", generateShortHex(4))
		}
	}

	return row
}

func sampleCategoricalIndex(categories []string, weights []float64) int {
	if len(weights) != len(categories) {
		return int(RandomFloat() * float64(len(categories)))
	}

	var sum float64
	for _, w := range weights {
		sum += w
	}
	if sum <= 0 {
		return int(RandomFloat() * float64(len(categories)))
	}

	r := RandomFloat() * sum
	var cumulative float64
	for i, w := range weights {
		cumulative += w
		if r <= cumulative {
			return i
		}
	}
	return len(categories) - 1
}

func generateSyntheticUUID() string {
	b := make([]byte, 16)
	_, _ = rand.Read(b)
	b[6] = (b[6] & 0x0f) | 0x40 // Version 4
	b[8] = (b[8] & 0x3f) | 0x80 // Variant RFC4122
	return fmt.Sprintf("synth-%08x-%04x-%04x-%04x-%012x",
		b[0:4], b[4:6], b[6:8], b[8:10], b[10:16])
}

func generateShortHex(numBytes int) string {
	b := make([]byte, numBytes)
	_, _ = rand.Read(b)
	return hex.EncodeToString(b)
}

