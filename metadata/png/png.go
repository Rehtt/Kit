package png

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"hash/crc32"
	"io"
	"os"
)

const (
	pngHeader = "\x89PNG\r\n\x1a\n"
	iTXtType  = "iTXt"
)

type Metadata struct {
	Keyword           string
	LanguageTag       string
	TranslatedKeyword string
	Text              string
}

func ReadMetadata(filename string) ([]Metadata, error) {
	f, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	return ReadMetadataFromReader(f)
}

func ReadMetadataFromReader(r io.Reader) ([]Metadata, error) {
	header := make([]byte, 8)
	if _, err := io.ReadFull(r, header); err != nil {
		return nil, err
	}
	if string(header) != pngHeader {
		return nil, fmt.Errorf("不是有效的PNG文件")
	}

	var metadataList []Metadata
	for {
		var length uint32
		if err := binary.Read(r, binary.BigEndian, &length); err != nil {
			if err == io.EOF {
				break
			}
			return nil, err
		}

		chunkType := make([]byte, 4)
		if _, err := io.ReadFull(r, chunkType); err != nil {
			return nil, err
		}

		chunkData := make([]byte, length)
		if _, err := io.ReadFull(r, chunkData); err != nil {
			return nil, err
		}

		var crc uint32
		if err := binary.Read(r, binary.BigEndian, &crc); err != nil {
			return nil, err
		}

		crcData := append(chunkType, chunkData...)
		if crc32.ChecksumIEEE(crcData) != crc {
			return nil, fmt.Errorf("CRC校验失败")
		}

		if string(chunkType) == iTXtType {
			metadata, err := parseITXt(chunkData)
			if err != nil {
				return nil, err
			}
			metadataList = append(metadataList, metadata)
		}

		if string(chunkType) == "IEND" {
			break
		}
	}

	return metadataList, nil
}

func parseITXt(data []byte) (Metadata, error) {
	var metadata Metadata
	pos := 0

	keywordEnd := bytes.IndexByte(data[pos:], 0)
	if keywordEnd == -1 {
		return metadata, fmt.Errorf("iTXt块格式错误")
	}
	metadata.Keyword = string(data[pos : pos+keywordEnd])
	pos += keywordEnd + 1

	if pos+2 > len(data) {
		return metadata, fmt.Errorf("iTXt块格式错误")
	}
	pos += 2 // skip compression flag and method

	langEnd := bytes.IndexByte(data[pos:], 0)
	if langEnd == -1 {
		return metadata, fmt.Errorf("iTXt块格式错误")
	}
	metadata.LanguageTag = string(data[pos : pos+langEnd])
	pos += langEnd + 1

	transEnd := bytes.IndexByte(data[pos:], 0)
	if transEnd == -1 {
		return metadata, fmt.Errorf("iTXt块格式错误")
	}
	metadata.TranslatedKeyword = string(data[pos : pos+transEnd])
	pos += transEnd + 1

	metadata.Text = string(data[pos:])
	return metadata, nil
}

func WriteMetadata(inputFile, outputFile string, metadata []Metadata) error {
	input, err := os.Open(inputFile)
	if err != nil {
		return err
	}
	defer input.Close()

	output, err := os.Create(outputFile)
	if err != nil {
		return err
	}
	defer output.Close()

	return WriteMetadataToWriter(input, output, metadata)
}

func WriteMetadataToWriter(r io.Reader, w io.Writer, metadata []Metadata) error {
	header := make([]byte, 8)
	if _, err := io.ReadFull(r, header); err != nil {
		return err
	}
	if string(header) != pngHeader {
		return fmt.Errorf("不是有效的PNG文件")
	}

	if _, err := w.Write(header); err != nil {
		return err
	}

	iTXtInserted := false
	for {
		var length uint32
		if err := binary.Read(r, binary.BigEndian, &length); err != nil {
			if err == io.EOF {
				break
			}
			return err
		}

		chunkType := make([]byte, 4)
		if _, err := io.ReadFull(r, chunkType); err != nil {
			return err
		}

		chunkData := make([]byte, length)
		if _, err := io.ReadFull(r, chunkData); err != nil {
			return err
		}

		var crc uint32
		if err := binary.Read(r, binary.BigEndian, &crc); err != nil {
			return err
		}

		if string(chunkType) == "IEND" && !iTXtInserted {
			for _, m := range metadata {
				iTXtChunk, err := createITXtChunk(m)
				if err != nil {
					return err
				}
				if _, err := w.Write(iTXtChunk); err != nil {
					return err
				}
			}
			iTXtInserted = true
		}

		if err := binary.Write(w, binary.BigEndian, length); err != nil {
			return err
		}
		if _, err := w.Write(chunkType); err != nil {
			return err
		}
		if _, err := w.Write(chunkData); err != nil {
			return err
		}
		if err := binary.Write(w, binary.BigEndian, crc); err != nil {
			return err
		}

		if string(chunkType) == "IEND" {
			break
		}
	}

	return nil
}

func createITXtChunk(metadata Metadata) ([]byte, error) {
	var buf bytes.Buffer
	buf.WriteString(metadata.Keyword)
	buf.WriteByte(0)
	buf.WriteByte(0) // uncompressed
	buf.WriteByte(0)
	buf.WriteString(metadata.LanguageTag)
	buf.WriteByte(0)
	buf.WriteString(metadata.TranslatedKeyword)
	buf.WriteByte(0)
	buf.WriteString(metadata.Text)

	chunkData := buf.Bytes()
	var result bytes.Buffer

	if err := binary.Write(&result, binary.BigEndian, uint32(len(chunkData))); err != nil {
		return nil, err
	}

	result.WriteString(iTXtType)
	result.Write(chunkData)

	crcData := append([]byte(iTXtType), chunkData...)
	crc := crc32.ChecksumIEEE(crcData)
	if err := binary.Write(&result, binary.BigEndian, crc); err != nil {
		return nil, err
	}

	return result.Bytes(), nil
}
