package database

import (
	"bytes"
	"errors"
	"fmt"
	"image"
	"image/png"
	"time"
)

func UploadImage(filename string, img image.Image) error {
	if printfulDb == nil {
		return errors.New("database is not initialized. Did you forgot to init postgre ?")
	}

	buf := bytes.Buffer{}
	err := png.Encode(&buf, img)
	if err != nil {
		return err
	}

	_, err = imagesDb.Exec(`INSERT INTO images (filename, image, created)
	VALUES ($1, $2, $3)
	ON CONFLICT (filename) DO UPDATE SET
	image = $2,
	created = $3`,
		filename,
		buf,
		time.Now().Unix(),
	)

	if err != nil {
		return fmt.Errorf("failed to insert image "+filename+" : <%w>", err)
	}

	return nil
	/*
		uploadStream, err := imagesBucket.OpenUploadStream(filename)
		if err != nil {
			return err
		}

		defer uploadStream.Close()

		buf := bytes.Buffer{}
		err = png.Encode(&buf, img)
		if err != nil {
			return err
		}

		//log.Println(buf)
		fileSize, err := uploadStream.Write(buf.Bytes())
		log.Println(fileSize, err)

		return nil
	*/
}

func GetImage(filename string) ([]byte, error) {
	return nil, nil
	/*
		downloadStream, err := imagesBucket.OpenDownloadStreamByName(filename)
		if err != nil {
			return nil, err
		}
		defer downloadStream.Close()

		p := make([]byte, downloadStream.GetFile().Length)
		if _, err = downloadStream.Read(p); err != nil {
			return nil, err
		}

		return p, nil
	*/
}
