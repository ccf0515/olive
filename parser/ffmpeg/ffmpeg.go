package ffmpeg

import (
	"io"
	"log"
	"os/exec"
	"sync"

	"github.com/luxcgo/lifesaver/parser"
)

const (
	userAgent = "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_12_6) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/59.0.3071.115 Safari/537.36"
)

func init() {
	parser.SharedManager.Register(
		new(ffmpeg),
	)
}

type ffmpeg struct {
	cmd      *exec.Cmd
	cmdStdIn io.WriteCloser

	closeOnce sync.Once
	stop      chan struct{}
}

func (p *ffmpeg) New() parser.Parser {
	return &ffmpeg{
		stop: make(chan struct{}),
	}
}

func (p *ffmpeg) Stop() {
	p.closeOnce.Do(func() {
		close(p.stop)
	})
}

func (p *ffmpeg) Type() string {
	return "ffmpeg"
}

func (p *ffmpeg) Parse(streamURL string, out string) (err error) {
	log.Println(streamURL)
	log.Println("work")
	p.cmd = exec.Command(
		"ffmpeg",
		"-nostats",
		"-progress", "-",
		"-y", "-re",
		"-user_agent", userAgent,
		// "-referer", live.GetRawUrl(),
		// "-timeout", "60000000",
		"-i", streamURL,
		"-c", "copy",
		"-bsf:a", "aac_adtstoasc",
		"-f", "flv",
		out,
	)
	if p.cmdStdIn, err = p.cmd.StdinPipe(); err != nil {
		return err
	}
	if err = p.cmd.Start(); err != nil {
		p.cmd.Process.Kill()
		return err
	}
	go func() {
		<-p.stop
		p.cmdStdIn.Write([]byte("q"))
	}()
	return p.cmd.Wait()
}