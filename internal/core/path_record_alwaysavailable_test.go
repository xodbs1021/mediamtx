package core

import (
	"net/http"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/bluenviron/gortsplib/v5"
	"github.com/bluenviron/gortsplib/v5/pkg/description"
	"github.com/pion/rtp"
	"github.com/stretchr/testify/require"

	"github.com/bluenviron/mediamtx/internal/test"
)

func writeH264TestPacket(c *gortsplib.Client, medi *description.Media, seq uint16, ts uint32) error {
	return c.WritePacketRTP(medi, &rtp.Packet{
		Header: rtp.Header{
			Version:        2,
			Marker:         true,
			PayloadType:    96,
			SequenceNumber: seq,
			Timestamp:      ts,
			SSRC:           563423,
		},
		Payload: []byte{5},
	})
}

func waitForFileCount(t *testing.T, dir string, count int, timeout time.Duration) []os.DirEntry {
	deadline := time.Now().Add(timeout)
	for {
		files, err := os.ReadDir(dir)
		if err == nil && len(files) >= count {
			return files
		}
		if time.Now().After(deadline) {
			t.Fatalf("timed out waiting for %d file(s) in %s", count, dir)
		}
		time.Sleep(50 * time.Millisecond)
	}
}

func dirCountAndSize(t *testing.T, dir string) (int, int64) {
	files, err := os.ReadDir(dir)
	require.NoError(t, err)
	var size int64
	for _, f := range files {
		fi, err2 := os.Stat(filepath.Join(dir, f.Name()))
		require.NoError(t, err2)
		size += fi.Size()
	}
	return len(files), size
}

func TestPathRecordAlwaysAvailableRecorded(t *testing.T) {
	dir := t.TempDir()

	p, ok := newInstance(t,
		"record: yes\n"+
			"recordPath: "+filepath.Join(dir, "%path/%Y-%m-%d_%H-%M-%S-%f")+"\n"+
			"paths:\n"+
			"  mystream:\n"+
			"    alwaysAvailable: yes\n"+
			"    alwaysAvailableRecorded: false\n"+
			"    alwaysAvailableTracks:\n"+
			"      - codec: H264\n")
	require.Equal(t, true, ok)
	defer p.Close()

	// while offline, the offline substream must not be recorded
	time.Sleep(1 * time.Second)
	_, err := os.ReadDir(filepath.Join(dir, "mystream"))
	require.True(t, os.IsNotExist(err))

	// while a publisher is connected, recording must run
	media0 := test.UniqueMediaH264()

	source := gortsplib.Client{}
	err = source.StartRecording(
		"rtsp://localhost:8554/mystream",
		&description.Session{Medias: []*description.Media{media0}})
	require.NoError(t, err)

	for i := range 4 {
		err = writeH264TestPacket(&source, media0, 1123+uint16(i), 45343+90000*uint32(i))
		require.NoError(t, err)
	}

	waitForFileCount(t, filepath.Join(dir, "mystream"), 1, 3*time.Second)

	// after the publisher disconnects, recording must stop:
	// no new files may appear and the finalized file must not keep growing
	source.Close()
	time.Sleep(1 * time.Second)

	n1, size1 := dirCountAndSize(t, filepath.Join(dir, "mystream"))
	require.Equal(t, 1, n1)

	time.Sleep(1 * time.Second)

	n2, size2 := dirCountAndSize(t, filepath.Join(dir, "mystream"))
	require.Equal(t, 1, n2)
	require.Equal(t, size1, size2)
}

func TestPathRecordAlwaysAvailableRecordedStaticSource(t *testing.T) {
	dir := t.TempDir()

	p, ok := newInstance(t,
		"recordPath: "+filepath.Join(dir, "%path/%Y-%m-%d_%H-%M-%S-%f")+"\n"+
			"paths:\n"+
			"  feeder:\n"+
			"  mystream:\n"+
			"    source: rtsp://localhost:8554/feeder\n"+
			"    record: yes\n"+
			"    alwaysAvailable: yes\n"+
			"    alwaysAvailableRecorded: false\n"+
			"    alwaysAvailableTracks:\n"+
			"      - codec: H264\n")
	require.Equal(t, true, ok)
	defer p.Close()

	// keep a publisher on the feeder path so the static source can pull it
	media0 := test.UniqueMediaH264()

	feeder := gortsplib.Client{}
	err := feeder.StartRecording(
		"rtsp://localhost:8554/feeder",
		&description.Session{Medias: []*description.Media{media0}})
	require.NoError(t, err)

	go func() {
		seq := uint16(1123)
		ts := uint32(45343)
		for {
			if err2 := writeH264TestPacket(&feeder, media0, seq, ts); err2 != nil {
				return
			}
			seq++
			ts += 9000
			time.Sleep(100 * time.Millisecond)
		}
	}()

	// once the static source connects (first retry cycle), recording must start
	waitForFileCount(t, filepath.Join(dir, "mystream"), 1, 10*time.Second)

	// when the source disconnects, recording must stop while the offline segment plays
	feeder.Close()
	time.Sleep(1500 * time.Millisecond)

	n1, size1 := dirCountAndSize(t, filepath.Join(dir, "mystream"))

	time.Sleep(1500 * time.Millisecond)

	n2, size2 := dirCountAndSize(t, filepath.Join(dir, "mystream"))
	require.Equal(t, n1, n2)
	require.Equal(t, size1, size2)
}

func TestPathRecordAlwaysAvailableRecordedReloadWhileOffline(t *testing.T) {
	dir := t.TempDir()

	p, ok := newInstance(t,
		"api: yes\n"+
			"recordPath: "+filepath.Join(dir, "%path/%Y-%m-%d_%H-%M-%S-%f")+"\n"+
			"paths:\n"+
			"  feeder:\n"+
			"  mystream:\n"+
			"    source: rtsp://localhost:8554/feeder\n"+
			"    record: yes\n"+
			"    alwaysAvailable: yes\n"+
			"    alwaysAvailableRecorded: false\n"+
			"    alwaysAvailableTracks:\n"+
			"      - codec: H264\n")
	require.Equal(t, true, ok)
	defer p.Close()

	time.Sleep(500 * time.Millisecond)

	// a reload while the source is offline must not start recording the offline substream
	tr := &http.Transport{}
	defer tr.CloseIdleConnections()
	hc := &http.Client{Transport: tr}

	httpRequest(t, hc, http.MethodPatch, "http://localhost:9997/v3/config/paths/patch/mystream", map[string]any{
		"recordDeleteAfter": "48h",
	}, nil)

	time.Sleep(2 * time.Second)

	_, err := os.ReadDir(filepath.Join(dir, "mystream"))
	require.True(t, os.IsNotExist(err))
}
