package config

import (
  "github.com/pkg/errors"
  "github.com/fsnotify/fsnotify"

)

type watcher struct {
  emitter

  fsWatcher *fsnotify.Watcher
  close     chan struct{}
  closed    chan struct{}
}

func newWatcher(path string, callback func()) (*watcher, error) {
  fsWatcher, err := fsnotify.NewWatcher()
  if err != nil {
    return nil, errors.Wrapf(err, "Failed to create fs notify watcher for %s", path)
  }

  w := &watcher{
    fsWatcher:  fsWatcher,
    close:      make(chan struct{}),
    closed:     make(chan struct{}),
  }

  return w, nil
}

func (watcher *watcher) Close() error {
  close(watcher.close)
  <-watcher.closed

  return nil
}