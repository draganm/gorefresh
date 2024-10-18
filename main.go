package main

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"os/exec"
	"os/signal"
	"syscall"
	"time"

	"github.com/draganm/gorefresh/build"
	"github.com/draganm/gorefresh/depdirs"
	"github.com/draganm/gosha/gosha"
	"github.com/fsnotify/fsnotify"
	"github.com/samber/lo"
	"github.com/urfave/cli/v2"
	"golang.org/x/sync/errgroup"
)

func main() {

	app := &cli.App{

		Flags: []cli.Flag{},

		Action: func(c *cli.Context) error {

			if c.Args().Len() == 0 {
				return fmt.Errorf("no module directory provided")
			}

			moduleDir := c.Args().First()

			st, err := os.Stat(moduleDir)
			if err != nil {
				return fmt.Errorf("failed to stat module directory: %w", err)
			}

			if !st.IsDir() {
				return fmt.Errorf("module directory is not a directory")
			}

			ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
			defer stop()

			eg, ctx := errgroup.WithContext(ctx)

			shaChan := make(chan []byte, 20)

			eg.Go(func() error {

				defer close(shaChan)

				lastSha, err := gosha.CalculatePackageSHA(moduleDir, false, false)
				if err != nil {
					return fmt.Errorf("failed to calculate package sha: %w", err)
				}

				shaChan <- lastSha

				watchedDirs := []string{}

				w, err := fsnotify.NewWatcher()
				if err != nil {
					return fmt.Errorf("failed to create watcher: %w", err)
				}

				defer w.Close()

				updateWatches := func() error {
					depDirs, err := depdirs.DependencyDirs(moduleDir)
					if err != nil {
						return fmt.Errorf("failed to get dependency directories: %w", err)
					}

					toAdd := lo.Without(depDirs, watchedDirs...)
					toRemove := lo.Without(watchedDirs, depDirs...)

					for _, d := range toRemove {

						_, err = os.Stat(d)

						switch {
						case os.IsNotExist(err):
							continue
						case err != nil:
							return fmt.Errorf("failed to stat %s: %w", d, err)
						}

						err = w.Remove(d)
						if err != nil {
							return fmt.Errorf("failed to remove %s from watcher: %w", d, err)
						}
					}

					for _, d := range toAdd {
						err = w.Add(d)
						if err != nil {
							return fmt.Errorf("failed to add %s to watcher: %w", d, err)
						}
					}

					watchedDirs = depDirs

					fmt.Println("updated watched dirs:", "watching", len(watchedDirs), "added", len(toAdd), "removed", len(toRemove))

					return nil

				}

				err = updateWatches()
				if err != nil {
					fmt.Println("failed to update watches:", err)
				}

				for ctx.Err() == nil {

					_, err = readLast(ctx, w.Events)
					if err != nil {
						return err
					}

					sha, err := gosha.CalculatePackageSHA(moduleDir, false, false)
					if err != nil {
						return fmt.Errorf("failed to calculate package sha: %w", err)
					}

					err = updateWatches()
					if err != nil {
						fmt.Println("failed to update watches:", err)
						continue
					}

					if bytes.Equal(sha, lastSha) {
						continue
					}

					lastSha = sha
					shaChan <- sha

				}
				return ctx.Err()

			})

			type builtBinary struct {
				binary string
				err    error
			}

			binaryChan := make(chan builtBinary, 20)

			eg.Go(func() error {
				for {
					_, err := readLast(ctx, shaChan)
					if err != nil {
						return fmt.Errorf("failed to read next sha: %w", err)
					}

					binary, err := build.BuildBinary(ctx, moduleDir)
					if err != nil {
						binaryChan <- builtBinary{err: err}
						continue
					}

					binaryChan <- builtBinary{binary: binary}
				}

			})

			eg.Go(func() error {

				procArgs := c.Args().Slice()[1:]

				startProcess := func(binary builtBinary) func() {
					fmt.Println()
					fmt.Println("---- starting ----")
					fmt.Println()

					if binary.err != nil {
						fmt.Println("could not build binary", binary.err)
						return func() {}
					}

					startTime := time.Now()

					fmt.Println("Started:", startTime.Format(time.DateTime))
					fmt.Println()
					procContext, procCancel := context.WithCancel(ctx)

					cmd := exec.CommandContext(procContext, binary.binary, procArgs...)
					cmd.Stdout = os.Stdout
					cmd.Stderr = os.Stderr

					err = cmd.Start()
					if err != nil {
						procCancel()
						fmt.Printf("Start failed: %v\n", err)
						return func() {}
					}

					procDoneChan := make(chan struct{})

					go func() {
						defer close(procDoneChan)
						defer os.Remove(binary.binary)

						err = cmd.Wait()
						fmt.Println()
						if err != nil {
							fmt.Printf("Failed after %.2f seconds: %v\n", time.Since(startTime).Seconds(), err)
							return
						}
						fmt.Printf("Terminated after %.2f seconds\n", time.Since(startTime).Seconds())
					}()

					return func() {
						procCancel()
						<-procDoneChan
					}

				}

				procCancel := func() {}

				for {
					binary, err := readLast(ctx, binaryChan)
					if err != nil {
						procCancel()
						return fmt.Errorf("failed to get next binary: %w", err)
					}

					procCancel()

					procCancel = startProcess(binary)
				}
			})

			return eg.Wait()

		},
	}
	app.RunAndExitOnError()
}

var errChannelClosed = fmt.Errorf("channel closed")

func readLast[T any](ctx context.Context, c chan T) (v T, err error) {
	select {
	case <-ctx.Done():
		return v, ctx.Err()
	case v, ok := <-c:
		if !ok {
			return v, errChannelClosed
		}
		for {
			select {
			case v, ok = <-c:
				if !ok {
					return v, errChannelClosed
				}
			default:
				return v, nil
			}
		}
	}
}
