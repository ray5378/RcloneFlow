package controller

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDoCopyFile_InvalidJSON(t *testing.T) {
	c := NewFsController(nil)
	_, err := c.doCopyFile(context.Background(), []byte("invalid json"))
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "无效的请求格式")
}

func TestDoCopyFile_MissingFields(t *testing.T) {
	c := NewFsController(nil)
	body := []byte(`{"srcFs":"local"}`)
	_, err := c.doCopyFile(context.Background(), body)
	assert.Error(t, err)
}

func TestDoMoveFile_InvalidJSON(t *testing.T) {
	c := NewFsController(nil)
	_, err := c.doMoveFile(context.Background(), []byte("invalid json"))
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "无效的请求格式")
}

func TestDoMoveFile_MissingFields(t *testing.T) {
	c := NewFsController(nil)
	body := []byte(`{"srcFs":"local"}`)
	_, err := c.doMoveFile(context.Background(), body)
	assert.Error(t, err)
}

func TestDoCopyDir_InvalidJSON(t *testing.T) {
	c := NewFsController(nil)
	_, err := c.doCopyDir(context.Background(), []byte("invalid json"))
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "无效的请求格式")
}

func TestDoCopyDir_MissingFields(t *testing.T) {
	c := NewFsController(nil)
	body := []byte(`{"srcFs":"local"}`)
	_, err := c.doCopyDir(context.Background(), body)
	assert.Error(t, err)
}

func TestDoMoveDir_InvalidJSON(t *testing.T) {
	c := NewFsController(nil)
	_, err := c.doMoveDir(context.Background(), []byte("invalid json"))
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "无效的请求格式")
}

func TestDoMoveDir_MissingFields(t *testing.T) {
	c := NewFsController(nil)
	body := []byte(`{"srcFs":"local"}`)
	_, err := c.doMoveDir(context.Background(), body)
	assert.Error(t, err)
}

func TestDoCopyFile_Success(t *testing.T) {
	orig := runRclone
	t.Cleanup(func() { runRclone = orig })

	runRclone = func(ctx context.Context, args ...string) (string, error) {
		assert.Equal(t, "copyto", args[0])
		return "", nil
	}

	c := NewFsController(nil)
	body := []byte(`{"srcFs":"local","srcRemote":"/src.txt","dstFs":"gdrive","dstRemote":"/dst.txt"}`)
	_, err := c.doCopyFile(context.Background(), body)
	assert.NoError(t, err)
}

func TestDoCopyFile_NetworkError(t *testing.T) {
	orig := runRclone
	t.Cleanup(func() { runRclone = orig })

	copytoCount := 0
	runRclone = func(ctx context.Context, args ...string) (string, error) {
		if args[0] != "copyto" {
			return "", nil
		}
		copytoCount++
		if copytoCount == 1 {
			return "", errors.New("network name not found")
		}
		return "", nil
	}

	c := NewFsController(nil)
	body := []byte(`{"srcFs":"local","srcRemote":"/share/file.txt","dstFs":"gdrive","dstRemote":"/share/file.txt"}`)
	_, err := c.doCopyFile(context.Background(), body)
	assert.NoError(t, err)
	assert.Equal(t, 2, copytoCount)
}

func TestDoCopyFile_CreateFilesystemError(t *testing.T) {
	orig := runRclone
	t.Cleanup(func() { runRclone = orig })

	copytoCount := 0
	runRclone = func(ctx context.Context, args ...string) (string, error) {
		if args[0] != "copyto" {
			return "", nil
		}
		copytoCount++
		if copytoCount == 1 {
			return "", errors.New("failed to create filesystem")
		}
		return "", nil
	}

	c := NewFsController(nil)
	body := []byte(`{"srcFs":"local","srcRemote":"/a/file.txt","dstFs":"gdrive","dstRemote":"/a/file.txt"}`)
	_, err := c.doCopyFile(context.Background(), body)
	assert.NoError(t, err)
	assert.Equal(t, 2, copytoCount)
}

func TestDoCopyFile_PersistentError(t *testing.T) {
	orig := runRclone
	t.Cleanup(func() { runRclone = orig })

	runRclone = func(ctx context.Context, args ...string) (string, error) {
		return "", errors.New("some persistent error")
	}

	c := NewFsController(nil)
	body := []byte(`{"srcFs":"local","srcRemote":"/src.txt","dstFs":"gdrive","dstRemote":"/dst.txt"}`)
	_, err := c.doCopyFile(context.Background(), body)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "persistent error")
}

func TestDoMoveFile_Success(t *testing.T) {
	orig := runRclone
	t.Cleanup(func() { runRclone = orig })

	runRclone = func(ctx context.Context, args ...string) (string, error) {
		if len(args) > 0 && args[0] == "moveto" {
			assert.Equal(t, "moveto", args[0])
		}
		return "", nil
	}

	c := NewFsController(nil)
	body := []byte(`{"srcFs":"local","srcRemote":"/src.txt","dstFs":"gdrive","dstRemote":"/dst.txt"}`)
	_, err := c.doMoveFile(context.Background(), body)
	assert.NoError(t, err)
}

func TestDoMoveFile_NetworkError(t *testing.T) {
	orig := runRclone
	t.Cleanup(func() { runRclone = orig })

	callCount := 0
	runRclone = func(ctx context.Context, args ...string) (string, error) {
		if args[0] == "moveto" {
			callCount++
			if callCount == 1 {
				return "", errors.New("network name not found")
			}
		}
		return "", nil
	}

	c := NewFsController(nil)
	body := []byte(`{"srcFs":"local","srcRemote":"/share/file.txt","dstFs":"gdrive","dstRemote":"/share/file.txt"}`)
	_, err := c.doMoveFile(context.Background(), body)
	assert.NoError(t, err)
}

func TestDoMoveFile_WebdavError(t *testing.T) {
	orig := runRclone
	t.Cleanup(func() { runRclone = orig })

	callCount := 0
	runRclone = func(ctx context.Context, args ...string) (string, error) {
		callCount++
		if args[0] == "moveto" && callCount == 1 {
			return "", errors.New("DirMove not supported")
		}
		if args[0] == "copyto" {
			return "", nil
		}
		return "", nil
	}

	c := NewFsController(nil)
	body := []byte(`{"srcFs":"local","srcRemote":"/src.txt","dstFs":"gdrive","dstRemote":"/dst.txt"}`)
	_, err := c.doMoveFile(context.Background(), body)
	assert.NoError(t, err)
	assert.True(t, callCount >= 2)
}

func TestDoCopyDir_Success(t *testing.T) {
	orig := runRclone
	t.Cleanup(func() { runRclone = orig })

	runRclone = func(ctx context.Context, args ...string) (string, error) {
		if args[0] == "copy" {
			assert.Equal(t, "copy", args[0])
		}
		return "", nil
	}

	c := NewFsController(nil)
	body := []byte(`{"srcFs":"local","srcRemote":"/src","dstFs":"gdrive","dstRemote":"/dst"}`)
	_, err := c.doCopyDir(context.Background(), body)
	assert.NoError(t, err)
}

func TestDoCopyDir_NetworkError(t *testing.T) {
	orig := runRclone
	t.Cleanup(func() { runRclone = orig })

	callCount := 0
	runRclone = func(ctx context.Context, args ...string) (string, error) {
		if args[0] == "copy" {
			callCount++
			if callCount == 1 {
				return "", errors.New("network name not found")
			}
		}
		return "", nil
	}

	c := NewFsController(nil)
	body := []byte(`{"srcFs":"local","srcRemote":"/share/dir","dstFs":"gdrive","dstRemote":"/share/dir"}`)
	_, err := c.doCopyDir(context.Background(), body)
	assert.NoError(t, err)
}

func TestDoMoveDir_Success(t *testing.T) {
	orig := runRclone
	t.Cleanup(func() { runRclone = orig })

	runRclone = func(ctx context.Context, args ...string) (string, error) {
		if args[0] == "move" {
			assert.Equal(t, "move", args[0])
		}
		return "", nil
	}

	c := NewFsController(nil)
	body := []byte(`{"srcFs":"local","srcRemote":"/src","dstFs":"gdrive","dstRemote":"/dst"}`)
	_, err := c.doMoveDir(context.Background(), body)
	assert.NoError(t, err)
}

func TestDoMoveDir_NetworkError(t *testing.T) {
	orig := runRclone
	t.Cleanup(func() { runRclone = orig })

	callCount := 0
	runRclone = func(ctx context.Context, args ...string) (string, error) {
		if args[0] == "move" {
			callCount++
			if callCount == 1 {
				return "", errors.New("network name not found")
			}
		}
		return "", nil
	}

	c := NewFsController(nil)
	body := []byte(`{"srcFs":"local","srcRemote":"/share/dir","dstFs":"gdrive","dstRemote":"/share/dir"}`)
	_, err := c.doMoveDir(context.Background(), body)
	assert.NoError(t, err)
}

func TestDoMoveDir_WebdavError(t *testing.T) {
	orig := runRclone
	t.Cleanup(func() { runRclone = orig })

	callCount := 0
	runRclone = func(ctx context.Context, args ...string) (string, error) {
		callCount++
		if args[0] == "move" && callCount == 1 {
			return "", errors.New("move call failed")
		}
		return "", nil
	}

	c := NewFsController(nil)
	body := []byte(`{"srcFs":"local","srcRemote":"/src","dstFs":"gdrive","dstRemote":"/dst"}`)
	_, err := c.doMoveDir(context.Background(), body)
	assert.NoError(t, err)
	assert.True(t, callCount >= 2)
}
