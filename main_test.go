package main

import (
	"bytes"
	"fmt"
	"math/rand"
	"net"
	"os"
	"testing"

	"github.com/klauspost/compress/zstd"
)

func Test_SendAndReceive(t *testing.T) {
	dirIn, err := os.MkdirTemp(".", "test_in")
	if err != nil {
		panic(err)
	}
	defer os.RemoveAll(dirIn)

	dirOut, err := os.MkdirTemp(".", "test_out")
	if err != nil {
		panic(err)
	}
	defer os.RemoveAll(dirOut)

	const fileNum = 100
	const maxSizeB = 1024 * 1024 * 10 //10Mb

	inFileNames := [fileNum]string{}
	outFileNames := [fileNum]string{}

	for i := 0; i < fileNum; i++ {
		file, err := os.CreateTemp(dirIn, "rndfile")
		if err != nil {
			panic(err)
		}
		b := make([]byte, rand.Int()%maxSizeB)
		_, err = rand.Read(b)
		if err != nil {
			panic(err)
		}
		_, err = file.Write(b)
		if err != nil {
			panic(err)
		}
		inFileNames[i] = file.Name()
		outFileNames[i] = file.Name()
		file.Close()
	}

	outConn, inConn := net.Pipe()
	fmt.Println("start testing")

	//sendFiles(inFileNames[:], inConn)
	errChn := make(chan error)

	go func() {
		err := sendFiles(inFileNames[:], inConn, 1024*1024)
		errChn <- err
	}()

	rErr := getFiles(dirOut, outConn, 1024*1024)
	if rErr != nil {
		t.Error(rErr)
		return
	}

	sErr := <-errChn
	if sErr != nil {
		t.Error(sErr)
		return
	}

	for i := 0; i < fileNum; i++ {
		f1, err := os.ReadFile(inFileNames[i])
		if err != nil {
			panic(err)
		}
		f2, err := os.ReadFile(outFileNames[i])
		if err != nil {
			panic(err)
		}
		if !bytes.Equal(f1, f2) {
			t.Errorf("The files are different after the transfer")
		}
	}

}

func Test_SendAndReceiveZstd(t *testing.T) {
	dirIn, err := os.MkdirTemp(".", "test_in")
	if err != nil {
		panic(err)
	}
	defer os.RemoveAll(dirIn)

	dirOut, err := os.MkdirTemp(".", "test_out")
	if err != nil {
		panic(err)
	}
	defer os.RemoveAll(dirOut)

	const fileNum = 100
	const maxSizeB = 1024 * 1024 * 10 //10Mb

	inFileNames := [fileNum]string{}
	outFileNames := [fileNum]string{}

	for i := 0; i < fileNum; i++ {
		file, err := os.CreateTemp(dirIn, "rndfile")
		if err != nil {
			panic(err)
		}
		b := make([]byte, rand.Int()%maxSizeB)
		_, err = rand.Read(b)
		if err != nil {
			panic(err)
		}
		_, err = file.Write(b)
		if err != nil {
			panic(err)
		}
		inFileNames[i] = file.Name()
		outFileNames[i] = file.Name()
		file.Close()
	}

	outConn, inConn := net.Pipe()
	fmt.Println("start testing")

	//sendFiles(inFileNames[:], inConn)
	errChn := make(chan error)

	go func() {
		zstdConn, err := zstd.NewWriter(inConn, zstd.WithEncoderLevel(zstd.SpeedFastest))
		if err != nil {
			errChn <- err
			return
		}
		fmt.Println("Use zstd")

		err = sendFiles(inFileNames[:], zstdConn, 1024*1024)
		zstdConn.Close()
		errChn <- err
	}()

	zstdConn, err := zstd.NewReader(outConn)
	if err != nil {
		t.Error(err)
		return
	}
	 
	fmt.Println("Use zstd")
	defer func() { go zstdConn.Close() }() //may be blocked

	rErr := getFiles(dirOut, zstdConn, 1024*1024)
	if rErr != nil {
		t.Error(rErr)
		return
	}
	sErr := <-errChn
	if sErr != nil {
		t.Error(sErr)
		return
	}

	for i := 0; i < fileNum; i++ {
		f1, err := os.ReadFile(inFileNames[i])
		if err != nil {
			panic(err)
		}
		f2, err := os.ReadFile(outFileNames[i])
		if err != nil {
			panic(err)
		}
		if !bytes.Equal(f1, f2) {
			t.Errorf("The files are different after the transfer")
		}
	}

}
