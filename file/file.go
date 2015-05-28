package file

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"io/ioutil"
	"log"
	"mime/multipart"

	"github.com/daemonl/databath"
	"github.com/daemonl/go_gsd/shared"
	"github.com/mitchellh/goamz/aws"
	"github.com/mitchellh/goamz/s3"
)

type FileHandler struct {
	BucketName string
	Path       string
	Model      *databath.Model
}

func GetFileHandler(bucketName string, path string, Model *databath.Model) *FileHandler {
	fh := FileHandler{
		BucketName: bucketName,
		Path:       path,
		Model:      Model,
	}
	return &fh
}

func (h *FileHandler) getBucket() (*s3.Bucket, error) {
	auth, err := aws.EnvAuth()
	if err != nil {
		return nil, err
	}
	client := s3.New(auth, aws.APSoutheast2)
	bucket := client.Bucket(h.BucketName)
	return bucket, nil
}

func (h *FileHandler) Upload(request shared.IRequest) error {

	var functionName string
	var fileCollection string
	var collectionRef string
	var collectionId uint64

	err := request.URLMatch(&functionName, &fileCollection, &collectionRef, &collectionId)
	if err != nil {
		return err
	}

	_, r := request.GetRaw()
	if r.Method != "POST" && r.Method != "PUT" {
		request.WriteString("Must post a file (1)")
		return nil
	}

	mpr, err := r.MultipartReader()
	if err != nil {
		return err
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
		return nil
	}

	origName := part.FileName()

	randBytes := make([]byte, 22, 22)
	_, _ = rand.Reader.Read(randBytes)
	fileName := hex.EncodeToString(randBytes)

	bucket, err := h.getBucket()
	if err != nil {
		return err
	}

	upload, err := ioutil.ReadAll(part)
	if err != nil {
		return err
	}
	err = bucket.Put(h.Path+fileName, upload, "application/octet-stream", s3.Private)
	if err != nil {
		return err
	}
	log.Println("File Written")

	dbEntry := map[string]interface{}{
		collectionRef: collectionId,
		"file":        fileName,
		"filename":    origName,
	}

	err = h.writeDatabaseEntry(request, dbEntry, fileCollection)
	if err != nil {
		return err
	}

	request.WriteString(`
		<script type='text/javascript'>
		window.top.file_done()
		</script>
		Uploaded Successfully.
	`)
	return nil
}
func (h *FileHandler) Download(request shared.IRequest) error {
	return nil
	/*
		var functionName string
		var fileCollection string
		var fileId uint64

		err := request.URLMatch(&functionName, &fileCollection, &fileId)
		if err != nil {
			return err
		}

		_, r := request.GetRaw()
		if r.Method != "GET" {
			request.WriteString("Must get")
			return nil
		}

		rqueryConditions := databath.RawQueryConditions{
			Collection: &fileCollection,
			Pk:         &fileId,
		}
		qc, _ := rqueryConditions.TranslateToQuery()

		query, err := databath.GetQuery(request.GetContext(), h.Model, qc, false)
		if err != nil {
			return err
		}
		sqlString, parameters, err := query.BuildSelect()
		if err != nil {
			return err
		}

		db, err := request.DB()
		if err != nil {
			return err
		}

		row, err := query.RunQueryWithSingleResult(db, sqlString, parameters)
		if err != nil {
			return err
		}
		fn, ok := row["file"].(string)
		if !ok {

			return nil
		}
		origName, ok := row["filename"].(string)
		if !ok {

			return nil
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
	*/
	return nil
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
