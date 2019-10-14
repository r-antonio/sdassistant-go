package main

import (
	context "context"
	fmt "fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"runtime"
	"strings"
	"time"

	"google.golang.org/grpc/metadata"
)

type SharingServiceServerImpl struct {
	UnimplementedSharingServiceServer
}

func openbrowser(url string) {
	var err error

	switch runtime.GOOS {
	case "linux":
		err = exec.Command("xdg-open", url).Start()
	case "windows":
		err = exec.Command("rundll32", "url.dll,FileProtocolHandler", url).Start()
	case "darwin":
		err = exec.Command("open", url).Start()
	default:
		err = fmt.Errorf("unsupported platform")
	}
	if err != nil {
		log.Fatal(err)
	}
}

func (s *SharingServiceServerImpl) ShareLink(ctx context.Context, req *Link) (*Status, error) {
	openbrowser(req.Url)
	return &Status{
		Message: "OK",
		Code:    StatusCode_Ok,
	}, nil
}

func (s *SharingServiceServerImpl) Upload(stream SharingService_UploadServer) error {
	start := time.Now()
	md, _ := metadata.FromIncomingContext(stream.Context())
	log.Printf("%v", md)
	filename := md.Get("name")[0] + ".part"
	f, err := os.Create(filename)
	if err != nil {
		log.Fatalf("Error: %v", err)
	}
	defer f.Close()
	for {
		chunk, err := stream.Recv()
		if err != nil {
			if err == io.EOF {
				goto END
			}
			return err
		}
		log.Printf("Wow such length: %v", len(chunk.Content))
		f.Write(chunk.Content)
	}

END:
	_ = stream.SendAndClose(&Status{
		Message: "OK",
		Code:    StatusCode_Ok,
	})
	os.Rename(filename, strings.TrimSuffix(filename, ".part"))
	elapsed := time.Since(start)
	log.Printf("Time elapsed: %v", elapsed.Milliseconds())
	return nil
}
