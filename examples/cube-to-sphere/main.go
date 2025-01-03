package main

import (
	"math"

	"github.com/EliCDavis/polyform/formats/gltf"
	"github.com/EliCDavis/polyform/generator"
	"github.com/EliCDavis/polyform/generator/artifact"
	"github.com/EliCDavis/polyform/generator/parameter"
	"github.com/EliCDavis/polyform/modeling"
	"github.com/EliCDavis/polyform/modeling/marching"
	"github.com/EliCDavis/polyform/modeling/meshops"
	"github.com/EliCDavis/polyform/nodes"
	"github.com/EliCDavis/vector/vector3"
)

type CubeToSphereAnimation = nodes.StructNode[modeling.Mesh, CubeToSphereAnimationData]

type CubeToSphereAnimationData struct {
	Time       nodes.NodeOutput[float64]
	Resolution nodes.NodeOutput[float64]
}

func (csa CubeToSphereAnimationData) Process() (modeling.Mesh, error) {
	time := math.Max(math.Min(csa.Time.Value(), 1), 0)

	box := marching.Box(vector3.Float64{}, vector3.New(0.7, 0.5, 0.5), 1)
	sphere := marching.Sphere(vector3.Float64{}, 0.5*time, 1)

	return marching.
		CombineFields(box, sphere).
		March(modeling.PositionAttribute, csa.Resolution.Value(), 0), nil
}

func main() {

	animation := CubeToSphereAnimation{
		Data: CubeToSphereAnimationData{
			Time: &parameter.Float64{
				Name:         "Time",
				DefaultValue: 0.,
			},
			Resolution: &parameter.Float64{
				Name:         "Mesh Resolution",
				DefaultValue: 30,
			},
		},
	}

	smoothedMeshNode := &meshops.SmoothNormalsNode{
		Data: meshops.SmoothNormalsNodeData{
			Mesh: &meshops.LaplacianSmoothNode{
				Data: meshops.LaplacianSmoothNodeData{
					Mesh: animation.Out(),
					Iterations: &parameter.Int{
						Name:         "Smoothing Iterations",
						DefaultValue: 20,
					},
					SmoothingFactor: &parameter.Float64{
						Name:         "Smoothing Factor",
						DefaultValue: .1,
					},
				},
			},
		},
	}

	app := generator.App{
		Name:        "Cube to Sphere",
		Description: "Smoothly blend a cube into a sphere",
		Version:     "1.0.0",
		Producers: map[string]nodes.NodeOutput[generator.Artifact]{
			"mesh.glb": &GltfArtifact{
				Data: GltfArtifactData{
					Mesh: smoothedMeshNode,
				},
			},
		},
	}

	err := app.Run()

	if err != nil {
		panic(err)
	}
}

type GltfArtifact = nodes.StructNode[generator.Artifact, GltfArtifactData]

type GltfArtifactData struct {
	Mesh nodes.NodeOutput[modeling.Mesh]
}

func (csa GltfArtifactData) Process() (generator.Artifact, error) {
	mesh := csa.Mesh.Value()
	return &artifact.Gltf{
		Scene: gltf.PolyformScene{
			Models: []gltf.PolyformModel{
				{
					Name: "Mesh",
					Mesh: &mesh,
				},
			},
		},
	}, nil
}
