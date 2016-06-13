package comp

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	"net/url"
	"os"
	"path"

	"github.com/h2oai/steamY/lib/fs"
)

type CompileServer struct {
	Address string
}

func NewServer(address string) *CompileServer {
	return &CompileServer{
		address,
	}
}

func (cs *CompileServer) url(p string) string {
	return (&url.URL{Scheme: "http", Host: cs.Address, Path: p}).String()
}

func (cs *CompileServer) Ping() error {
	if _, err := http.Get(cs.url("Ping")); err != nil {
		return err
	}
	return nil
}

func uploadFile(filePath, kind string, w *multipart.Writer) error {
	dst, err := w.CreateFormFile(kind, path.Base(filePath))
	if err != nil {
		return fmt.Errorf("Failed writing to buffer: %v", err)
	}
	src, err := os.Open(filePath)
	if err != nil {
		return fmt.Errorf("Failed opening file: %v", err)
	}
	if _, err := io.Copy(dst, src); err != nil {
		return fmt.Errorf("Failed copying file: %v", err)
	}

	return nil
}

func uploadJavaFiles(url, pojoPath, jarPath string) (*http.Response, error) {
	pojoPath, err := fs.ResolvePath(pojoPath)
	if err != nil {
		return nil, err
	}
	jarPath, err = fs.ResolvePath(jarPath)
	if err != nil {
		return nil, err
	}

	b := &bytes.Buffer{}
	writer := multipart.NewWriter(b)

	if err := uploadFile(pojoPath, "pojo", writer); err != nil {
		return nil, err
	}
	if err := uploadFile(jarPath, "jar", writer); err != nil {
		return nil, err
	}

	ct := writer.FormDataContentType()
	writer.Close()

	res, err := http.Post(url, ct, b)
	if err != nil {
		return nil, fmt.Errorf("Failed uploading file: %v", err)
	}

	if res.StatusCode != 200 {
		body, err := ioutil.ReadAll(res.Body)
		if err != nil {
			return nil, fmt.Errorf("Failed reading upload response: %v", err)
		}
		return nil, fmt.Errorf("Failed uploading file: %s / %s", res.Status, string(body))
	}

	return res, nil
}

func (cs *CompileServer) CompilePojo(modelPath, jarPath, servlet string) (string, error) {
	res, err := uploadJavaFiles(cs.url(servlet), modelPath, jarPath)
	if err != nil {
		return "", err
	}
	defer res.Body.Close()

	outname := path.Base(modelPath)
	outname = outname[0 : len(outname)-5]

	var p string
	d := path.Dir(modelPath)
	if servlet == "compile" {
		p = path.Join(d, outname+".jar")
	} else if servlet == "makewar" {
		p = path.Join(d, outname+".war")
	}

	p, err = fs.ResolvePath(p)
	if err != nil {
		return "", err
	}

	dst, err := os.OpenFile(p, os.O_WRONLY|os.O_CREATE, 0666)
	if err != nil {
		return "", fmt.Errorf("Download file createion failed: %s: %v", p, err)
	}
	defer dst.Close()

	if _, err := io.Copy(dst, res.Body); err != nil {
		return "", fmt.Errorf("Download file copy failed: Service to %s: %v", p, err)
	}
	return p, nil
}
