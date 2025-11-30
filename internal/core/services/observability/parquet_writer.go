package observability

import (
	"bytes"
	"fmt"

	"github.com/parquet-go/parquet-go"
	"github.com/parquet-go/parquet-go/compress/zstd"

	"brokle/internal/core/domain/observability"
)

// ParquetWriter writes RawTelemetryRecord slices to Parquet format with ZSTD compression.
type ParquetWriter struct {
	compressionLevel int
}

// NewParquetWriter creates a new Parquet writer. Compression level: 1-22 (3 is balanced default).
func NewParquetWriter(compressionLevel int) *ParquetWriter {
	if compressionLevel < 1 {
		compressionLevel = 1
	}
	if compressionLevel > 22 {
		compressionLevel = 22
	}
	return &ParquetWriter{
		compressionLevel: compressionLevel,
	}
}

func (w *ParquetWriter) getZstdLevel() zstd.Level {
	switch {
	case w.compressionLevel <= 1:
		return zstd.SpeedFastest
	case w.compressionLevel <= 3:
		return zstd.SpeedDefault
	case w.compressionLevel <= 9:
		return zstd.SpeedBetterCompression
	default:
		return zstd.SpeedBestCompression
	}
}

// WriteRecords converts RawTelemetryRecord slice to Parquet bytes with ZSTD compression.
func (w *ParquetWriter) WriteRecords(records []observability.RawTelemetryRecord) ([]byte, error) {
	if len(records) == 0 {
		return nil, fmt.Errorf("no records to write")
	}

	var buf bytes.Buffer
	writer := parquet.NewGenericWriter[observability.RawTelemetryRecord](
		&buf,
		parquet.Compression(&zstd.Codec{Level: w.getZstdLevel()}),
	)

	_, err := writer.Write(records)
	if err != nil {
		return nil, fmt.Errorf("failed to write parquet records: %w", err)
	}

	if err := writer.Close(); err != nil {
		return nil, fmt.Errorf("failed to close parquet writer: %w", err)
	}

	return buf.Bytes(), nil
}
