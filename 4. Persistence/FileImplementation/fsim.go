package main

import (
	"flag"
	"fmt"
	"math/rand"
	"os"
	"strings"
)

var (
	DEBUG      = false
	PrintOps   = true
	PrintState = true
	PrintFinal = true
)

func dprint(str string) {
	if DEBUG {
		fmt.Println(str)
	}
}

func assert(cond bool, msg string) {
	if !cond {
		panic("Assertion failed: " + msg)
	}
}

type Bitmap struct {
	size         int
	bmap         []int
	numAllocated int
}

func NewBitmap(size int) *Bitmap {
	return &Bitmap{
		size:         size,
		bmap:         make([]int, size),
		numAllocated: 0,
	}
}

func (b *Bitmap) Alloc() int {
	for i := 0; i < b.size; i++ {
		if b.bmap[i] == 0 {
			b.bmap[i] = 1
			b.numAllocated++
			return i
		}
	}
	return -1
}

func (b *Bitmap) Free(num int) {
	assert(b.bmap[num] == 1, "freeing unallocated block")
}

func (b *Bitmap) MarkAllocated(num int) {
	assert(b.bmap[num] == 0, "marking already allocated block")
	b.numAllocated++
	b.bmap[num] = 1
}

func (b *Bitmap) NumFree() int {
	return b.size - b.numAllocated
}

func (b *Bitmap) Dump() string {
	s := ""
	for i := 0; i < len(b.bmap); i++ {
		s += string(b.bmap[i])
	}
	return s
}

type DirEntry struct {
	name string
	inum int
}

type Block struct {
	ftype   string
	dirUsed int
	maxUsed int
	dirList []DirEntry
	data    string
}

func NewBlock(ftype string) *Block {
	assert(ftype == "d" || ftype == "f" || ftype == "free", "invalid block type")
	return &Block{
		ftype:   ftype,
		dirUsed: 0,
		maxUsed: 32,
		dirList: []DirEntry{},
		data:    "",
	}
}

func (b *Block) Dump() string {
	if b.ftype == "free" {
		return "[]"
	} else if b.ftype == "d" {
		rc := ""
		for _, d := range b.dirList {
			short := fmt.Sprintf("(%s,%d)", d.name, d.inum)
			if rc == "" {
				rc = short
			} else {
				rc += " " + short
			}
		}
		return "[" + rc + "]"
	}

	return "[" + b.data + "]"
}

func (b *Block) SetType(ftype string) {
	assert(b.ftype == "free", "block not free")
	b.ftype = ftype
}

func (b *Block) AddData(data string) {
	assert(b.ftype == "f", "not a file block")
	b.data = data
}

func (b *Block) GetNumEntries() int {
	assert(b.ftype == "d", "not a directory block")
	return b.dirUsed
}

func (b *Block) GetFreeEntries() int {
	assert(b.ftype == "d", "not a directory block")
	return b.maxUsed - b.dirUsed
}

func (b *Block) GetEntry(num int) DirEntry {
	assert(b.ftype == "d", "not a directory block")
	assert(num < b.dirUsed, "invalid num")
	return b.dirList[num]
}

func (b *Block) AddDirEntry(name string, inum int) {
	assert(b.ftype == "d", "not a directory file")
	b.dirList = append(b.dirList, DirEntry{name: name, inum: inum})
	b.dirUsed++
	assert(b.dirUsed <= b.maxUsed, "over max used entry")
}

func (b *Block) DelDirEntry(name string) {
	assert(b.ftype == "d", "not a directory file")
	tname := strings.Split(string(name), "/")
	dname := tname[len(tname)-1]
	for i, entry := range b.dirList {
		if entry.name == dname {
			b.dirList = append(b.dirList[:i], b.dirList[i+1:]...)
			b.dirUsed--
			return
		}
	}
	panic("Directory entry not found")
}

func (b *Block) DirEntryExists(name string) bool {
	assert(b.ftype == "d", "not a directiry file")
	for _, d := range b.dirList {
		if name == d.name {
			return true
		}
	}
	return false
}

func (b *Block) Free() {
	assert(b.ftype != "free", "freeing already free block")
	if b.ftype == "d" {
		assert(b.dirUsed == 2, "only dot, dotdot here")
		b.dirUsed = 0
	}
	b.data = ""
	b.ftype = "free"
}

type Inode struct {
	ftype  string
	addr   int //block number that inode points to
	refCnt int //number of directory entries or file point to inode
}

func NewInode() *Inode {
	in := &Inode{}
	in.SetAll("free", -1, 1)
	return in
}

func (i *Inode) SetAll(ftype string, addr int, refCnt int) {
	assert(ftype == "d" || ftype == "f" || ftype == "free", "invalid inode type")
	i.ftype = ftype
	i.addr = addr
	i.refCnt = refCnt
}

func (i *Inode) IncRefCnt() {
	i.refCnt++
}

func (i *Inode) DecRefCnt() {
	i.refCnt--
}

func (i *Inode) GetRefCnt() int {
	return i.refCnt
}

func (i *Inode) SetAddr(block int) { i.addr = block }

func (i *Inode) SetType(ftype string) {
	assert(ftype == "d" || ftype == "f" || ftype == "free", "invalid ftype inode")
	i.ftype = ftype
}

func (i *Inode) GetSize() int {
	if i.addr == -1 {
		return 0
	} else {
		return 1
	}
}

func (i *Inode) GetAddr() int {
	return i.addr
}

func (i *Inode) GetType() string {
	return i.ftype
}

func (i *Inode) Free() {
	i.ftype = "free"
	i.addr = -1
}

type FileSystem struct {
	numInodes  int
	numData    int
	ibitmap    *Bitmap
	inodes     []*Inode
	dbitmap    *Bitmap
	data       []*Block
	ROOT       int
	files      []string
	dirs       []string
	nameToInum map[string]int
	rng        *rand.Rand
}

func NewFileSystem(numInodes, numData int, seed int64) *FileSystem {
	fs := &FileSystem{
		numInodes:  numInodes,
		numData:    numData,
		ibitmap:    NewBitmap(numInodes),
		dbitmap:    NewBitmap(numData),
		ROOT:       0,
		files:      []string{},
		dirs:       []string{"/"},
		nameToInum: map[string]int{"/": 0},
		rng:        rand.New(rand.NewSource(seed)),
	}

	for i := 0; i < numInodes; i++ {
		fs.inodes = append(fs.inodes, NewInode())
	}

	for i := 0; i < numData; i++ {
		fs.data = append(fs.data, NewBlock("free"))
	}

	fs.ibitmap.MarkAllocated(fs.ROOT)
	fs.inodes[fs.ROOT].SetAll("d", 0, 2)
	fs.dbitmap.MarkAllocated(fs.ROOT)
	fs.data[0].SetType("d")
	fs.data[0].AddDirEntry(".", fs.ROOT)
	fs.data[0].AddDirEntry("..", fs.ROOT)

	return fs
}

func (fs *FileSystem) Dump() {
	fmt.Println("inode bitmap ", fs.ibitmap.Dump())
	fmt.Print("inodes       ")
	for i := 0; i < fs.numInodes; i++ {
		ftype := fs.inodes[i].GetType()
		if ftype == "free" {
			fmt.Print("[]")
		} else {
			fmt.Printf("[%s a:%d r:%d]", ftype, fs.inodes[i].GetAddr(), fs.inodes[i].GetRefCnt())
		}
	}
	fmt.Println()
	fmt.Println("data bitmap  ", fs.dbitmap.Dump())
	fmt.Print("data         ")
	for i := 0; i < fs.numData; i++ {
		fmt.Print(fs.data[i].Dump())
	}
	fmt.Println()
}

func (fs *FileSystem) makeName() string {
	fChars := []rune{'b', 'c', 'd', 'f', 'g', 'h', 'j', 'k', 'l', 'm', 'n', 'p', 's', 't', 'v', 'w', 'x', 'y', 'z'}
	sChars := []rune{'a', 'e', 'i', 'o', 'u'}
	lChars := []rune{'b', 'c', 'd', 'f', 'g', 'j', 'k', 'l', 'm', 'n', 'p', 's', 't', 'v', 'w', 'x', 'y', 'z'}

	f := fChars[fs.rng.Intn(len(fChars))]
	s := sChars[fs.rng.Intn(len(sChars))]
	l := lChars[fs.rng.Intn(len(lChars))]

	return string([]rune{f, s, l})
}

func (fs *FileSystem) InodeAlloc() int {
	return fs.ibitmap.Alloc()
}

func (fs *FileSystem) InodeFree(inum int) {
	fs.ibitmap.Free(inum)
	fs.inodes[inum].Free()
}

func (fs *FileSystem) DataAlloc() int {
	return fs.dbitmap.Alloc()
}

func (fs *FileSystem) DataFree(bnum int) {
	fs.dbitmap.Free(bnum)
	fs.data[bnum].Free()
}

func (fs *FileSystem) GetParent(name string) string {
	tmp := strings.Split(name, "/")
	if len(tmp) == 2 {
		return "/"
	}
	pname := ""
	for i := 0; i < len(tmp)-1; i++ {
		pname = pname + "/" + tmp[i]
	}
	return pname
}

func (fs *FileSystem) DeleteFile(tfile string) int {
	if PrintOps {
		fmt.Printf("unlink(\"%s\");\n", tfile)
	}
	inum := fs.nameToInum[tfile]
	ftype := fs.inodes[inum].GetType()
	if fs.inodes[inum].GetRefCnt() == 1 {
		dblock := fs.inodes[inum].GetAddr()
		if dblock != -1 {
			fs.DataFree(dblock)
		}
		fs.InodeFree(inum)
	} else {
		fs.inodes[inum].DecRefCnt()
	}

	parent := fs.GetParent(tfile)
	pinum := fs.nameToInum[parent]
	pblock := fs.inodes[pinum].GetAddr()
	if ftype == "d" {
		fs.inodes[pinum].DecRefCnt()
	}
	fs.data[pblock].DelDirEntry(tfile)
	for i, f := range fs.files {
		if f == tfile {
			fs.files = append(fs.files[:i], fs.files[i+1:]...)
			break
		}
	}
	delete(fs.nameToInum, tfile)
	return 0
}

func (fs *FileSystem) CreateLink(target, newFile, parent string) int {
	parentInum := fs.nameToInum[parent]
	pblock := fs.inodes[parentInum].GetAddr()
	if fs.data[pblock].GetFreeEntries() <= 0 {
		return -1
	}
	if fs.data[pblock].DirEntryExists(newFile) {
		return -1
	}

	tinum := fs.nameToInum[target]
	fs.inodes[tinum].IncRefCnt()

	tmp := strings.Split(newFile, "/")
	ename := tmp[len(tmp)-1]
	fs.data[pblock].AddDirEntry(ename, tinum)
	return tinum
}

func (fs *FileSystem) CreateFile(parent, newFile, ftype string) int {
	parentInum := fs.nameToInum[parent]

	pblock := fs.inodes[parentInum].GetAddr()
	if fs.data[pblock].GetFreeEntries() <= 0 {
		return -1
	}

	block := fs.inodes[parentInum].GetAddr()
	if fs.data[block].DirEntryExists(newFile) {
		return -1
	}
	inum := fs.InodeAlloc()
	if inum == -1 {
		return -1
	}

	fblock := -1
	refCnt := -1
	if ftype == "d" {
		refCnt = 2
		fblock = fs.DataAlloc()
		if fblock == -1 {
			fs.InodeFree(inum)
			return -1
		} else {
			fs.data[fblock].SetType("d")
			fs.data[fblock].AddDirEntry(".", inum)
			fs.data[fblock].AddDirEntry("..", parentInum)
		}
	} else {
		refCnt = 1
	}

	fs.inodes[inum].SetAll(ftype, fblock, refCnt)

	if ftype == "d" {
		fs.inodes[parentInum].IncRefCnt()
	}
	fs.data[pblock].AddDirEntry(newFile, inum)
	return inum
}

func (fs *FileSystem) WriteFile(tfile, data string) int {
	inum := fs.nameToInum[tfile]
	curSize := fs.inodes[inum].GetSize()

	if curSize == 1 {
		return -1
	}

	fblock := fs.dbitmap.Alloc()
	if fblock == -1 {
		return -1
	}
	fs.data[fblock].SetType("f")
	fs.data[fblock].AddData(data)
	fs.inodes[inum].SetAddr(fblock)

	if PrintOps {
		fmt.Printf("fd=open(\"%s\", O_WRONLY|O_APPEND); write(fd, buf, BLOCKSIZE); close(fd);\n", tfile)
	}
	return 0
}

func (fs *FileSystem) doDelete() int {
	if len(fs.files) == 0 {
		return -1
	}
	dfile := fs.files[fs.rng.Intn(len(fs.files))]
	return fs.DeleteFile(dfile)
}

func (fs *FileSystem) doLink() int {
	if len(fs.files) == 0 {
		return -1
	}
	parent := fs.dirs[fs.rng.Intn(len(fs.dirs))]
	nfile := fs.makeName()
	target := fs.files[fs.rng.Intn(len(fs.files))]

	fullName := parent + "/" + nfile
	if parent == "/" {
		fullName = parent + nfile
	}

	inum := fs.CreateLink(target, nfile, parent)
	if inum >= 0 {
		fs.files = append(fs.files, fullName)
		fs.nameToInum[fullName] = inum
		if PrintOps {
			fmt.Printf("link(\"%s\", \"%s\");\n", target, fullName)
		}
		return 0
	}
	return -1
}

func (fs *FileSystem) doCreate(ftype string) int {
	parent := fs.dirs[fs.rng.Intn(len(fs.dirs))]
	nfile := fs.makeName()

	fullName := parent + "/" + nfile
	if parent == "/" {
		fullName = parent + nfile
	}

	inum := fs.CreateFile(parent, nfile, ftype)
	if inum >= 0 {
		if ftype == "d" {
			fs.dirs = append(fs.dirs, fullName)
		} else {
			fs.files = append(fs.files, fullName)
		}
		fs.nameToInum[fullName] = inum

		if PrintOps {
			pDir := parent
			if parent == "/" {
				pDir = ""
			}
			if ftype == "d" {
				fmt.Printf("mkdir(\"%s/%s\");\n", pDir, nfile)
			} else {
				fmt.Printf("creat(\"%s/%s\");\n", pDir, nfile)
			}
		}
		return 0
	}
	return -1
}

func (fs *FileSystem) doAppend() int {
	if len(fs.files) == 0 {
		return -1
	}
	afile := fs.files[fs.rng.Intn(len(fs.files))]
	data := string(rune('a' + fs.rng.Intn(26)))
	return fs.WriteFile(afile, data)
}

func (fs *FileSystem) Run(numRequests int) {
	fmt.Println("Initial state")
	fmt.Println()
	fs.Dump()
	fmt.Println()

	for i := 0; i < numRequests; i++ {
		if !PrintOps {
			fmt.Println("Which operation took place?")
		}
		rc := -1
		for rc == -1 {
			r := fs.rng.Float64()
			if r < 0.3 {
				rc = fs.doAppend()
			} else if r < 0.5 {
				rc = fs.doDelete()
			} else if r < 0.7 {
				rc = fs.doLink()
			} else {
				if fs.rng.Float64() < 0.75 {
					rc = fs.doCreate("f")
				} else {
					rc = fs.doCreate("d")
				}
			}

			if fs.ibitmap.NumFree() == 0 {
				fmt.Println("File system out of inodes; rerun with more via command-line flag?")
				os.Exit(1)
			}
			if fs.dbitmap.NumFree() == 0 {
				fmt.Println("File system out of data blocks; rerun with more via command-line flag?")
				os.Exit(1)
			}
		}

		if PrintState {
			fmt.Println()
			fs.Dump()
			fmt.Println()
		} else {
			fmt.Println()
			fmt.Println("  State of file system (inode bitmap, inodes, data bitmap, data)?")
			fmt.Println()
		}
	}

	if PrintFinal {
		fmt.Println()
		fmt.Println("Summary of files, directories::")
		fmt.Println()
		fmt.Printf("  Files:       %v\n", fs.files)
		fmt.Printf("  Directories: %v\n", fs.dirs)
		fmt.Println()
	}
}

func main1() {
	seed := flag.Int("s", 0, "the random seed")
	numInodes := flag.Int("i", 8, "number of inodes in file system")
	numData := flag.Int("d", 8, "number of data blocks in file system")
	numRequests := flag.Int("n", 10, "number of requests to simulate")
	reverse := flag.Bool("r", false, "instead of printing state, print ops")
	printFinalFlag := flag.Bool("p", false, "print the final set of files/dirs")
	compute := flag.Bool("c", false, "compute answers for me")

	flag.Parse()

	fmt.Println("ARG seed", *seed)
	fmt.Println("ARG numInodes", *numInodes)
	fmt.Println("ARG numData", *numData)
	fmt.Println("ARG numRequests", *numRequests)
	fmt.Println("ARG reverse", *reverse)
	fmt.Println("ARG printFinal", *printFinalFlag)
	fmt.Println()

	// Handle flags logic
	if *reverse {
		PrintState = false
		PrintOps = true
	} else {
		PrintState = true
		PrintOps = false
	}

	if *compute {
		PrintOps = true
		PrintState = true
	}

	PrintFinal = *printFinalFlag

	// Khởi tạo FileSystem
	fs := NewFileSystem(*numInodes, *numData, int64(*seed))

	// Chạy mô phỏng
	fs.Run(*numRequests)
}
