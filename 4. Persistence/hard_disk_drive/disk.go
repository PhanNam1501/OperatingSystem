package main

import (
	"flag"
	"fmt"
	"math"
	"math/rand"
	"os"
	"strconv"
	"strings"
)

const MAXTRACKS = 1000

const (
	STATENull   = 0
	STATESeek   = 1
	STATERotate = 2
	STATEXfer   = 3
	STATEDone   = 4
)

func mod(a, b float64) float64 {
	m := math.Mod(a, b)
	if m < 0 {
		m += b
	}
	return m
}

type Request struct {
	block int
	index int
}

type BlockInfo struct {
	track int
	angle float64
	name  int
}

type Disk struct {
	addr          string
	addrDesc      string
	lateAddr      string
	lateAddrDesc  string
	policy        string
	seekSpeedBase float64
	rotateSpeed   float64
	skew          int
	window        int
	compute       bool
	zoning        string
	rng           *rand.Rand

	blockInfoList    []BlockInfo
	blockToTrackMap  map[int]int
	blockToAngleMap  map[int]float64
	tracksBeginEnd   map[int][2]int
	blockAngleOffset []int
	maxBlock         int

	requests     []int
	lateRequests []int
	lateCount    int

	trackWidth int
	tracks     map[int]float64

	armTrack    float64
	armSpeed    float64
	armX1       float64
	armTargetX1 float64
	armTarget   float64

	requestQueue []Request
	requestState []int
	requestCount int
	currWindow   int
	fairWindow   int

	currentIndex int
	currentBlock int
	state        int

	angle float64
	timer int

	seekBegin int
	rotBegin  int
	xferBegin int

	seekTotal float64
	rotTotal  float64
	xferTotal float64

	isDone bool
}

func NewDisk(addr, addrDesc, lateAddr, lateAddrDesc, policy string, seedSpeed, rotateSpeed float64, skew, window int, compute bool, zoning string, seed int) *Disk {
	d := &Disk{
		addr:            addr,
		addrDesc:        addrDesc,
		lateAddr:        lateAddr,
		lateAddrDesc:    lateAddrDesc,
		policy:          policy,
		seekSpeedBase:   seedSpeed,
		rotateSpeed:     rotateSpeed,
		skew:            skew,
		window:          window,
		compute:         compute,
		zoning:          zoning,
		rng:             rand.New(rand.NewSource(int64(seed))),
		blockToTrackMap: make(map[int]int),
		blockToAngleMap: make(map[int]float64),
		tracksBeginEnd:  make(map[int][2]int),
		tracks:          make(map[int]float64),
		currentIndex:    -1,
		currentBlock:    -1,
		state:           STATENull,
	}

	d.InitBlockLayout()
	d.requests = d.MakeRequests(d.addr, d.addrDesc)
	d.lateRequests = d.MakeRequests(d.lateAddr, d.lateAddrDesc)

	if d.policy == "BSATF" && d.window != -1 {
		d.fairWindow = d.window
	} else {
		d.fairWindow = -1
	}

	fmt.Printf("REQUESTS %v\n\n", d.requests)

	if len(d.lateRequests) > 0 {
		fmt.Printf("LATE REQUESTS %v\n\n", d.lateRequests)
	}

	if !d.compute {
		fmt.Println("\nFor the requests above, compute the seek, rotate, and transfer times.")
		fmt.Println("Use -c to see the answers.\n")
	}

	d.trackWidth = 40
	d.tracks[0] = 140
	d.tracks[1] = d.tracks[0] - float64(d.trackWidth)
	d.tracks[2] = d.tracks[1] - float64(d.trackWidth)

	if d.seekSpeedBase > 1 && d.trackWidth%int(d.seekSpeedBase) != 0 {
		fmt.Printf("Seek speed (%d) must divide evenly into track width (%d)\n", int(d.seekSpeedBase), d.trackWidth)
		os.Exit(1)
	}

	d.armTrack = 0
	d.armSpeed = float64(seedSpeed)
	distFromSpindle := d.tracks[int(d.armTrack)]
	d.armX1 = -(distFromSpindle * math.Cos(0)) - 20

	for i, req := range d.requests {
		d.AddQueueEntry(req, i)
	}

	d.currWindow = d.window

	return d
}

func (d *Disk) InitBlockLayout() {
	zones := strings.Split(d.zoning, ",")
	if len(zones) != 3 {
		fmt.Println("Zoning must have exactly 3 comma-separated values")
		os.Exit(1)
	}
	for i := 0; i < len(zones); i++ {
		val, _ := strconv.Atoi(zones[i])
		d.blockAngleOffset = append(d.blockAngleOffset, val/2)
	}

	track := 0
	angleOffset := 2 * d.blockAngleOffset[track]
	for angle := 0; angle < 360; angle += angleOffset {
		block := angle / angleOffset
		d.blockToTrackMap[block] = track
		d.blockToAngleMap[block] = float64(angle)
		d.blockInfoList = append(d.blockInfoList, BlockInfo{track, float64(angle), block})
		d.tracksBeginEnd[track] = [2]int{0, block}
	}
	pblock := d.tracksBeginEnd[track][1] + 1

	track = 1
	skew := d.skew
	angleOffset = 2 * d.blockAngleOffset[track]
	for angle := 0; angle < 360; angle += angleOffset {
		block := (angle / angleOffset) + pblock
		d.blockToTrackMap[block] = track
		d.blockToAngleMap[block] = float64(angle + (angleOffset * skew))
		d.blockInfoList = append(d.blockInfoList, BlockInfo{track, float64(angle + (angleOffset * skew)), block})
		d.tracksBeginEnd[track] = [2]int{pblock, block}
	}
	pblock = d.tracksBeginEnd[track][1] + 1

	track = 2
	skew = 2 * d.skew
	angleOffset = 2 * d.blockAngleOffset[track]
	for angle := 0; angle < 360; angle += angleOffset {
		block := (angle / angleOffset) + pblock
		d.blockToTrackMap[block] = track
		d.blockToAngleMap[block] = float64(angle + (angleOffset * skew))
		d.blockInfoList = append(d.blockInfoList, BlockInfo{track, float64(angle + (angleOffset * skew)), block})
		d.tracksBeginEnd[track] = [2]int{pblock, block}
	}
	d.maxBlock = d.tracksBeginEnd[track][1] + 1

	for i := range d.blockToAngleMap {
		d.blockToAngleMap[i] = mod(d.blockToAngleMap[i]+180, 360)
	}
}

func (d *Disk) PrintAddrDescMessage(value string) {
	fmt.Printf("Bad address description (%s)\n", value)
	fmt.Println("The address description must be a comma-separated list of length three...")
	os.Exit(1)
}

func (d *Disk) MakeRequests(addr, addrDesc string) []int {
	var numRequests, maxRequest, minRequest int
	if addr == "-1" {
		desc := strings.Split(addrDesc, ",")
		if len(desc) != 3 {
			d.PrintAddrDescMessage(addrDesc)
		}
		numRequests, _ = strconv.Atoi(desc[0])
		maxRequest, _ = strconv.Atoi(desc[1])
		minRequest, _ = strconv.Atoi(desc[2])

		if maxRequest == -1 {
			maxRequest = d.maxBlock
		}
		var tmpList []int
		for i := 0; i < numRequests; i++ {
			tmpList = append(tmpList, int(d.rng.Float64()*float64(maxRequest))+minRequest)
		}
		return tmpList
	} else {
		var tmpList []int
		for _, x := range strings.Split(addr, ",") {
			val, _ := strconv.Atoi(x)
			tmpList = append(tmpList, val)

		}
		return tmpList
	}
}

func (d *Disk) AddQueueEntry(block, index int) {
	d.requestQueue = append(d.requestQueue, Request{block, index})
	d.requestState = append(d.requestState, STATENull)
}

func (d *Disk) SwitchState(newState int) {
	d.state = newState
	d.requestState[d.currentIndex] = newState
}

func (d *Disk) RadiallyCloseTo(a1, a2 float64) bool {
	v := math.Abs(a1 - a2)
	return v < d.rotateSpeed
}

func (d *Disk) DoneWithTransfer() bool {
	angleOffset := float64(d.blockAngleOffset[int((d.armTrack))])
	targetAngle := mod(d.blockToAngleMap[d.currentBlock]+angleOffset, 360)

	if d.RadiallyCloseTo(d.angle, targetAngle) {
		d.SwitchState(STATEDone)
		d.requestCount++
		return true
	}
	return false
}

func (d *Disk) DoneWithRotation() bool {
	angleOffset := float64(d.blockAngleOffset[int(d.armTrack)])
	targetAngle := mod(d.blockToAngleMap[d.currentBlock]-angleOffset, 360)

	if d.RadiallyCloseTo(d.angle, targetAngle) {
		d.SwitchState(STATEXfer)
		return true
	}
	return false
}

func (d *Disk) PlanSeek(track int) {
	d.seekBegin = d.timer
	d.SwitchState(STATESeek)
	if float64(track) == d.armTrack {
		d.rotBegin = d.timer
		d.SwitchState(STATERotate)
		return
	}
	d.armTarget = float64(track)
	d.armTargetX1 = -d.tracks[track] - float64(d.trackWidth)/2.0

	if float64(track) >= d.armTrack {
		d.armSpeed = d.seekSpeedBase
	} else {
		d.armSpeed = -d.seekSpeedBase
	}
}

func (d *Disk) DoneWithSeek() bool {
	d.armX1 += d.armSpeed
	if (d.armSpeed > 0.0 && d.armX1 >= d.armTargetX1) || (d.armSpeed < 0.0 && d.armX1 <= d.armTargetX1) {
		d.armTrack = d.armTarget
		return true
	}
	return false
}

func (d *Disk) DoSATF(rList []Request) (int, int) {
	minBlock := -1
	minIndex := -1
	minEst := -1.0

	for _, req := range rList {
		block := req.block
		index := req.index
		if d.requestState[index] == STATEDone {
			continue
		}
		track := d.blockToTrackMap[block]
		angle := d.blockToAngleMap[block]

		dist := math.Abs(d.armTrack - float64(track))
		seekEst := (float64(d.trackWidth) / d.seekSpeedBase) * dist

		angleOffset := float64(d.blockAngleOffset[track])
		angleAtArrival := mod(d.angle+(seekEst*d.rotateSpeed), 360)
		rotDist := mod((angle-d.angle)-angleAtArrival, 360)
		rotEst := rotDist / d.rotateSpeed
		xferEst := (angleOffset * 2.0) / d.rotateSpeed
		totalEst := seekEst + rotEst + xferEst

		if minEst == -1.0 || totalEst < minEst {
			minEst = totalEst
			minBlock = block
			minIndex = index
		}
	}

	return minBlock, minIndex
}

func (d *Disk) DoSSTF(rList []Request) []Request {
	minDist := float64(MAXTRACKS)
	var trackList []Request

	for _, req := range rList {
		if d.requestState[req.index] == STATEDone {
			continue
		}
		track := float64(d.blockToTrackMap[req.block])
		dist := math.Abs(d.armTrack - track)

		if dist < minDist {
			trackList = []Request{req}
			minDist = dist
		} else if dist == minDist {
			trackList = append(trackList, req)
		}
	}
	return trackList
}

func (d *Disk) UpdateWindow() {
	if d.fairWindow == -1 && d.currWindow > 0 && d.currWindow < len(d.requestQueue) {
		d.currWindow++
	}
}

func (d *Disk) GetWindow() int {
	if d.currWindow <= -1 {
		return len(d.requestQueue)
	}
	if d.fairWindow != -1 {
		if d.requestCount > 0 && (d.requestCount%d.fairWindow == 0) {
			d.currWindow = d.currWindow + d.fairWindow
		}
		return d.currWindow
	}
	return d.currWindow
}

func (d *Disk) GetNextIO() {
	if d.requestCount == len(d.requestQueue) {
		d.PrintStats()
		d.isDone = true
		return
	}

	if d.policy == "FIFO" {
		req := d.requestQueue[d.requestCount]
		d.currentBlock, d.currentIndex = req.block, req.index
	} else if d.policy == "SATF" || d.policy == "BSATF" {
		endIndex := d.GetWindow()
		if endIndex > len(d.requestQueue) {
			endIndex = len(d.requestQueue)
		}
		d.currentBlock, d.currentIndex = d.DoSATF(d.requestQueue[0:endIndex])
	} else if d.policy == "SSTF" {
		trackList := d.DoSSTF(d.requestQueue[0:d.GetWindow()])
		d.currentBlock, d.currentIndex = d.DoSATF(trackList)
	} else {
		fmt.Printf("policy (%s) not implemented\n", d.policy)
		os.Exit(1)
	}

	d.PlanSeek(d.blockToTrackMap[d.currentBlock])

	if len(d.lateRequests) > 0 && d.lateCount < len(d.lateRequests) {
		d.AddQueueEntry(d.lateRequests[d.lateCount], len(d.requestQueue))
		d.lateCount++
	}
}

func (d *Disk) DoRequestStats() {
	seekTime := float64(d.rotBegin - d.seekBegin)
	rotTime := float64(d.xferBegin - d.rotBegin)
	xferTime := float64(d.timer - d.xferBegin)
	totalTime := float64(d.timer - d.seekBegin)

	if d.compute {
		fmt.Printf("Block: %3d  Seek:%3d  Rotate:%3d  Transfer:%3d  Total:%4d\n",
			d.currentBlock, int(seekTime), int(rotTime), int(xferTime), int(totalTime))
	}

	d.seekTotal += seekTime
	d.rotTotal += rotTime
	d.xferTotal += xferTime
}

func (d *Disk) PrintStats() {
	if d.compute {
		fmt.Printf("\nTOTALS      Seek:%3d  Rotate:%3d  Transfer:%3d  Total:%4d\n",
			int(d.seekTotal), int(d.rotTotal), int(d.xferTotal), d.timer)
	}
}

func (d *Disk) Animate() {
	d.timer++
	d.angle += d.rotateSpeed
	if d.angle >= 360.0 {
		d.angle -= 360.0
	}

	if d.state == STATESeek {
		if d.DoneWithSeek() {
			d.rotBegin = d.timer
			d.SwitchState(STATERotate)
		}
	}
	if d.state == STATERotate {
		if d.DoneWithRotation() {
			d.xferBegin = d.timer
			d.SwitchState(STATEXfer)
		}
	}
	if d.state == STATEXfer {
		if d.DoneWithTransfer() {
			d.DoRequestStats()
			d.SwitchState(STATEDone)
			d.UpdateWindow()
			currentBlock := d.currentBlock
			d.GetNextIO()
			if d.isDone {
				return
			}
			nextBlock := d.currentBlock
			if d.blockToTrackMap[currentBlock] == d.blockToTrackMap[nextBlock] {
				if (currentBlock == d.tracksBeginEnd[int(d.armTrack)][1] && nextBlock == d.tracksBeginEnd[int(d.armTrack)][0]) || (currentBlock+1 == nextBlock) {
					d.rotBegin, d.seekBegin, d.xferBegin = d.timer, d.timer, d.timer
					d.SwitchState(STATEXfer)
				}
			}
		}
	}
}

func (d *Disk) Go() {
	d.GetNextIO()
	for !d.isDone {
		d.Animate()
	}
}

func main() {
	seed := flag.Int("s", 0, "Random seed")
	addr := flag.String("a", "-1", "Request list (comma-separated) [-1 -> use addrDesc]")
	addrDesc := flag.String("A", "5,-1,0", "Num requests, max request (-1->all), min request")
	seekSpeed := flag.Float64("S", 1.0, "Speed of seek")
	rotateSpeed := flag.Float64("R", 1.0, "Speed of rotation")
	policy := flag.String("p", "FIFO", "Scheduling policy (FIFO, SSTF, SATF, BSATF)")
	window := flag.Int("w", -1, "Size of scheduling window (-1 -> all)")
	skew := flag.Int("o", 0, "Amount of skew (in blocks)")
	zoning := flag.String("z", "30,30,30", "Angles between blocks on outer,middle,inner tracks")
	graphics := flag.Bool("G", false, "Turn on graphics (Ignored in Go version)")
	lateAddr := flag.String("l", "-1", "Late: request list (comma-separated) [-1 -> random]")
	lateAddrDesc := flag.String("L", "0,-1,0", "Num requests, max request (-1->all), min request")
	compute := flag.Bool("c", false, "Compute the answers")

	flag.Parse()

	fmt.Printf("OPTIONS seed %d\n", *seed)
	fmt.Printf("OPTIONS addr %s\n", *addr)
	fmt.Printf("OPTIONS addrDesc %s\n", *addrDesc)
	fmt.Printf("OPTIONS seekSpeed %v\n", *seekSpeed)
	fmt.Printf("OPTIONS rotateSpeed %v\n", *rotateSpeed)
	fmt.Printf("OPTIONS skew %d\n", *skew)
	fmt.Printf("OPTIONS window %d\n", *window)
	fmt.Printf("OPTIONS policy %s\n", *policy)
	fmt.Printf("OPTIONS compute %v\n", *compute)
	fmt.Printf("OPTIONS graphics %v (Note: Graphics not supported in Go CLI)\n", *graphics)
	fmt.Printf("OPTIONS zoning %s\n", *zoning)
	fmt.Printf("OPTIONS lateAddr %s\n", *lateAddr)
	fmt.Printf("OPTIONS lateAddrDesc %s\n\n", *lateAddrDesc)

	if *window == 0 {
		fmt.Printf("Scheduling window (%d) must be positive or -1 (which means a full window)\n", *window)
		os.Exit(1)
	}

	if *graphics && !*compute {
		fmt.Println("\nWARNING: Setting compute flag to True, as graphics are requested (but not drawn in Go)\n")
		*compute = true
	}

	disk := NewDisk(*addr, *addrDesc, *lateAddr, *lateAddrDesc, *policy, *seekSpeed, *rotateSpeed, *skew, *window, *compute, *zoning, *seed)
	disk.Go()
}
