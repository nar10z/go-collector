package accumulator_example

import (
	"context"
	"testing"
	"time"

	acc "github.com/lrweck/accumulator"
	goaccum "github.com/nar10z/go-accumulator"
	"golang.org/x/sync/errgroup"
)

const (
	flushSize     = 10000
	flushInterval = time.Second
)

type Data struct {
	i int
}

func Benchmark_accum(b *testing.B) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()

	b.Run("#1.1 go-accumulator, channel", func(b *testing.B) {
		b.ResetTimer()
		summary := 0

		accumulator, _ := goaccum.New[*Data](flushSize, flushInterval, func(events []*Data) error {
			summary += len(events)
			time.Sleep(time.Microsecond)
			return nil
		})

		for i := 0; i < b.N; i++ {
			_ = accumulator.AddAsync(ctx, &Data{i: i})
		}

		accumulator.Stop()

		if summary != b.N {
			b.Fail()
		}
	})
	b.Run("#1.2 go-accumulator, list", func(b *testing.B) {
		b.ResetTimer()

		summary := 0

		accumulator, _ := goaccum.NewWithStorage[*Data](flushSize, flushInterval, func(events []*Data) error {
			summary += len(events)
			time.Sleep(time.Microsecond)
			return nil
		}, goaccum.List)

		for i := 0; i < b.N; i++ {
			_ = accumulator.AddAsync(ctx, &Data{i: i})
		}

		accumulator.Stop()

		if summary != b.N {
			b.Fail()
		}
	})
	b.Run("#1.3 go-accumulator, slice", func(b *testing.B) {
		b.ResetTimer()

		summary := 0

		accumulator, _ := goaccum.NewWithStorage[*Data](flushSize, flushInterval, func(events []*Data) error {
			summary += len(events)
			time.Sleep(time.Microsecond)
			return nil
		}, goaccum.Slice)

		for i := 0; i < b.N; i++ {
			_ = accumulator.AddAsync(ctx, &Data{i: i})
		}

		accumulator.Stop()

		if summary != b.N {
			b.Fail()
		}
	})
	b.Run("#1.4 go-accumulator, stdList", func(b *testing.B) {
		b.ResetTimer()

		summary := 0

		accumulator, _ := goaccum.NewWithStorage[*Data](flushSize, flushInterval, func(events []*Data) error {
			summary += len(events)
			time.Sleep(time.Microsecond)
			return nil
		}, goaccum.StdList)

		for i := 0; i < b.N; i++ {
			_ = accumulator.AddAsync(ctx, &Data{i: i})
		}

		accumulator.Stop()

		if summary != b.N {
			b.Fail()
		}
	})

	b.Run("#2.1 go-accumulator, channel sync", func(b *testing.B) {
		b.ResetTimer()

		summary := 0

		accumulator, _ := goaccum.New[*Data](flushSize, flushInterval, func(events []*Data) error {
			summary += len(events)
			time.Sleep(time.Microsecond)
			return nil
		})

		var errGr errgroup.Group
		errGr.SetLimit(flushSize)
		for i := 0; i < b.N; i++ {
			errGr.Go(func() error {
				return accumulator.AddSync(ctx, &Data{i: i})
			})
		}

		_ = errGr.Wait()
		accumulator.Stop()

		if summary != b.N {
			b.Fail()
		}
	})
	b.Run("#2.2 go-accumulator, list sync", func(b *testing.B) {
		b.ResetTimer()

		summary := 0

		accumulator, _ := goaccum.NewWithStorage[*Data](flushSize, flushInterval, func(events []*Data) error {
			summary += len(events)
			time.Sleep(time.Microsecond)
			return nil
		}, goaccum.List)

		var errGr errgroup.Group
		errGr.SetLimit(flushSize)
		for i := 0; i < b.N; i++ {
			errGr.Go(func() error {
				return accumulator.AddSync(ctx, &Data{i: i})
			})
		}

		_ = errGr.Wait()
		accumulator.Stop()

		if summary != b.N {
			b.Fail()
		}
	})
	b.Run("#2.3 go-accumulator, slice sync", func(b *testing.B) {
		b.ResetTimer()

		summary := 0

		accumulator, _ := goaccum.NewWithStorage[*Data](flushSize, flushInterval, func(events []*Data) error {
			summary += len(events)
			time.Sleep(time.Microsecond)
			return nil
		}, goaccum.Slice)

		var errGr errgroup.Group
		errGr.SetLimit(flushSize)
		for i := 0; i < b.N; i++ {
			errGr.Go(func() error {
				return accumulator.AddSync(ctx, &Data{i: i})
			})
		}

		_ = errGr.Wait()
		accumulator.Stop()

		if summary != b.N {
			b.Fail()
		}
	})
	b.Run("#2.4 go-accumulator, stdList sync", func(b *testing.B) {
		b.ResetTimer()

		summary := 0

		accumulator, _ := goaccum.NewWithStorage[*Data](flushSize, flushInterval, func(events []*Data) error {
			summary += len(events)
			time.Sleep(time.Microsecond)
			return nil
		}, goaccum.StdList)

		var errGr errgroup.Group
		errGr.SetLimit(flushSize)
		for i := 0; i < b.N; i++ {
			errGr.Go(func() error {
				return accumulator.AddSync(ctx, &Data{i: i})
			})
		}

		_ = errGr.Wait()
		accumulator.Stop()

		if summary != b.N {
			b.Fail()
		}
	})

	b.Run("#3. lrweck/accumulator", func(b *testing.B) {
		b.ResetTimer()

		summary := 0

		inputChan := make(chan *Data, flushSize)
		batch := acc.New(inputChan, flushSize, flushInterval)

		go func() {
			for i := 0; i < b.N; i++ {
				inputChan <- &Data{i: i}
			}
			close(inputChan)
		}()

		_ = batch.Accumulate(ctx, func(o acc.CallOrigin, items []*Data) {
			summary += len(items)
			time.Sleep(time.Microsecond)
		})

		if summary != b.N {
			b.Fail()
		}
	})
}
