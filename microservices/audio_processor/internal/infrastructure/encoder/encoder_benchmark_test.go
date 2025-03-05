package encoder_test

import (
	"bytes"
	"context"
	"github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/audio_processor/internal/infrastructure/encoder"
	"github.com/Tomas-vilte/ButakeroMusicBotGo/microservices/audio_processor/internal/logger"
	"github.com/stretchr/testify/require"
	"io"
	"runtime"
	"testing"
	"time"
)

func generateTestAudioData(sizeMB int) io.Reader {
	audioData := make([]byte, sizeMB*1024*1024)
	for i := range audioData {
		audioData[i] = byte(i % 256)
	}
	return bytes.NewReader(audioData)
}

func BenchmarkAudioEncoding(b *testing.B) {
	testCases := []struct {
		name string
		size int
	}{
		{"1MB Audio", 1},
		{"2MB Audio", 2},
	}

	for _, tc := range testCases {
		b.Run(tc.name, func(b *testing.B) {
			log, err := logger.NewZapLogger()
			require.NoError(b, err)
			ffmpegEncoder := encoder.NewFFmpegEncoder(log)

			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				audioData := generateTestAudioData(tc.size)

				ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)

				session, err := ffmpegEncoder.Encode(ctx, audioData, encoder.StdEncodeOptions)
				if err != nil {
					cancel()
					b.Fatalf("Failed to start encoding: %v", err)
				}

				for {
					_, err := session.ReadFrame()
					if err == io.EOF {
						break
					}
					if err != nil {
						cancel()
						session.Cleanup()
						b.Fatalf("Error reading frame: %v", err)
					}
				}

				// Cleanup
				session.Cleanup()
				cancel()
			}
		})
	}
}

func BenchmarkResourceUsage(b *testing.B) {
	testCases := []struct {
		name string
		size int
	}{
		{"1MB Audio", 1},
		{"2MB Audio", 2},
	}

	for _, tc := range testCases {
		b.Run(tc.name, func(b *testing.B) {
			log, err := logger.NewZapLogger()
			require.NoError(b, err)
			ffmpegEncoder := encoder.NewFFmpegEncoder(log)

			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				audioData := generateTestAudioData(tc.size)

				ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)

				startTime := time.Now()
				var memStats runtime.MemStats
				runtime.ReadMemStats(&memStats)
				startAlloc := memStats.Alloc

				session, err := ffmpegEncoder.Encode(ctx, audioData, encoder.StdEncodeOptions)
				if err != nil {
					cancel()
					b.Fatalf("Failed to start encoding: %v", err)
				}

				for {
					_, err := session.ReadFrame()
					if err == io.EOF {
						break
					}
					if err != nil {
						cancel()
						session.Cleanup()
						b.Fatalf("Error reading frame: %v", err)
					}
				}

				endTime := time.Now()
				runtime.ReadMemStats(&memStats)
				endAlloc := memStats.Alloc

				b.ReportMetric(float64(endTime.Sub(startTime).Milliseconds()), "ms")
				b.ReportMetric(float64(endAlloc-startAlloc)/1024, "kB_allocated")

				session.Cleanup()
				cancel()
			}
		})
	}
}
