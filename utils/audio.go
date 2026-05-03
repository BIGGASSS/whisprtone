package utils

import (
	"encoding/base64"
	"fmt"
	"io"
	"os"
	"os/exec"
	"encoding/binary"
)

type MicStream struct {
	cmd    *exec.Cmd
	stdout io.ReadCloser
}

func EncodeAudio(file string) string {
	data, err := os.ReadFile(file)
	if err != nil {
		panic(err)
	}

	encoded := base64.StdEncoding.EncodeToString(data)
	return encoded
}

func StartMic() (*MicStream, error) {
	cmd := exec.Command(
		"pw-record",
		"--format", "s16",
		"--rate", "48000",
		"--channels", "1",
		"--raw",
		"-",
	)

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return nil, err
	}

	if err := cmd.Start(); err != nil {
		return nil, err
	}

	return &MicStream{cmd: cmd, stdout: stdout}, nil
}

func (m *MicStream) ReadFrame(buf []byte) error {
	_, err := io.ReadFull(m.stdout, buf)
	return err
}

func (m *MicStream) Close() error {
	if m.cmd.Process != nil {
		_ = m.cmd.Process.Kill()
	}
	return m.cmd.Wait()
}

func writeWavHeader(file *os.File, dataSize int, sampleRate int, channels int, bitsPerSample int) error {
   byteRate := sampleRate * channels * bitsPerSample / 8
   blockAlign := channels * bitsPerSample / 8

   file.WriteString("RIFF")
   binary.Write(file, binary.LittleEndian, uint32(36+dataSize))
   file.WriteString("WAVE")

   file.WriteString("fmt ")
   binary.Write(file, binary.LittleEndian, uint32(16))
   binary.Write(file, binary.LittleEndian, uint16(1))
   binary.Write(file, binary.LittleEndian, uint16(channels))
   binary.Write(file, binary.LittleEndian, uint32(sampleRate))
   binary.Write(file, binary.LittleEndian, uint32(byteRate))
   binary.Write(file, binary.LittleEndian, uint16(blockAlign))
   binary.Write(file, binary.LittleEndian, uint16(bitsPerSample))

   file.WriteString("data")
   binary.Write(file, binary.LittleEndian, uint32(dataSize))

   return nil
}

func ConstructAudio(filename string, duration int) error {
 	const (
        sampleRate    = 48000
        channels      = 1
        bitsPerSample = 16
    )

	mic, err := StartMic()
    if err != nil {
        return err
    } else {
    	fmt.Println("Record started")
    }
    defer mic.Close()

    file, err := os.Create(filename)
    if err != nil {
        return err
    }
    defer file.Close()

    dataSize := sampleRate * channels * bitsPerSample / 8 * duration

    if err := writeWavHeader(file, dataSize, sampleRate, channels, bitsPerSample); err != nil {
    	return err
    }

    buf := make([]byte, 1920)
    bytesWritten := 0

    for bytesWritten < dataSize {
        err := mic.ReadFrame(buf)
        if err != nil {
        	return err
        }

        remaining := dataSize - bytesWritten
        if len(buf) > remaining {
           buf = buf[:remaining]
        }

        n, err := file.Write(buf)
        if err != nil {
        	return err
        }

        bytesWritten += n
    }

    fmt.Println("Record ended")
    return nil
}
