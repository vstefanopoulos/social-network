package eventsub

// type EventBox[K comparable, P any] struct {
// 	pubTimeout time.Duration //if a channel is busy, how long to wait before timing out the message
// 	seed       maphash.Seed
// 	shardCount uint64
// 	shards     []*shard[K, P]
// 	buffer     int
// }

// func NewEventBox[K comparable, P any](shardCount int, pubTimeout time.Duration, channelBuffer int) EventBox[K, P] {
// 	if shardCount == 0 {
// 		panic("you can't have zero count of shards. they probably be should be quite a lot, like 100+")
// 	}

// 	s := []*shard[K, P]{}
// 	for range shardCount {
// 		s = append(s, newShard[K, P]())
// 	}

// 	return EventBox[K, P]{
// 		pubTimeout: pubTimeout,
// 		seed:       maphash.MakeSeed(),
// 		shardCount: uint64(len(s)),
// 		shards:     s,
// 		buffer:     channelBuffer,
// 	}
// }

// var count = 0

// // subscribe to a key, and receive payloads from the given channel. TODO the channel may close?
// func (box *EventBox[K, P]) Sub(key K) (<-chan P, func()) {
// 	channel := make(chan P, box.buffer)
// 	shard := box.getShard(key)

// 	shard.mu.Lock()
// 	chanSlice, ok := shard.event2Channel[key]
// 	if !ok {
// 		chanSlice = make([]chan P, 0)
// 		shard.event2Channel[key] = chanSlice
// 	}
// 	shard.event2Channel[key] = append(chanSlice, channel)
// 	count++
// 	shard.mu.Unlock()
// 	fmt.Println("Added to box, total: ", count)

// 	close := func() {
// 		fmt.Println("called close")
// 		shard.mu.Lock()
// 		count--
// 		slice := shard.event2Channel[key]
// 		if len(slice) == 1 {
// 			delete(shard.event2Channel, key)
// 			fmt.Println("removed from box, total", count)
// 			shard.mu.Unlock()
// 			return
// 		}
// 		index := slices.Index(slice, channel)
// 		slice = slices.Delete(slice, index, index)
// 		shard.event2Channel[key] = slice

// 		shard.mu.Unlock()
// 		fmt.Println("removed from box, total", count)
// 	}

// 	return channel, close
// }

// var ErrTimeout = errors.New("nobody ready to read message")

// func (box *EventBox[K, P]) Pub(key K, payload P) error {
// 	shard := box.getShard(key)
// 	shard.mu.RLock()
// 	chanSlice, ok := shard.event2Channel[key]
// 	if !ok {
// 		shard.mu.RUnlock()
// 		return nil
// 	}

// 	newSlice := make([]chan P, 0, 5)
// 	for _, c := range chanSlice {
// 		newSlice = append(newSlice, c)
// 	}
// 	shard.mu.RUnlock()
// 	timeouts := 0
// 	for _, c := range newSlice {

// 		c <- payload

// 	}

// 	if timeouts > 0 {
// 		return fmt.Errorf("timeouts occured: %d, %w", timeouts, ErrTimeout)
// 	}
// 	return nil
// }

// func (box *EventBox[K, P]) getShard(key K) *shard[K, P] {
// 	shardIndex := maphash.Comparable(box.seed, key) % uint64(len(box.shards))
// 	return box.shards[shardIndex]
// }

// type shard[K comparable, P any] struct {
// 	mu            sync.RWMutex
// 	event2Channel map[K][]chan P
// }

// func newShard[K comparable, P any]() *shard[K, P] {
// 	s := shard[K, P]{
// 		mu:            sync.RWMutex{},
// 		event2Channel: make(map[K][]chan P),
// 	}
// 	return &s
// }

// // PrintChannelCount prints the total number of channels in all shards
// func (box *EventBox[K, P]) PrintChannelCount() {
// 	total := 0
// 	for _, shard := range box.shards {
// 		shard.mu.RLock()
// 		for _, channels := range shard.event2Channel {
// 			total += len(channels)
// 		}
// 		shard.mu.RUnlock()
// 	}
// 	fmt.Printf("Total channels in all shards: %d\n", total)
// }

// func Bench() {
// 	box := NewEventBox[int, int](100, time.Second, 5)

// 	wg := sync.WaitGroup{}

// 	x := func() {
// 		key := rand.IntN(10000000000000)
// 		nums := []int{}
// 		c, close := box.Sub(key)
// 		for i := range 200 {
// 			nums = append(nums, i)
// 		}

// 		wg.Go(func() {
// 			for _, n := range nums {
// 				time.Sleep(time.Millisecond * 15)
// 				box.Pub(key, n)
// 			}
// 		})

// 		wg.Go(func() {
// 			defer close()
// 			n := 0
// 			for message := range c {
// 				time.Sleep(time.Millisecond * 15)
// 				if nums[n] != message {
// 					panic("failed test!")
// 				}
// 				if len(nums)-1 == n {
// 					return
// 				}
// 				n++
// 			}

// 		})

// 	}
// 	for range 10000 {
// 		time.Sleep(time.Millisecond * 1)
// 		x()
// 	}

// 	wg.Wait()
// 	box.PrintChannelCount()
// 	time.Sleep(time.Millisecond * 500)
// 	fmt.Println("fin")
// }
