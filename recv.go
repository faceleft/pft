package main

import (
	"encoding/binary"
	"fmt"
	"io"
	"net"
	"os"
	"path"
)

func getFiles(destDir string, conn net.Conn) error {
	defer conn.Close()
	err := checkHeaders(RCV_HEADER, conn)
	if err != nil {
		return err
	}
	recvBuf := make([]byte, BUFSIZE)
	fmt.Println("Started receiving")

	for {
		sizesBuf := [16]byte{}
		_, err = io.ReadFull(conn, sizesBuf[:])
		if err != nil {
			return err
		}

		nameSize := binary.BigEndian.Uint64(sizesBuf[0:8])
		fileSize := binary.BigEndian.Uint64(sizesBuf[8:16])

		if nameSize == 0 {
			break
		}

		nameBuf := make([]byte, nameSize)
		_, err = io.ReadFull(conn, nameBuf)
		if err != nil {
			return err
		}
		fullName := string(nameBuf)
		fileName := path.Join(destDir, path.Base(fullName))
		tmpName := fileName + ".pft_tmp"

		file, err := os.Create(tmpName)
		if err != nil {
			return err
		}
		defer os.Remove(tmpName)
		defer file.Close()

		//fmt.Printf("Getting: %v\n", fullName)

		remaining := fileSize
		percentage := int64(-1)

		for remaining > 0 {
			var msg_size int = BUFSIZE
			if remaining < uint64(BUFSIZE) {
				msg_size = int(remaining)
			}

			nRead, err := conn.Read(recvBuf[:msg_size])
			if err != nil {
				return err
			}

			_, err = file.Write(recvBuf[:nRead])
			if err != nil {
				return err
			}

			remaining -= uint64(nRead)
			if int64(100-(remaining*100)/fileSize) != percentage {
				percentage = int64(100 - (remaining*100)/fileSize)
				fmt.Print("\033[2K\r")
				printLine(fileName, float64(percentage))
			}
		}

		err = os.Rename(tmpName, fileName)
		if err != nil {
			fmt.Println(err)
		}
		fmt.Println("")
		//fmt.Printf("Done: %v\n", fullName)
	}
	fmt.Println("Finished receiving")
	return nil
}
