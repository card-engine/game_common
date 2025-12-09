package utils

import (
	"bytes"
	"compress/gzip"
	"io"

	"github.com/andybalholm/brotli"
	"github.com/klauspost/compress/zstd"
)

type CompressType uint8

const (
	CompressNone CompressType = iota
	CompressBrotli
	CompressGzip
	CompressZstd
)

// ===========================结果集对外使用的接口===============================
func CompressData(compressType CompressType, data []byte) ([]byte, error) {
	switch compressType {
	case CompressNone:
		return data, nil
	case CompressBrotli:
		return compressDataBrotli(data)
	case CompressGzip:
		return compressDataGzip(data)
	case CompressZstd:
		return compressDataZstd(data)
	default:
		return data, nil
	}
}

func DecompressData(compressType CompressType, data []byte) ([]byte, error) {
	switch compressType {
	case CompressNone:
		return data, nil
	case CompressBrotli:
		return decompressDataBrotli(data)
	case CompressGzip:
		return decompressDataGzip(data)
	case CompressZstd:
		return decompressDataZstd(data)
	default:
		return data, nil
	}
}

// 用于更改数据的压缩类型，
func ChangeCompressType(fromCompressType CompressType, toCompressType CompressType, data []byte) ([]byte, error) {
	// 如果加密类型是一致的，无需更改
	if fromCompressType == toCompressType {
		return data, nil
	}
	// 先解密数据
	decryptedData, err := DecompressData(fromCompressType, data)
	if err != nil {
		return nil, err
	}
	// 再加密数据
	return CompressData(toCompressType, decryptedData)
}

// =============================brotli====================================
func decompressDataBrotli(data []byte) ([]byte, error) {
	reader := brotli.NewReader(bytes.NewReader(data))
	var buf bytes.Buffer
	_, err := io.Copy(&buf, reader)
	return buf.Bytes(), err
}

func compressDataBrotli(data []byte) ([]byte, error) {
	var buf bytes.Buffer
	writer := brotli.NewWriterLevel(&buf, brotli.BestCompression)
	_, err := writer.Write(data)
	if err != nil {
		return nil, err
	}
	writer.Close()
	return buf.Bytes(), nil
}

// =============================Gzip====================================
func decompressDataGzip(data []byte) ([]byte, error) {
	reader, err := gzip.NewReader(bytes.NewReader(data))
	if err != nil {
		return nil, err
	}
	defer reader.Close()
	var buf bytes.Buffer
	_, err = io.Copy(&buf, reader)
	if err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func compressDataGzip(data []byte) ([]byte, error) {
	var buf bytes.Buffer
	writer, err := gzip.NewWriterLevel(&buf, gzip.BestCompression)
	if err != nil {
		return nil, err
	}
	_, err = writer.Write(data)
	if err != nil {
		writer.Close()
		return nil, err
	}
	if err = writer.Close(); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

// =============================Zstd====================================
func decompressDataZstd(data []byte) ([]byte, error) {
	reader, err := zstd.NewReader(bytes.NewReader(data))
	if err != nil {
		return nil, err
	}
	defer reader.Close()
	var buf bytes.Buffer
	_, err = io.Copy(&buf, reader)
	if err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func compressDataZstd(data []byte) ([]byte, error) {
	var buf bytes.Buffer
	writer, err := zstd.NewWriter(&buf, zstd.WithEncoderLevel(zstd.SpeedBestCompression))
	if err != nil {
		return nil, err
	}
	_, err = writer.Write(data)
	if err != nil {
		writer.Close()
		return nil, err
	}
	if err = writer.Close(); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

//=================================================================
