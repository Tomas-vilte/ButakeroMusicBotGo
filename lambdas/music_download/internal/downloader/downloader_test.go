package downloader

import (
	"errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"testing"
)

func TestYtDlpDownloader_DownloadSong(t *testing.T) {

	t.Run("Successful download and upload", func(t *testing.T) {
		mockExecutor := new(MockCommandExecutor)
		mockUploader := new(MockUploader)
		mockLogger := new(MockLogger)

		downloader := NewDownloader(mockUploader, mockLogger, mockExecutor, "/yt-dlp")
		songURL := "https://www.youtube.com/watch?v=CnuFA6PkOT8"
		key := "song.m4a"
		expectedOutput := []byte("song data")

		mockExecutor.On("ExecuteCommand", mock.Anything, "sh", mock.Anything).Return(expectedOutput, nil)
		mockUploader.On("UploadToS3", mock.Anything, mock.AnythingOfType("*bytes.Reader"), key).Return(nil)

		err := downloader.DownloadSong(songURL, key)

		assert.NoError(t, err)
		mockExecutor.AssertExpectations(t)
		mockUploader.AssertExpectations(t)
	})

	t.Run("ExecuteCommand error", func(t *testing.T) {
		mockExecutor := new(MockCommandExecutor)
		mockUploader := new(MockUploader)
		mockLogger := new(MockLogger)

		downloader := NewDownloader(mockUploader, mockLogger, mockExecutor, "/yt-dlp")
		songURL := "https://www.youtube.com/watch?v=CnuFA6PkOT8"
		key := "song.m4a"
		expectedError := errors.New("command execution failed")

		mockExecutor.On("ExecuteCommand", mock.Anything, "sh", mock.Anything).Return([]byte{}, expectedError)
		mockLogger.On("Error", "Error al ejecutar yt-dlp", mock.Anything)

		err := downloader.DownloadSong(songURL, key)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "error al ejecutar yt-dlp")
		mockExecutor.AssertExpectations(t)
		mockLogger.AssertExpectations(t)
	})

	t.Run("UploadToS3 error", func(t *testing.T) {
		mockExecutor := new(MockCommandExecutor)
		mockUploader := new(MockUploader)
		mockLogger := new(MockLogger)

		downloader := NewDownloader(mockUploader, mockLogger, mockExecutor, "/yt-dlp")
		songURL := "https://www.youtube.com/watch?v=CnuFA6PkOT8"
		key := "song.m4a"
		expectedOutput := []byte("song data")
		expectedError := errors.New("upload failed")

		mockExecutor.On("ExecuteCommand", mock.Anything, "sh", mock.Anything).Return(expectedOutput, nil)
		mockUploader.On("UploadToS3", mock.Anything, mock.AnythingOfType("*bytes.Reader"), key).Return(expectedError)
		mockLogger.On("Error", "Error al subir datos a s3", mock.Anything)

		err := downloader.DownloadSong(songURL, key)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "error al subir datos a S3")
		mockExecutor.AssertExpectations(t)
		mockUploader.AssertExpectations(t)
		mockLogger.AssertExpectations(t)
	})
}
