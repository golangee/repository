package fs

import (
	"bytes"
	"context"
	"crypto/sha256"
	"github.com/golangee/repository/iter"
	"io"
	"math/rand"
	"strconv"
	"sync"
	"testing"
)

type testBlob struct {
	name string
	data []byte
}

func Test_blobRepoRaces(t *testing.T) {
	ctx := context.Background()
	repo := must(NewBlobRepository(Dir(t.TempDir())))
	must("", repo.DeleteAll(ctx))

	const (
		maxFiles    = 10
		concurrency = 1000
	)

	// insert racy stuff
	blobs := genTestBlobs(concurrency)
	var wg sync.WaitGroup
	for i := 0; i < concurrency; i++ {
		wg.Add(1)
		go func(n int) {
			defer wg.Done()

			name := "racy" + strconv.Itoa(concurrency%maxFiles)
			blob := blobs[i]
			w := must(repo.Write(ctx, name))
			must(w.Write(blob.data))
			must("", w.Close())
		}(i)
	}

	wg.Wait()
	repo.assertEmptyMutexes()

	// check consistency, we don't really know what we have written
	// but last 32byte must be the sha of the data
	ids := must(iter.Collect(must(repo.FindAll(ctx))))
	for _, id := range ids {
		r := must(repo.Read(ctx, id))
		buf := must(io.ReadAll(r))
		must("", r.Close())

		sum := sha256.Sum256(buf[:len(buf)-32])
		if !bytes.Equal(sum[:], buf[len(buf)-32:]) {
			t.Fatalf("checksum mismatch")
		}
	}
	repo.assertEmptyMutexes()

}

func Test_blobRepo(t *testing.T) {
	ctx := context.Background()
	repo := must(NewBlobRepository(Dir(t.TempDir())))
	must("", repo.DeleteAll(ctx))
	if n := must(repo.Count(ctx)); n != 0 {
		t.Fatalf("expected 0 bot got %v", n)
	}

	blobs := genTestBlobs(100)

	// insert
	repo.assertEmptyMutexes()
	for i, blob := range blobs {
		w := must(repo.Write(ctx, blob.name))
		must(w.Write(blob.data))
		must("", w.Close())

		if n := must(repo.Count(ctx)); n != int64(i)+1 {
			t.Fatalf("expected %v bot got %v", i+1, n)
		}
	}
	repo.assertEmptyMutexes()

	// read
	for _, blob := range blobs {
		r := must(repo.Read(ctx, blob.name))
		buf := must(io.ReadAll(r))
		must("", r.Close())

		if !bytes.Equal(buf, blob.data) {
			t.Fatalf("expected \n%v but got\n%v", blob.data, buf)
		}
	}
	repo.assertEmptyMutexes()

	// findAll
	ids := must(iter.Collect(must(repo.FindAll(ctx))))
	if len(ids) != len(blobs) {
		t.Fatalf("expected %v but got %v", len(blobs), len(ids))
	}

	for _, id := range ids {
		found := false
		for _, blob := range blobs {
			if blob.name == id {
				found = true
				break
			}
		}
		if !found {
			t.Fatalf("id %v not found", id)
		}
	}
	repo.assertEmptyMutexes()

	// delete
	for i, id := range ids {
		must("", repo.Delete(ctx, id))
		if n := must(repo.Count(ctx)); n+int64(i+1) != int64(len(blobs)) {
			t.Fatalf("count and delete mismatch")
		}

		if _, err := repo.Read(ctx, id); err == nil {
			t.Fatal("expected error but got nil")
		}
	}
	repo.assertEmptyMutexes()
}

func genTestBlobs(count int) []testBlob {
	var res []testBlob
	r := rand.New(rand.NewSource(123))
	for i := 0; i < count; i++ {
		buf := make([]byte, r.Int31n(32*1024))
		if _, err := r.Read(buf); err != nil {
			panic(err)
		}

		res = append(res, testBlob{
			name: "test-" + strconv.Itoa(i),
			data: buf,
		})
	}

	res = append(res, testBlob{
		name: "zero",
		data: []byte{},
	})

	// calc standalone checksums
	for _, re := range res {
		sum := sha256.Sum256(re.data)
		re.data = append(re.data, sum[:]...)
	}

	return res
}

func must[T any](t T, err error) T {
	if err != nil {
		panic(err)
	}

	return t
}

func Test_validName(t *testing.T) {

	tests := []struct {
		name string
		want bool
	}{
		{"", false},
		{"/", false},
		{".", false},
		{"a/b/c", true},
		{"a/b /c", false},
		{"a/b b/c", false},
		{"A/b/c", false},
		{"%20/b/c", false},
		{"/a", false},
		{"a", true},
		{"a/b/./c", false},
		{"a/b//c", false},
		{"a/b/../c", false},
		{"a/b/c/", false},
		{"myname.txt", true},
		{"my-context/my_ticket/file.txt", true},
		{"1/2/3", true},
		{"550e8400-e29b-11d4-a716-446655440000", true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ValidName(tt.name); got != tt.want {
				t.Errorf("validName() = %v, want %v", got, tt.want)
			}
		})
	}
}
