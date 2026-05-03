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

func RecordUntil(filename string, stopCh <-chan struct{}) error {
 	const (
        sampleRate    = 48000
        channels      = 1
        bitsPerSample = 16
    )

	mic, err := StartMic()
    if err != nil {
        return err
    }

    file, err := os.Create(filename)
    if err != nil {
        return err
    }

    // Write placeholder 44-byte WAV header (will fix sizes after recording stops)
    header := make([]byte, 44)
    if _, err := file.Write(header); err != nil {
        file.Close()
        return err
    }

    fmt.Println("Recording... (press ESC to stop)")

    // Recording goroutine
    done := make(chan error, 1)
    go func() {
        buf := make([]byte, 1920)
        for {
            err := mic.ReadFrame(buf)
            if err != nil {
                done <- err
                return
            }
            if _, err := file.Write(buf); err != nil {
                done <- err
                return
            }
        }
    }()

    // Wait for stop signal
    <-stopCh

    fmt.Println("Recording stopped")

    // Kill the mic process to unblock ReadFrame
    mic.Close()

    // Wait for the recording goroutine to finish (error is expected from killed process)
    <-done

    // Determine data size (everything after the 44-byte header)
    fileInfo, err := file.Stat()
    if err != nil {
        file.Close()
        return err
    }
    dataSize := int(fileInfo.Size() - 44)

    // Seek back and write the real WAV header with correct sizes
    if _, err := file.Seek(0, 0); err != nil {
        file.Close()
        return err
    }
    if err := writeWavHeader(file, dataSize, sampleRate, channels, bitsPerSample); err != nil {
        file.Close()
        return err
    }

    file.Close()
    fmt.Println("Saved", filename)
    return nil
}
