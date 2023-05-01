package main

import (
	"fmt"
	"math/rand"
	"sync"
	"sync/atomic"
	"time"
)

type VehicleType int

const (
	Gas VehicleType = iota
	Diesel
	LPG
	Electric
)

type Vehicle struct {
	ID   int
	Type VehicleType
}

type Pump struct {
	Type          VehicleType
	Queue         chan Vehicle
	WaitTimeMin   time.Duration
	WaitTimeMax   time.Duration
	StationID     int
	TotalVehicles int
	TotalTime     time.Duration
}

type CashRegister struct {
	ID            int
	Queue         chan Vehicle
	WaitTimeMin   time.Duration
	WaitTimeMax   time.Duration
	TotalVehicles int
	TotalTime     time.Duration
}

// Vytvari nove vozidlo
func NewRandomVehicle(vehicleID int) Vehicle {
	vehicleType := rand.Intn(4) // Генерируем случайное число от 0 до 3 включительно
	return Vehicle{
		ID:   vehicleID,
		Type: VehicleType(vehicleType),
	}
}

// Vytvari stanice
func createPumps() []Pump {
	pumps := make([]Pump, 17)

	for i := 0; i < 4; i++ {
		pumps[i] = Pump{Type: Gas, Queue: make(chan Vehicle, 2), WaitTimeMin: 1 * time.Second, WaitTimeMax: 5 * time.Second, StationID: i + 1}
	}

	for i := 4; i < 8; i++ {
		pumps[i] = Pump{Type: Diesel, Queue: make(chan Vehicle, 2), WaitTimeMin: 1 * time.Second, WaitTimeMax: 5 * time.Second, StationID: i + 1}
	}

	pumps[8] = Pump{Type: LPG, Queue: make(chan Vehicle, 2), WaitTimeMin: 1 * time.Second, WaitTimeMax: 5 * time.Second, StationID: 9}

	for i := 9; i < 17; i++ {
		pumps[i] = Pump{Type: Electric, Queue: make(chan Vehicle, 2), WaitTimeMin: 3 * time.Second, WaitTimeMax: 10 * time.Second, StationID: i + 1}
	}

	return pumps
}

// Pokladny
func createCashRegisters() []CashRegister {
	cashRegisters := make([]CashRegister, 2)

	for i := 0; i < 2; i++ {
		cashRegisters[i] = CashRegister{ID: i + 1, Queue: make(chan Vehicle, 2), WaitTimeMin: 500 * time.Millisecond, WaitTimeMax: 2 * time.Second}
	}

	return cashRegisters
}

// Vyber volne pumpy
func selectAvailablePump(pumps []Pump, vehicleType VehicleType) *Pump {
	for i := range pumps {
		if pumps[i].Type == vehicleType && len(pumps[i].Queue) < cap(pumps[i].Queue) {
			return &pumps[i]
		}
	}
	return nil
}

// Vyber volne pokladny
func selectAvailableCashRegister(cashRegisters []CashRegister) *CashRegister {
	for i := range cashRegisters {
		if len(cashRegisters[i].Queue) < cap(cashRegisters[i].Queue) {
			return &cashRegisters[i]
		}
	}
	return nil
}

func randomDuration(min, max time.Duration) time.Duration {
	delta := max - min
	randomValue := time.Duration(rand.Int63n(int64(delta)))
	return min + randomValue
}

// Obsluhuje auto na stanici
func (p *Pump) ProcessVehicle(vehicle Vehicle) {
	waitTime := randomDuration(p.WaitTimeMin, p.WaitTimeMax)
	time.Sleep(waitTime)

	p.TotalVehicles++
	p.TotalTime += waitTime

	fmt.Printf("Vehicle %d of type %d is being processed at pump %d\n", vehicle.ID, vehicle.Type, p.StationID)
}

// Obsluhuje auto na pokladne
func (c *CashRegister) ProcessVehicle(vehicle Vehicle) {
	waitTime := randomDuration(c.WaitTimeMin, c.WaitTimeMax)
	time.Sleep(waitTime)

	c.TotalVehicles++
	c.TotalTime += waitTime

	fmt.Printf("Vehicle %d of type %d is being processed at cash register %d\n", vehicle.ID, vehicle.Type, c.ID)
}


// Obsluha auta
func serviceVehicle(vehicle Vehicle, pumps []Pump, cashRegisters []CashRegister, wg *sync.WaitGroup, skippedVehicles *int32) {
	defer wg.Done()

	pump := selectAvailablePump(pumps, vehicle.Type)

	if pump != nil {
		pump.Queue <- vehicle
		done := make(chan struct{})
		go func() {
			pump.ProcessVehicle(vehicle)
			<-pump.Queue
			done <- struct{}{}
		}()

		<-done

		for {
			cashRegister := selectAvailableCashRegister(cashRegisters)
			if cashRegister != nil {
				cashRegister.Queue <- vehicle
				go func() {
					cashRegister.ProcessVehicle(vehicle)
					<-cashRegister.Queue
					done <- struct{}{}
				}()

				<-done
				break
			} else {
				time.Sleep(100 * time.Millisecond)
			}
		}
	} else {
		fmt.Printf("Vehicle %d of type %d left without refueling because all pumps of this type are busy\n", vehicle.ID, vehicle.Type)
		atomic.AddInt32(skippedVehicles, 1)
	}
}


func runSimulation(duration time.Duration, pumps []Pump, cashRegisters []CashRegister) {
	start := time.Now()
	vehicleID := 1
	var skippedVehicles int32

	var wg sync.WaitGroup

	for {
		elapsed := time.Since(start)
		if elapsed >= duration {
			break
		}

		vehicle := NewRandomVehicle(vehicleID)
		vehicleID++

		wg.Add(1)
		go serviceVehicle(vehicle, pumps, cashRegisters, &wg, &skippedVehicles)

		randomDelay := time.Duration(rand.Intn(901)+100) * time.Millisecond
		time.Sleep(randomDelay)
	}

	wg.Wait()

	fmt.Println("Simulation completed. Statistics:")
	fmt.Printf("Skipped vehicles: %d\n", skippedVehicles)

	var totalPumpTime, totalPumpVehicles int
	for _, pump := range pumps {
		totalPumpTime += int(pump.TotalTime)
		totalPumpVehicles += pump.TotalVehicles
	}
	avgPumpTime := float64(totalPumpTime) / float64(totalPumpVehicles)
	fmt.Printf("Average refueling time: %.2f seconds\n", avgPumpTime/float64(time.Second))

	var totalCashRegisterTime, totalCashRegisterVehicles int
	for _, cashRegister := range cashRegisters {
		totalCashRegisterTime += int(cashRegister.TotalTime)
		totalCashRegisterVehicles += cashRegister.TotalVehicles
	}
	avgCashRegisterTime := float64(totalCashRegisterTime) / float64(totalCashRegisterVehicles)
	fmt.Printf("Average cash register waiting time: %.2f seconds\n", avgCashRegisterTime/float64(time.Second))
}

func main() {
	//vehicleID := 1
	//vehicle := NewRandomVehicle(vehicleID)
	//fmt.Printf("Vehicle %d has fuel type %d\n", vehicle.ID, vehicle.Type)
	//
	//pumps := createPumps()
	//// Выведем информацию о созданных заправочных станциях
	//for _, pump := range pumps {
	//	fmt.Printf("Pump ID: %d, Type: %d, WaitTimeMin: %v, WaitTimeMax: %v\n", pump.StationID, pump.Type, pump.WaitTimeMin, pump.WaitTimeMax)
	//}
	//cashRegisters := createCashRegisters()
	//
	//// Выведем информацию о созданных кассах
	//for _, cashRegister := range cashRegisters {
	//	fmt.Printf("Cash Register ID: %d\n", cashRegister.ID)
	//}
	//
	//pumps := createPumps()
	//cashRegisters := createCashRegisters()
	//
	//vehicleID := 1
	//vehicle := NewRandomVehicle(vehicleID)
	//
	//var wg sync.WaitGroup
	//wg.Add(1)
	//serviceVehicle(vehicle, pumps, cashRegisters, &wg)
	//wg.Wait()

	pumps := createPumps()
	cashRegisters := createCashRegisters()
	// Je mozne menit cas
	runSimulation(1*time.Minute, pumps, cashRegisters)
}
