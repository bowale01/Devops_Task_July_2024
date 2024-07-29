package main

import (
	"context"
	"encoding/json"
	"log"
	"math"
	"math/rand"
	"net/http"
	"slices"
	"sync"
	"time"

	"github.com/docker/docker/pkg/namesgenerator"
)

type rider struct {
	id        int64
	name      string
	entrance  string
	vipStatus bool
}
type rollercoaster struct {
	nextId      int64
	idMu        sync.Mutex
	rideQueue   []*rider
	rideQueueMu sync.Mutex
	ride        []*rider
	rideMu      sync.Mutex
}

const (
	rideDuration = 5000
	numberOfCars = 8
	carCapacity  = 2
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	rc := rollercoaster{
		ride: make([]*rider, numberOfCars*carCapacity),
	}
	rc.start(ctx)
	err := http.ListenAndServe(":3000", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		d := &struct {
			Entrance string `json:"entrance"`
		}{}
		json.NewDecoder(r.Body).Decode(d)
		isVIP := false
		if rand.Float64() >= 0.7 {
			isVIP = true
		}

		rc.idMu.Lock()
		rider := &rider{id: rc.nextId, entrance: d.Entrance, name: namesgenerator.GetRandomName(0), vipStatus: isVIP}
		rc.nextId = rc.nextId + 1
		rc.idMu.Unlock()
		if rider.vipStatus {
			rc.rideQueue = slices.Insert(rc.rideQueue, 0, rider)
		} else {
			rc.rideQueue = append(rc.rideQueue, rider)
		}
		log.Printf("Entrace %s: %s entered the queue. Size: %d\n", d.Entrance, rider.name, len(rc.rideQueue))
	}))
	if err != nil {
		panic(err)
	}
}

func (rc *rollercoaster) start(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		default:
			rc.rideMu.Lock()
			i := 0
			for ; i <= len(rc.rideQueue)-1; i++ {
				if i == carCapacity*numberOfCars {
					break
				}
				rc.rideQueueMu.Lock()
				r := rc.rideQueue[0]
				car := int(math.Abs(float64((i)/2))) + 1
				carSeat := i % carCapacity
				rc.seatRider(r, car, carSeat)
				rc.rideQueue = rc.rideQueue[1:]
				rc.rideQueueMu.Unlock()
			}
			log.Println("Ride: started")
			time.Sleep(10000 * time.Millisecond)
			log.Println("Ride: finished")
			rc.rideMu.Unlock()
		}
	}
}

func (rc *rollercoaster) seatRider(r *rider, car, carSeat int) {
	log.Printf("Ride: %s entering the car %d in seat %d\n", r.name, car, carSeat)
	time.Sleep(400 * time.Millisecond)
	rc.ride = append(rc.ride, r)
	log.Printf("Ride: %s entered the car %d in seat %d\n", r.name, car, carSeat)
}
