package pts

import (
	"bufio"
	"errors"
	"io"
	"strconv"
	"strings"

	"github.com/EliCDavis/polyform/modeling"
	"github.com/EliCDavis/vector/vector3"
)

func ParseVec3(xTxt, yTxt, zTxt string) (vector3.Float64, error) {
	x, err := strconv.ParseFloat(xTxt, 64)
	if err != nil {
		return vector3.Zero[float64](), err
	}

	y, err := strconv.ParseFloat(yTxt, 64)
	if err != nil {
		return vector3.Zero[float64](), err
	}

	z, err := strconv.ParseFloat(zTxt, 64)
	if err != nil {
		return vector3.Zero[float64](), err
	}

	return vector3.New(x, y, z), nil
}

func ReadPointCloud(in io.Reader) (*modeling.Mesh, error) {
	scanner := bufio.NewScanner(in)

	scanner.Scan()
	countText := scanner.Text()
	parsedCount, err := strconv.Atoi(countText)
	if err != nil {
		return nil, err
	}

	readVerts := make([]vector3.Float64, parsedCount)
	readColors := make([]vector3.Float64, parsedCount)
	intensity := make([]float64, parsedCount)

	readIntensity := false
	readColor := false

	curLine := 0
	for scanner.Scan() && curLine < parsedCount {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			return nil, errors.New("encountered empty line in pts")
		}

		contents := strings.Fields(line)

		if len(contents) > 2 {
			pos, err := ParseVec3(contents[0], contents[1], contents[2])
			if err != nil {
				return nil, err
			}
			readVerts[curLine] = pos
		}

		if len(contents) > 3 {
			i, err := strconv.ParseFloat(contents[3], 64)
			if err != nil {
				return nil, err
			}
			intensity[curLine] = i / 255.
			readIntensity = true
		}

		if len(contents) > 6 {
			pos, err := ParseVec3(contents[4], contents[5], contents[6])
			if err != nil {
				return nil, err
			}
			readColors[curLine] = pos.DivByConstant(255)
			readColor = true
		}

		curLine++
	}

	if scanner.Err() != nil {
		return nil, scanner.Err()
	}

	v3Data := make(map[string][]vector3.Float64)

	v3Data[modeling.PositionAttribute] = readVerts
	if readColor {
		v3Data[modeling.ColorAttribute] = readColors
	}

	v1Data := make(map[string][]float64)

	if readIntensity {
		v1Data[modeling.IntensityAttribute] = intensity
	}

	finalMesh := modeling.NewPointCloud(nil, v3Data, nil, v1Data)

	return &finalMesh, nil
}
