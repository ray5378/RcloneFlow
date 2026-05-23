package controller

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIsCASCompatibleRunSummary_Nil(t *testing.T) {
	assert.False(t, isCASCompatibleRunSummary(nil))
}

func TestIsCASCompatibleRunSummary_Empty(t *testing.T) {
	assert.False(t, isCASCompatibleRunSummary(map[string]any{}))
}

func TestIsCASCompatibleRunSummary_TransferDefaults_BoolTrue(t *testing.T) {
	sum := map[string]any{
		"transferDefaults": map[string]any{
			"openlistCasCompatible": true,
		},
	}
	assert.True(t, isCASCompatibleRunSummary(sum))
}

func TestIsCASCompatibleRunSummary_TransferDefaults_BoolFalse(t *testing.T) {
	sum := map[string]any{
		"transferDefaults": map[string]any{
			"openlistCasCompatible": false,
		},
	}
	assert.False(t, isCASCompatibleRunSummary(sum))
}

func TestIsCASCompatibleRunSummary_TransferDefaults_StringTrue(t *testing.T) {
	sum := map[string]any{
		"transferDefaults": map[string]any{
			"openlistCasCompatible": "true",
		},
	}
	assert.True(t, isCASCompatibleRunSummary(sum))
}

func TestIsCASCompatibleRunSummary_TransferDefaults_StringFalse(t *testing.T) {
	sum := map[string]any{
		"transferDefaults": map[string]any{
			"openlistCasCompatible": "false",
		},
	}
	assert.False(t, isCASCompatibleRunSummary(sum))
}

func TestIsCASCompatibleRunSummary_TransferDefaults_StringWithSpaces(t *testing.T) {
	sum := map[string]any{
		"transferDefaults": map[string]any{
			"openlistCasCompatible": "  true  ",
		},
	}
	assert.True(t, isCASCompatibleRunSummary(sum))
}

func TestIsCASCompatibleRunSummary_TransferDefaults_CaseInsensitive(t *testing.T) {
	tests := []string{"TRUE", "True", "TRUE", "TrUe"}
	for _, val := range tests {
		sum := map[string]any{
			"transferDefaults": map[string]any{
				"openlistCasCompatible": val,
			},
		}
		assert.True(t, isCASCompatibleRunSummary(sum), "should be true for %s", val)
	}
}

func TestIsCASCompatibleRunSummary_TransferDefaults_NilValue(t *testing.T) {
	sum := map[string]any{
		"transferDefaults": map[string]any{
			"openlistCasCompatible": nil,
		},
	}
	assert.False(t, isCASCompatibleRunSummary(sum))
}

func TestIsCASCompatibleRunSummary_TrackingMode_CAS(t *testing.T) {
	sum := map[string]any{
		"trackingMode": "cas",
	}
	assert.True(t, isCASCompatibleRunSummary(sum))
}

func TestIsCASCompatibleRunSummary_TrackingMode_Copy(t *testing.T) {
	sum := map[string]any{
		"trackingMode": "copy",
	}
	assert.False(t, isCASCompatibleRunSummary(sum))
}

func TestIsCASCompatibleRunSummary_TrackingMode_CaseInsensitive(t *testing.T) {
	tests := []string{"CAS", "Cas", "cAs"}
	for _, mode := range tests {
		sum := map[string]any{
			"trackingMode": mode,
		}
		assert.True(t, isCASCompatibleRunSummary(sum), "should be true for mode %s", mode)
	}
}

func TestIsCASCompatibleRunSummary_ActiveTransfer_CAS(t *testing.T) {
	sum := map[string]any{
		"activeTransfer": map[string]any{
			"trackingMode": "cas",
		},
	}
	assert.True(t, isCASCompatibleRunSummary(sum))
}

func TestIsCASCompatibleRunSummary_ActiveTransfer_Copy(t *testing.T) {
	sum := map[string]any{
		"activeTransfer": map[string]any{
			"trackingMode": "copy",
		},
	}
	assert.False(t, isCASCompatibleRunSummary(sum))
}

func TestIsCASCompatibleRunSummary_Priority(t *testing.T) {
	sum := map[string]any{
		"transferDefaults": map[string]any{
			"openlistCasCompatible": true,
		},
		"trackingMode": "copy",
		"activeTransfer": map[string]any{
			"trackingMode": "move",
		},
	}
	assert.True(t, isCASCompatibleRunSummary(sum))
}

func TestEnrichRowMapSizesFromFinalSummary_Nil(t *testing.T) {
	rows := []map[string]any{
		{"path": "file1.txt", "sizeBytes": int64(0)},
	}
	enrichRowMapSizesFromFinalSummary(rows, nil)
	assert.Equal(t, int64(0), rows[0]["sizeBytes"])
}

func TestEnrichRowMapSizesFromFinalSummary_EmptySummary(t *testing.T) {
	rows := []map[string]any{
		{"path": "file1.txt", "sizeBytes": int64(0)},
	}
	enrichRowMapSizesFromFinalSummary(rows, map[string]any{})
	assert.Equal(t, int64(0), rows[0]["sizeBytes"])
}

func TestEnrichRowMapSizesFromFinalSummary_NoActiveTransfer(t *testing.T) {
	rows := []map[string]any{
		{"path": "file1.txt", "sizeBytes": int64(0)},
	}
	summary := map[string]any{
		"someOther": "data",
	}
	enrichRowMapSizesFromFinalSummary(rows, summary)
	assert.Equal(t, int64(0), rows[0]["sizeBytes"])
}

func TestEnrichRowMapSizesFromFinalSummary_WithActiveTransfer(t *testing.T) {
	rows := []map[string]any{
		{"path": "file1.txt", "sizeBytes": int64(0)},
		{"path": "file2.txt", "sizeBytes": int64(0)},
		{"path": "file3.txt", "sizeBytes": int64(100)},
	}
	summary := map[string]any{
		"activeTransfer": map[string]any{
			"completed": []map[string]any{
				{"path": "file1.txt", "sizeBytes": int64(1024)},
				{"path": "file2.txt", "sizeBytes": int64(2048)},
			},
		},
	}
	enrichRowMapSizesFromFinalSummary(rows, summary)
	assert.Equal(t, int64(1024), rows[0]["sizeBytes"])
	assert.Equal(t, int64(2048), rows[1]["sizeBytes"])
	assert.Equal(t, int64(100), rows[2]["sizeBytes"])
}

func TestEnrichRowMapSizesFromFinalSummary_WindowsPath(t *testing.T) {
	rows := []map[string]any{
		{"path": "dir/file.txt", "sizeBytes": int64(0)},
	}
	summary := map[string]any{
		"activeTransfer": map[string]any{
			"completed": []map[string]any{
				{"path": "dir\\file.txt", "sizeBytes": int64(512)},
			},
		},
	}
	enrichRowMapSizesFromFinalSummary(rows, summary)
	assert.Equal(t, int64(512), rows[0]["sizeBytes"])
}

func TestEnrichRowMapSizesFromFinalSummary_EmptyPaths(t *testing.T) {
	rows := []map[string]any{
		{"path": "", "sizeBytes": int64(0)},
	}
	summary := map[string]any{
		"activeTransfer": map[string]any{
			"completed": []map[string]any{
				{"path": "", "sizeBytes": int64(100)},
			},
		},
	}
	enrichRowMapSizesFromFinalSummary(rows, summary)
	assert.Equal(t, int64(0), rows[0]["sizeBytes"])
}

func TestEnrichRowMapSizesFromFinalSummary_ZeroSize(t *testing.T) {
	rows := []map[string]any{
		{"path": "file1.txt", "sizeBytes": int64(0)},
	}
	summary := map[string]any{
		"activeTransfer": map[string]any{
			"completed": []map[string]any{
				{"path": "file1.txt", "sizeBytes": int64(0)},
			},
		},
	}
	enrichRowMapSizesFromFinalSummary(rows, summary)
	assert.Equal(t, int64(0), rows[0]["sizeBytes"])
}

func TestEnrichRowMapSizesFromFinalSummary_CompletedIsNotList(t *testing.T) {
	rows := []map[string]any{
		{"path": "file1.txt", "sizeBytes": int64(0)},
	}
	summary := map[string]any{
		"activeTransfer": map[string]any{
			"completed": "not a list",
		},
	}
	enrichRowMapSizesFromFinalSummary(rows, summary)
	assert.Equal(t, int64(0), rows[0]["sizeBytes"])
}

func TestEnrichRowMapSizesFromFinalSummary_CompletedAsListOfAny(t *testing.T) {
	rows := []map[string]any{
		{"path": "file1.txt", "sizeBytes": 0},
	}
	summary := map[string]any{
		"activeTransfer": map[string]any{
			"completed": []any{
				map[string]any{"path": "file1.txt", "sizeBytes": int64(999)},
			},
		},
	}
	enrichRowMapSizesFromFinalSummary(rows, summary)
	assert.Equal(t, int64(999), rows[0]["sizeBytes"])
}

func TestIsCASAttemptObjectNotFoundSummaryRow_True(t *testing.T) {
	assert.True(t, isCASAttemptObjectNotFoundSummaryRow("Attempt 1/3 failed with 5 errors and", "object not found"))
	assert.True(t, isCASAttemptObjectNotFoundSummaryRow("<nil>", "Attempt 1/1 failed with 2 errors and: object not found"))
	assert.True(t, isCASAttemptObjectNotFoundSummaryRow("<nil>", "Attempt 1/1 failed with 2 errors and: permission denied"))
}

func TestIsCASAttemptObjectNotFoundSummaryRow_False(t *testing.T) {
	assert.False(t, isCASAttemptObjectNotFoundSummaryRow("file.txt", "permission denied"))
	assert.False(t, isCASAttemptObjectNotFoundSummaryRow("file.txt", "object not found"))
}
