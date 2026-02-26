package gui

import "limbs/internal/exporter"

type ExportRequest struct {
	SourceMount  string `json:"sourceMount"`
	ProjectsRoot string `json:"projectsRoot"`
	SamplesRoot  string `json:"samplesRoot"`
	USBDrive     string `json:"usbDrive"`
	ProjectName  string `json:"projectName"`
	ProjectFile  string `json:"projectFile"`
	DestRoot     string `json:"destRoot"`
	LimbsRoot    string `json:"limbsRoot"`
	Zip          bool   `json:"zip"`
	AllowMissing bool   `json:"allowMissing"`
}

type ErrorResponse struct {
	Error string `json:"error"`
}

type ExportResponse struct {
	ProjectName string         `json:"projectName"`
	OutputDir   string         `json:"outputDir"`
	ZipPath     string         `json:"zipPath,omitempty"`
	Stats       ExportStatsDTO `json:"stats"`
}

type ExportStatsDTO struct {
	ReferencesFound  int                `json:"referencesFound"`
	UniqueReferences int                `json:"uniqueReferences"`
	SamplesCopied    int                `json:"samplesCopied"`
	MissingSamples   []MissingSampleDTO `json:"missingSamples"`
	Collisions       []CollisionDTO     `json:"collisions"`
}

type MissingSampleDTO struct {
	ReferencePath string `json:"referencePath"`
	ResolvedPath  string `json:"resolvedPath"`
	Reason        string `json:"reason"`
}

type CollisionDTO struct {
	SourcePath   string `json:"sourcePath"`
	AssignedName string `json:"assignedName"`
}

func toExportResponse(res exporter.Result) ExportResponse {
	out := ExportResponse{
		ProjectName: res.ProjectName,
		OutputDir:   res.OutputDir,
		ZipPath:     res.ZipPath,
		Stats: ExportStatsDTO{
			ReferencesFound:  res.Stats.ReferencesFound,
			UniqueReferences: res.Stats.UniqueReferences,
			SamplesCopied:    res.Stats.SamplesCopied,
			MissingSamples:   make([]MissingSampleDTO, 0, len(res.Stats.MissingSamples)),
			Collisions:       make([]CollisionDTO, 0, len(res.Stats.Collisions)),
		},
	}

	for _, m := range res.Stats.MissingSamples {
		out.Stats.MissingSamples = append(out.Stats.MissingSamples, MissingSampleDTO{
			ReferencePath: m.ReferencePath,
			ResolvedPath:  m.ResolvedPath,
			Reason:        m.Reason,
		})
	}

	for _, c := range res.Stats.Collisions {
		out.Stats.Collisions = append(out.Stats.Collisions, CollisionDTO{
			SourcePath:   c.SourcePath,
			AssignedName: c.AssignedName,
		})
	}

	return out
}
