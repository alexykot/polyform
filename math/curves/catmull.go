package curves

import (
	"math"
	"sort"

	"github.com/EliCDavis/polyform/nodes"
	"github.com/EliCDavis/vector/vector3"
)

const defaultEpsilon = 0.001

type CatmullRomCurveParameters struct {
	P0, P1, P2, P3 vector3.Float64
	Alpha          float64
	Epsilon        float64
}

func (crcp CatmullRomCurveParameters) Curve() CatmullRomCurve {
	epsilon := defaultEpsilon
	if crcp.Epsilon > 0 {
		epsilon = crcp.Epsilon
	}
	return CatmullRomCurve{
		p0:      crcp.P0,
		p1:      crcp.P1,
		p2:      crcp.P2,
		p3:      crcp.P3,
		alpha:   crcp.Alpha,
		epsilon: epsilon,
	}
}

type catmullRomCurveDistSegment struct {
	time     float64
	distance float64
	point    vector3.Float64
}

type sortByTime []catmullRomCurveDistSegment

func (a sortByTime) Len() int           { return len(a) }
func (a sortByTime) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a sortByTime) Less(i, j int) bool { return a[i].time < a[j].time }

type CatmullRomCurve struct {
	p0, p1, p2, p3 vector3.Float64
	alpha          float64
	epsilon        float64

	distance *float64
	segments []catmullRomCurveDistSegment
}

func (crc *CatmullRomCurve) calcLength(a, b float64) float64 {
	start := crc.Time(a)
	end := crc.Time(b)

	dist := end.Distance(start)
	if dist > crc.epsilon {
		half := a + ((b - a) / 2)
		return crc.calcLength(a, half) + crc.calcLength(half, b)
	}

	crc.segments = append(
		crc.segments,
		catmullRomCurveDistSegment{
			time:  a,
			point: start,
		},
		catmullRomCurveDistSegment{
			time:  b,
			point: end,
		},
	)

	return dist
}

func (crc *CatmullRomCurve) populateHelperData() {
	if crc.distance != nil {
		return
	}

	dist := crc.calcLength(0, 1)
	crc.distance = &dist
	sort.Sort(sortByTime(crc.segments))
	for i := 1; i < len(crc.segments); i++ {
		previous := crc.segments[i-1]
		seg := crc.segments[i]
		seg.distance = previous.distance + seg.point.Distance(previous.point)
		crc.segments[i] = seg
	}
}

func (crc *CatmullRomCurve) Length() float64 {
	crc.populateHelperData()
	return *crc.distance
}

func (crc CatmullRomCurve) Time(t float64) vector3.Float64 {
	// calculate knots
	const k0 = 0
	k1 := crc.getKnotInterval(crc.p0, crc.p1)
	k2 := crc.getKnotInterval(crc.p1, crc.p2) + k1
	k3 := crc.getKnotInterval(crc.p2, crc.p3) + k2

	// evaluate the point
	u := lerpUnclamped(k1, k2, t)
	A1 := remap(k0, k1, crc.p0, crc.p1, u)
	A2 := remap(k1, k2, crc.p1, crc.p2, u)
	A3 := remap(k2, k3, crc.p2, crc.p3, u)
	B1 := remap(k0, k2, A1, A2, u)
	B2 := remap(k1, k3, A2, A3, u)
	return remap(k1, k2, B1, B2, u)
}

func (crc *CatmullRomCurve) Distance(distance float64) vector3.Float64 {
	crc.populateHelperData()

	if distance <= 0 {
		return crc.segments[0].point
	}

	if distance >= *crc.distance {
		return crc.segments[len(crc.segments)-1].point
	}

	start := 0
	end := len(crc.segments) - 1
	for start <= end {
		mid := int(math.Floor(float64(end+start) / 2.))

		// value is in interval from previous to current element
		if distance >= crc.segments[mid-1].distance && distance <= crc.segments[mid].distance {
			a, b := crc.segments[mid].distance, crc.segments[mid-1].distance

			t := (distance - a) / (b - a)

			aPoint, bPoint := crc.segments[mid].point, crc.segments[mid-1].point
			return vector3.Lerp(aPoint, bPoint, t)

			// if math.Abs(distance-crc.segments[mid-1].distance) < math.Abs(distance-crc.segments[mid].distance) {
			// 	return crc.segments[mid-1].point
			// } else {
			// 	return crc.segments[mid].point
			// }
		} else {
			if crc.segments[mid].distance < distance {
				start = mid + 1
			} else {
				end = mid - 1
			}
		}
	}

	panic("uhhh")
}

func (crc CatmullRomCurve) getKnotInterval(a, b vector3.Float64) float64 {
	return math.Pow(a.Sub(b).LengthSquared(), 0.5*crc.alpha)
}

func remap(a, b float64, c, d vector3.Float64, u float64) vector3.Float64 {
	return vector3.Lerp(c, d, (u-a)/(b-a))
}

func lerpUnclamped(a, b, t float64) float64 {
	dir := b - a
	return a + (dir * t)
}

type CatmullRomSplineParameters struct {
	Points  []vector3.Float64
	Alpha   float64
	Epsilon float64
}

func (crcp CatmullRomSplineParameters) Spline() CatmullRomSpline {
	epsilon := defaultEpsilon
	if crcp.Epsilon > 0 {
		epsilon = crcp.Epsilon
	}

	if len(crcp.Points) == 0 {
		return CatmullRomSpline{
			alpha: crcp.Alpha,
			curves: []*CatmullRomCurve{{
				alpha:   crcp.Alpha,
				epsilon: epsilon,
			}},
		}
	}

	if len(crcp.Points) == 1 {
		return CatmullRomSpline{
			alpha: crcp.Alpha,
			curves: []*CatmullRomCurve{{
				alpha:   crcp.Alpha,
				epsilon: epsilon,
				p0:      crcp.Points[0],
				p1:      crcp.Points[0],
				p2:      crcp.Points[0],
				p3:      crcp.Points[0],
			}},
		}
	}

	if len(crcp.Points) == 2 {
		return CatmullRomSpline{
			alpha: crcp.Alpha,
			curves: []*CatmullRomCurve{{
				alpha:   crcp.Alpha,
				epsilon: epsilon,
				p0:      crcp.Points[0],
				p1:      crcp.Points[0],
				p2:      crcp.Points[1],
				p3:      crcp.Points[1],
			}},
		}
	}

	// One of the least satisfying arbitrary decisions I've made in this
	// library
	if len(crcp.Points) == 3 {
		return CatmullRomSpline{
			alpha: crcp.Alpha,
			curves: []*CatmullRomCurve{{
				alpha:   crcp.Alpha,
				epsilon: epsilon,
				p0:      crcp.Points[0],
				p1:      crcp.Points[1],
				p2:      crcp.Points[2],
				p3:      crcp.Points[2],
			}},
		}
	}

	curves := make([]*CatmullRomCurve, len(crcp.Points)-3)
	for i := 0; i < len(crcp.Points)-3; i++ {
		curves[i] = &CatmullRomCurve{
			p0:      crcp.Points[i],
			p1:      crcp.Points[i+1],
			p2:      crcp.Points[i+2],
			p3:      crcp.Points[i+3],
			alpha:   crcp.Alpha,
			epsilon: epsilon,
		}
	}

	return CatmullRomSpline{
		alpha:  crcp.Alpha,
		curves: curves,
	}
}

type CatmullRomSpline struct {
	alpha float64

	curves   []*CatmullRomCurve
	distance *float64
}

func (crc *CatmullRomSpline) Length() float64 {
	if crc.distance == nil {
		dist := 0.
		for _, curve := range crc.curves {
			dist += curve.Length()
		}
		crc.distance = &dist
	}
	return *crc.distance
}

func (crc *CatmullRomSpline) Dir(distance float64) vector3.Float64 {
	inc := crc.Length() / 1000

	if distance-inc < 0 {
		return crc.At(distance + inc).Sub(crc.At(distance)).Normalized()
	}
	return crc.At(distance).Sub(crc.At(distance - inc)).Normalized()
}

func (crc *CatmullRomSpline) At(distance float64) vector3.Float64 {
	if distance <= 0 {
		return crc.curves[0].Time(0)
	}

	if distance >= crc.Length() {
		return crc.curves[len(crc.curves)-1].Time(1)
	}

	remainingDistance := distance
	var curveToEvaluation *CatmullRomCurve = nil
	for _, curve := range crc.curves {
		if curve.Length() > remainingDistance {
			curveToEvaluation = curve
			break
		}
		remainingDistance -= curve.Length()
	}

	return curveToEvaluation.Distance(remainingDistance)
}

type CatmullRomSplineNode = nodes.StructNode[Spline, CatmullRomSplineNodeData]

type CatmullRomSplineNodeData struct {
	Points nodes.NodeOutput[[]vector3.Float64]
	Alpha  nodes.NodeOutput[float64]
}

func (r CatmullRomSplineNodeData) Process() (Spline, error) {

	alpha := 0.
	if r.Alpha != nil {
		alpha = r.Alpha.Value()
	}

	if r.Points == nil {
		return nil, nil
	}

	points := r.Points.Value()
	if len(points) < 4 {
		return nil, nil
	}

	spline := CatmullRomSplineParameters{
		Points: points,
		Alpha:  alpha,
	}.Spline()

	// UGGO: Force a calculation to fill all the temp data
	// Prevents two threads calling length at the same time,
	// causing it to populate things twice
	spline.Length()

	return &spline, nil
}