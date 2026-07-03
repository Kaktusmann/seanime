package directstream

import (
	"errors"
	"net/http"
	"net/url"
	"seanime/internal/nativeplayer"

	"github.com/labstack/echo/v4"
)

// ServeEchoStream is a proxy to the current stream.
// It sits in between the player and the real stream (whether it's a local file, torrent, or http stream).
//
// If this is an EBML stream, it gets the range request from the player, processes it to stream the correct subtitles, and serves the video.
// Otherwise, it just serves the video.
func (m *Manager) ServeEchoStream() http.Handler {
	return m.getStreamHandler()
}

// ServeCurrentUrlStream serves the raw bytes of the currently playing URL stream (e.g. a plugin custom source).
// Unlike ServeEchoStream, it doesn't validate the stream ID since the caller (a Nakama peer) has no way of knowing it.
// It's used so that watch party peers can fetch the same source the host is playing.
func (m *Manager) ServeCurrentUrlStream() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		m.playbackMu.Lock()
		stream, ok := m.currentStream.Get()
		m.playbackMu.Unlock()

		if !ok || stream.Type() != nativeplayer.StreamTypeURL {
			http.Error(w, "no active url stream", http.StatusNotFound)
			return
		}

		stream.GetStreamHandler().ServeHTTP(w, r)
	})
}

// ServeEchoAttachments serves the attachments loaded into memory from the current stream.
func (m *Manager) ServeEchoAttachments(c echo.Context) error {
	// Get the current stream
	stream, ok := m.currentStream.Get()
	if !ok {
		return errors.New("no stream")
	}

	filename := c.Param("*")

	filename, _ = url.PathUnescape(filename)

	// Get the attachment
	attachment, ok := stream.GetAttachmentByName(filename)
	if !ok {
		return errors.New("attachment not found")
	}

	return c.Blob(200, attachment.Mimetype, attachment.Data)
}
