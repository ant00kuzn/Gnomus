package world

import (
	"bytes"
	"sync"
)

type BlockID int

type Chunk struct {
	Blocks [16][256][16]BlockID // Простой плоский чанк 16x256x16
}

type World struct {
	Chunks map[[2]int]*Chunk // [chunkX, chunkZ] -> Chunk
	Mutex  sync.RWMutex
}

var worldd = &World{
	Chunks: make(map[[2]int]*Chunk),
}

// GetChunk возвращает чанк по координатам, создаёт если нет
func GetChunk(x, z int) *Chunk {
	worldd.Mutex.Lock()
	defer worldd.Mutex.Unlock()
	key := [2]int{x, z}
	ch, ok := worldd.Chunks[key]
	if !ok {
		ch = generateFlatChunk()
		worldd.Chunks[key] = ch
	}
	return ch
}

// generateFlatChunk создаёт плоский чанк (например, трава на 64 уровне)
func generateFlatChunk() *Chunk {
	ch := &Chunk{}
	for x := 0; x < 16; x++ {
		for z := 0; z < 16; z++ {
			for y := 0; y < 64; y++ {
				ch.Blocks[x][y][z] = 2 // 2 = grass (пример)
			}
		}
	}
	return ch
}

// SerializeChunkToPacket сериализует чанк в пакет Minecraft (очень упрощённо, только блоки, без lighting, biomes и т.д.)
func SerializeChunkToPacket(chunkX, chunkZ int, ch *Chunk) []byte {
	// Для простоты: только один чанк, полный, без биомов, lighting и т.д.
	var buf bytes.Buffer
	// Packet ID (0x22), chunkX, chunkZ, full chunk, primary bitmask, heightmaps, data size, data
	buf.Write(encodeVarInt(0x22)) // packet id
	buf.Write(intToBytes(chunkX)) // chunk X
	buf.Write(intToBytes(chunkZ)) // chunk Z
	buf.WriteByte(1)              // full chunk
	buf.Write([]byte{0xFF, 0xFF}) // primary bitmask (все секции)
	// Heightmaps (NBT, заглушка)
	buf.Write(encodeVarInt(1)) // length
	buf.WriteByte(0)           // dummy
	// Data size (заглушка)
	buf.Write(encodeVarInt(0))
	// Block data (очень упрощённо)
	for y := 0; y < 256; y++ {
		for z := 0; z < 16; z++ {
			for x := 0; x < 16; x++ {
				buf.WriteByte(byte(ch.Blocks[x][y][z]))
			}
		}
	}
	return buf.Bytes()
}

func intToBytes(i int) []byte {
	return []byte{
		byte(i >> 24),
		byte(i >> 16),
		byte(i >> 8),
		byte(i),
	}
}

// encodeVarInt encodes an integer as a Minecraft-style VarInt.
func encodeVarInt(value int) []byte {
	var buf []byte
	v := uint32(value)
	for {
		temp := byte(v & 0x7F)
		v >>= 7
		if v != 0 {
			temp |= 0x80
		}
		buf = append(buf, temp)
		if v == 0 {
			break
		}
	}
	return buf
}
