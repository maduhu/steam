package master

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path"

	"github.com/h2oai/steamY/lib/fs"
	"github.com/h2oai/steamY/master/az"
	srvweb "github.com/h2oai/steamY/srv/web"
)

type UploadHandler struct {
	az               az.Az
	workingDirectory string
	webService       srvweb.Service
}

func newUploadHandler(az az.Az, wd string, webService srvweb.Service) *UploadHandler {
	return &UploadHandler{az, wd, webService}
}

func (s *UploadHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	log.Println("File upload request received.")

	pz, azerr := s.az.Identify(r)
	if azerr != nil {
		log.Println(azerr)
		http.Error(w, fmt.Sprintf("Authentication failed: %s", azerr), http.StatusForbidden)
	}

	r.ParseMultipartForm(0)

	kind := r.FormValue("kind")

	src, handler, err := r.FormFile("file")
	if err != nil {
		log.Println("Upload form parse failed:", err)
		http.Error(w, fmt.Sprintf("Malformed request: %v", err), http.StatusBadRequest)
		return
	}
	defer src.Close()

	log.Println("Remote file: ", handler.Filename)

	fileBaseName := path.Base(handler.Filename)

	dstPath := path.Join(s.workingDirectory, fs.LibDir, kind, fileBaseName)
	if err := os.MkdirAll(path.Dir(dstPath), fs.DirPerm); err != nil {
		log.Println(err)
		http.Error(w, fmt.Sprintf("%v", err), http.StatusInternalServerError)
		return
	}

	dst, err := os.OpenFile(dstPath, os.O_WRONLY|os.O_CREATE, fs.FilePerm)
	if err != nil {
		log.Println("Upload file open operation failed:", err)
		http.Error(w, fmt.Sprintf("Error writing uploaded file to disk: %s", err), http.StatusInternalServerError)
		return
	}
	defer dst.Close()
	io.Copy(dst, src)

	if _, err := s.webService.AddEngine(pz, fileBaseName, dstPath); err != nil {
		log.Println("Failed saving engine to datastore", err)
		http.Error(w, fmt.Sprintf("Error saving engine to datastore: %v", err), http.StatusInternalServerError)
		return
	}

	log.Println("Engine uploaded:", dstPath)

}
