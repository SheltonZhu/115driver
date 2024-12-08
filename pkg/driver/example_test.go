package driver

import (
	"log"
	"os"
)

func ExamplePan115Client_ImportCredential() {
	cr := &Credential{}
	if err := cr.FromCookie("UID=xxx;CID=xxxx;SEID=xxx;KID=xxx;other=xxxx"); err != nil {
		log.Fatalf("Import credentail error: %s", err)
	}
	client := Defalut().ImportCredential(cr)
	if err := client.LoginCheck(); err != nil {
		log.Fatalf("Login error: %s", err)
	}
}

func ExamplePan115Client_Download() {
	client := Defalut()

	info, err := client.Download("pickcode")
	if err != nil {
		log.Fatalf("Get download info error: %s", err)
	}
	rs, err := info.Get()
	if err != nil {
		log.Fatalf("Get io reader error: %s", err)
	}
	f, _ := os.Create("test.mp4") // save to test.mp4
	defer func() {
		f.Close()
	}()
	_, err = f.ReadFrom(rs)
	if err != nil {
		log.Fatalf("Copy reader error: %s", err)
	}
}

func ExamplePan115Client_UploadSHA1() {
	client := Defalut()

	file, err := os.Open("/path/to/file")
	if err != nil {
		log.Fatalf("Open file error: %s", err)
	}
	d, _ := client.GetDigestResult(file)
	resp, err := client.UploadSHA1(d.Size, "filename", "dirID", d.PreID, d.QuickID, file)
	if err != nil {
		log.Fatalf("Fastupload error: %s", err)
	}
	success, err := resp.Ok()
	if err != nil {
		log.Fatalf("Fastupload error: %s", err)
	}
	if !success {
		log.Printf("file is not exist, need upload")
	}
}

func ExamplePan115Client_UploadFastOrByOSS() {
	client := Defalut()

	file, err := os.Open("/path/to/file")
	if err != nil {
		log.Fatalf("Open file error: %s", err)
	}
	s, _ := file.Stat()
	err = client.UploadFastOrByOSS("dirID", s.Name(), s.Size(), file)
	if err != nil {
		log.Fatalf("Upload by oss error: %s", err)
	}
}

func ExamplePan115Client_List() {
	client := Defalut()

	files, err := client.List("dirID")
	if err != nil {
		log.Fatalf("List file error: %s", err)
	}

	for _, file := range *files {
		log.Printf("file %v", file)
	}
}

func ExamplePan115Client_Move() {
	client := Defalut()

	err := client.Move("dirID", "fileID")
	if err != nil {
		log.Fatalf("Move file error: %s", err)
	}
}

func ExamplePan115Client_Copy() {
	client := Defalut()

	err := client.Copy("dirID", "fileID")
	if err != nil {
		log.Fatalf("Copy file error: %s", err)
	}
}

func ExamplePan115Client_Delete() {
	client := Defalut()

	err := client.Delete("fileID")
	if err != nil {
		log.Fatalf("Delete file error: %s", err)
	}
}

func ExamplePan115Client_Rename() {
	client := Defalut()

	err := client.Rename("fileID", "newname")
	if err != nil {
		log.Fatalf("Rename file error: %s", err)
	}
}

func ExamplePan115Client_Mkdir() {
	client := Defalut()

	cid, err := client.Mkdir("parentID", "name")
	if err != nil {
		log.Fatalf("Make directory error: %s", err)
	}
	log.Printf("cid is  %s", cid)
}
