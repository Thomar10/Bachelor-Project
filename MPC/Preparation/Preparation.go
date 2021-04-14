package Preparation

import (
	finite "MPC/Finite-fields"
	"MPC/Finite-fields/Binary"
	"math/big"
)

func CreateHyperMatrix(partySize int, field finite.Finite) [][]finite.Number {
	a := make([]finite.Number, partySize)
	b := make([]finite.Number, partySize)
	for i := 1; i <= partySize; i++ {
		a[i - 1] = finite.Number{
			Prime: big.NewInt(int64(i)),
			Binary: Binary.ConvertXToByte(i),
		}
		b[i - 1] = finite.Number{
			Prime: big.NewInt(int64(i + partySize + 1)),
			Binary: Binary.ConvertXToByte(i + partySize + 1),
		}
	}
	matrix := make([][]finite.Number, partySize)
	for i, _ := range matrix {
		matrix[i] = make([]finite.Number, partySize)
		for j, _ := range matrix {
			matrix[i][j] = finite.Number{
				Prime: big.NewInt( 1),
				Binary: Binary.ConvertXToByte(1),
			}
		}
	}
	for i, _ := range matrix {
		for j, _ := range matrix {
			for k := 0; k < partySize; k++ {
				if k == j {
					continue
				}else {
					//ak-neg
					ak := field.Add(a[k], finite.Number{
						Prime: field.GetSize().Prime,
						Binary: Binary.ConvertXToByte(0),
					})
					biak := field.Add(b[i], ak)
					ajak := field.Add(a[j], ak)
					ajakInverse := field.FindInverse(ajak, field.GetSize())
					biakajak := field.Mul(biak, ajakInverse)
					matrix[i][j] = field.Mul(biakajak, matrix[i][j])
					//((b[i] - a[k]) / (a[j] - a[k]) * matrix[i][j]) % 17
				}
			}
		}
	}
	return matrix
}

func ExtractRandomness(x []finite.Number, matrix [][]finite.Number, field finite.Finite, corrupts int) []finite.Number{
	y := make([]finite.Number, len(x))
	for i := 0; i < len(matrix); i++ {
		y[i] = finite.Number{Prime: big.NewInt(0), Binary: Binary.ConvertXToByte(0)}
		for j := 0; j < len(matrix[i]); j++ {
			y[i] = field.Add(y[i], field.Mul(matrix[i][j], x[j]))
			//y[i] = y[i] + matrix[i][j] * x[j]
		}
	}
	return y[:len(x) - corrupts]
}
