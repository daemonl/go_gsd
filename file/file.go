package file

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"github.com/daemonl/databath"
	"github.com/daemonl/go_gsd/shared"
	"io"
	"log"
	"mime/multipart"
	"os"
)

type FileHandler struct {
	Location string
	Model    *databath.Model
}

func GetFileHandler(location string, Model *databath.Model) *FileHandler {
	fh := FileHandler{
		Location: location,
		Model:    Model,
	}
	return &fh
}
func (h *FileHandler) Upload(request shared.IRequest) {

	var functionName string
	var fileCollection string
	var collectionRef string
	var collectionId uint64

	err := request.URLMatch(&functionName, &fileCollection, &collectionRef, &collectionId)
	if err != nil {
		request.DoError(err)
		log.Println(err)
		return
	}

	_, r := request.GetRaw()
	if r.Method != "POST" && r.Method != "PUT" {
		request.WriteString("Must post a file (1)")
		return
	}

	mpr, err := r.MultipartReader()
	if err != nil {
		request.DoError(err)
		log.Println(err)
		return
	}

	var part *multipart.Part
	for {
		thisPart, err := mpr.NextPart()
		if err != nil {
			break
		}
		if thisPart.FormName() == "attachment" {
			part = thisPart
			break
		}
	}
	if part == nil {
		request.WriteString("Must post a file (2)")
		return
	}

	origName := part.FileName()

	randBytes := make([]byte, 22, 22)
	_, _ = rand.Reader.Read(randBytes)
	fileName := hex.EncodeToString(randBytes)

	file, err := os.Create(h.Location + fileName)
	if err != nil {
		log.Println(err)
		return
	}
	log.Printf("Start read into %s\n", h.Location+fileName)
	for {
		b := make([]byte, 2048, 2048)
		i, err := part.Read(b)
		log.Printf("Read %d bytes\n", i)
		if err != nil || i < 1 {
			if err != nil {
				log.Println(err)
			}
			break
		}
		file.Write(b)
	}
	file.Close()
	part.Close()
	log.Println("File Written")

	dbEntry := map[string]interface{}{
		collectionRef: collectionId,
		"file":        fileName,
		"filename":    origName,
	}

	err = h.writeDatabaseEntry(request, dbEntry, fileCollection)
	if err != nil {
		log.Println(err)
		request.DoError(err)
		return
	}

	request.WriteString(`
		<script type='text/javascript'>
		window.top.file_done()
		</script>
		Uploaded Successfully.
	`)

}
func (h *FileHandler) Download(request shared.IRequest) {

	var functionName string
	var fileCollection string
	var fileId uint64

	err := request.URLMatch(&functionName, &fileCollection, &fileId)
	if err != nil {
		request.DoError(err)
		log.Println(err)
		return
	}

	_, r := request.GetRaw()
	if r.Method != "GET" {
		request.WriteString("Must get")
		return
	}

	rqueryConditions := databath.RawQueryConditions{
		Collection: &fileCollection,
		Pk:         &fileId,
	}
	qc, _ := rqueryConditions.TranslateToQuery()

	query, err := databath.GetQuery(request.GetContext(), h.Model, qc, false)
	if err != nil {
		log.Print(err)
		request.DoError(err)
		return
	}
	sqlString, parameters, err := query.BuildSelect()
	if err != nil {
		log.Print(err)
		request.DoError(err)
		return
	}

	db, err := request.DB()
	if err != nil {
		request.DoError(err)
		return
	}

	row, err := query.RunQueryWithSingleResult(db, sqlString, parameters)
	if err != nil {
		log.Print(err)
		request.DoError(err)
		return
	}
	fn, ok := row["file"].(string)
	if !ok {

		return
	}
	origName, ok := row["filename"].(string)
	if !ok {

		return
	}
	file, err := os.Open(h.Location + fn)
	if err != nil {
		log.Print(err)
		request.DoError(err)
		return
	}
	defer file.Close()
	w, _ := request.GetRaw()
	w.Header().Add("content-disposition", "attachment; filename="+origName)

	_, err = io.Copy(w, file)
	if err != nil {
		log.Print(err)
		request.DoError(err)
		return
	}
}

func (h *FileHandler) writeDatabaseEntry(request shared.IRequest, dbEntry map[string]interface{}, fileCollection string) error {

	qc := databath.GetMinimalQueryConditions(fileCollection, "form")

	q, err := databath.GetQuery(request.GetContext(), h.Model, qc, true)
	if err != nil {
		return err
	}
	sqlString, parameters, err := q.BuildInsert(dbEntry)
	if err != nil {
		return err
	}

	db, err := request.DB()
	if err != nil {
		return err
	}

	fmt.Println(sqlString)

	res, err := db.Exec(sqlString, parameters...)

	if err != nil {
		return err
	}

	pk, _ := res.LastInsertId()
	/*
		actionSummary := &shared_structs.ActionSummary{
			UserId:     request.Session().User().Id,
			Action:     "create",
			Collection: fileCollection,
			Pk:         uint64(pk),
			Fields:     dbEntry,
		}
	*/
	createObject := map[string]interface{}{
		"collection": fileCollection,
		"id":         uint64(pk),
		"object":     dbEntry,
	}

	request.Broadcast("create", createObject)

	return nil
}
