package main

import (
	"os"
	"path/filepath"
	"testing"
)

// newTestJob builds a job with two files of two parts each: no network is
// needed to exercise the state machine of the files.
func newTestJob(t *testing.T) *Job {
	t.Helper()
	dir := t.TempDir()
	j := &Job{
		ID:       "job-1",
		State:    "running",
		PartSize: 1024,
		Total:    4096,
		dir:      dir,
		// No worker: the tests exercise the state machine, they never
		// touch the network.
		workers: 0,
	}
	for i, name := range []string{"a.bin", "b.bin"} {
		f := &File{
			ID:       "file-" + string(rune('1'+i)),
			URL:      "http://127.0.0.1:8090/files/" + name,
			Name:     name,
			Priority: prioNormal,
			Total:    2048,
			State:    "pending",
		}
		for p := 0; p < 2; p++ {
			start := int64(p) * 1024
			part := &Part{
				Index: len(j.parts) + 1,
				Start: start,
				End:   start + 1023,
				Size:  1024,
				State: "pending",
				file:  f,
			}
			f.parts = append(f.parts, part)
			f.pending++
			j.parts = append(j.parts, part)
		}
		j.files = append(j.files, f)
		j.PartCnt += f.PartCnt
	}
	j.Total = 4096
	j.PartCnt = 4
	return j
}

func TestFilePauseKeepsTheBytesAndTheQueue(t *testing.T) {
	j := newTestJob(t)
	f := j.files[0]

	j.pauseFile(f)
	if f.State != "paused" || !f.paused {
		t.Fatalf("the file should be paused: %+v", f)
	}
	// The scheduler does not pick the parts of a paused file.
	if p := j.nextPending(); p != nil && p.file == f {
		t.Fatalf("the scheduler picked a part of a paused file: %+v", p)
	}
	// The other file keeps going.
	p := j.nextPending()
	if p == nil || p.file != j.files[1] {
		t.Fatalf("the scheduler should keep the other file going: %+v", p)
	}

	j.resumeFile(f)
	if f.State != "pending" || f.paused {
		t.Fatalf("the file should be back in the queue: %+v", f)
	}
}

func TestFileStopKeepsTheFileAndUnlocksTheURL(t *testing.T) {
	j := newTestJob(t)
	f := j.files[0]
	f.State = "running"
	f.Received = 512

	j.cancelFile(f)
	if f.State != "stopped" {
		t.Fatalf("the file should be stopped: %+v", f)
	}
	// A stopped file stays in the download: the operator can change its URL
	// and start it again.
	if len(j.files) != 2 {
		t.Fatalf("the stopped file left the download: %d files", len(j.files))
	}
	if p := j.nextPending(); p != nil && p.file == f {
		t.Fatal("the scheduler picked a part of a stopped file")
	}
}

func TestFileRestartDropsTheBytesAndQueuesTheFileAgain(t *testing.T) {
	j := newTestJob(t)
	f := j.files[0]
	f.State = "running"
	f.Received = 512
	f.Merged = filepath.Join(j.dir, "a.bin")
	f.parts[0].Received = 512
	f.parts[0].State = "done"

	if err := j.restartFile(f); err != nil {
		t.Fatalf("restart: %v", err)
	}
	if f.Received != 0 || f.Percent != 0 {
		t.Fatalf("the bytes of the file were not dropped: %+v", f)
	}
	if f.State != "pending" || f.paused {
		t.Fatalf("the file should be back in the queue: %+v", f)
	}
	if f.parts[0].Received != 0 || f.parts[0].State != "pending" {
		t.Fatalf("the part was not queued again: %+v", f.parts[0])
	}
	if f.pending != len(f.parts) {
		t.Fatalf("the pending counter is wrong: %d of %d", f.pending, len(f.parts))
	}
	if j.Received != 0 {
		t.Fatalf("the total of the job was not corrected: %d", j.Received)
	}
}

func TestRemoveFileDropsTheFileAndItsBytes(t *testing.T) {
	j := newTestJob(t)
	f := j.files[0]
	f.State = "running"
	f.Received = 1024

	// A part file on disk: the removal deletes it.
	path := j.partPath(f.parts[0].Index)
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	if err := os.WriteFile(path, []byte("bytes"), 0o644); err != nil {
		t.Fatalf("write: %v", err)
	}

	if err := j.removeFile(f); err != nil {
		t.Fatalf("remove: %v", err)
	}
	if len(j.files) != 1 || len(j.parts) != 2 {
		t.Fatalf("the file left the download: %d files, %d parts", len(j.files), len(j.parts))
	}
	if _, err := os.Stat(path); !os.IsNotExist(err) {
		t.Fatalf("the bytes of the removed file are still on disk: %v", err)
	}
	if j.Received != 0 || j.Total != 2048 {
		t.Fatalf("the totals of the job are wrong: %d of %d", j.Received, j.Total)
	}
}

func TestReplaceFileSwapsTheParts(t *testing.T) {
	j := newTestJob(t)
	old := j.files[0]
	old.State = "stopped"
	old.Received = 512

	fresh := &File{
		ID:       old.ID,
		URL:      "http://127.0.0.1:8090/files/new.bin",
		Name:     "new.bin",
		Priority: prioNormal,
		Total:    512,
		State:    "pending",
	}
	part := &Part{Index: 0, Start: 0, End: 511, Size: 512, State: "pending", file: fresh}
	fresh.parts = append(fresh.parts, part)

	j.replaceFile(old, fresh)
	if len(j.files) != 2 || j.files[0] != fresh {
		t.Fatalf("the file was not replaced: %+v", j.files)
	}
	if len(j.parts) != 3 {
		t.Fatalf("the parts of the old file are still in the queue: %d", len(j.parts))
	}
	if j.parts[2].Index != 3 {
		t.Fatalf("the new part has no number: %+v", j.parts[2])
	}
	if j.Total != 2560 {
		t.Fatalf("the total of the job is wrong: %d", j.Total)
	}
	if j.Received != 0 {
		t.Fatalf("the received bytes of the job are wrong: %d", j.Received)
	}
}

// A complete file can be stopped: its URL becomes editable again, so the
// operator can start it once more with another URL.
func TestStopOnACompleteFileUnlocksTheURL(t *testing.T) {
	j := newTestJob(t)
	f := j.files[0]
	f.State = "done"
	f.Received = 2048
	j.State = "done"

	j.cancelFile(f)
	if f.State != "stopped" {
		t.Fatalf("the complete file should be stopped: %+v", f)
	}
	if !f.paused {
		t.Fatalf("the stopped file should not be picked by the scheduler: %+v", f)
	}
	// The download is not complete any more: the operator can start the file
	// again.
	if j.State == "done" {
		t.Fatalf("the job is still complete: %+v", j.State)
	}
	// The file stays in the download: its URL can be changed.
	if len(j.files) != 2 {
		t.Fatalf("the stopped file left the download: %d files", len(j.files))
	}
}

// A restart drops the merged file of the old attempt: the link of the files
// table must not point to the bytes of a download that is not there any more.
func TestRestartRemovesTheMergedFileOfTheOldAttempt(t *testing.T) {
	j := newTestJob(t)
	f := j.files[0]
	f.State = "stopped"
	merged := filepath.Join(j.dir, "merged-a.bin")
	if err := os.WriteFile(merged, []byte("old bytes"), 0o644); err != nil {
		t.Fatalf("write: %v", err)
	}
	f.Merged = merged
	f.SHA = "abc"
	f.Verify = "size matches the server"

	if err := j.restartFile(f); err != nil {
		t.Fatalf("restart: %v", err)
	}
	if _, err := os.Stat(merged); !os.IsNotExist(err) {
		t.Fatalf("the merged file of the old attempt is still on disk: %v", err)
	}
	if f.Merged != "" || f.SHA != "" || f.Verify != "" {
		t.Fatalf("the file still points to the old attempt: %+v", f)
	}
	if f.State != "pending" {
		t.Fatalf("the file should be back in the queue: %+v", f)
	}
}
