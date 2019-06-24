package bagit

import (
	"encoding/hex"
	"encoding/json"
	"io"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"time"
)

// Create creates a new bagit archive
func (b *Bagit) Create(verbose bool) error {

	var err error

	var mapheader map[string]interface{}
	if len(*b.AddHeader) != 0 {
		mapheader = getaddHeader(*b.AddHeader)
	}

	if *b.HashAlg == "md5" {
		log.Println("WARNING: md5 has known collisions. You should not use md5.")
		log.Println("WARNING: Press Ctrl + C to cancel or wait 5 seconds to continue...")
		time.Sleep(5 * time.Second)
	}

	// create bagit directory
	err = os.Mkdir(*b.OutDir, 0700)
	e(err)

	if verbose {
		log.Println("Created output dir:\t" + *b.OutDir)
	}

	// create payload dir
	err = os.Mkdir(*b.OutDir+"/data", 0700)
	e(err)

	// create bagit.txt tag file
	fd, err := os.Create(*b.OutDir + "/bagit.txt")
	e(err)
	defer fd.Close()

	_, err = fd.WriteString("BagIt-Version: " + BagitVer + "\n")
	e(err)
	_, err = fd.WriteString("Tag-File-Character-Encoding: " + TagFileCharEnc)
	e(err)

	if verbose {
		log.Println("Created bagit.txt file")
	}

	// create manifest-ALG.txt file
	fm, err := os.Create(*b.OutDir + "/manifest-" + *b.HashAlg + ".txt")
	e(err)
	defer fm.Close()

	// create bag-info.txt file
	fi, err := os.Create(*b.OutDir + "/bag-info.txt")
	e(err)
	defer fi.Close()

	// copy source to data dir in new bag, calculate oxum and count bytes of payload
	err = filepath.Walk(*b.SrcDir, func(path string, info os.FileInfo, err error) error {
		if !info.IsDir() {
			b.Oxum.Filecount++
			fsize, err := os.Stat(path)
			e(err)
			b.Oxum.Bytes += fsize.Size()
			_, err = fm.WriteString(hex.EncodeToString(hashit(path, *b.HashAlg)) + " " + path + "\n")
			copy(path, *b.OutDir+"/data/"+path)

		} else {
			os.MkdirAll(*b.OutDir+"/data/"+path, 0700)
		}
		return nil
	})
	e(err)

	if len(*b.FetchFile) != 0 {
		// check if file exists
		_, err = os.Stat(*b.FetchFile)
		e(err)
		// copy fetch file to bag
		src, err := os.Open(*b.FetchFile)
		e(err)
		defer src.Close()

		dst, err := os.Create(*b.OutDir + "/fetch.txt")
		e(err)
		defer dst.Close()

		_, err = io.Copy(dst, src)
		e(err)

		if verbose {
			log.Println("Copied cetch.txt file to file")
		}

		// concatenate fetch manifest to manifest
		// iff all entries in fetch.txt have a length, add to oxum else warning
		//
	}

	if verbose {
		log.Println("Created manifest file")
	}

	// write bag-info.txt
	oxumbytes := int(b.Oxum.Bytes)
	_, err = fi.WriteString("Bag-Software-Agent: bagit <https://github.com/steffenfritz/bagit>\n")
	_, err = fi.WriteString("Bagging-Date: " + b.Timestamp + "\n")
	_, err = fi.WriteString("Payload-Oxum: " + strconv.Itoa(oxumbytes) + "." + strconv.Itoa(b.Oxum.Filecount) + "\n")

	if verbose {
		log.Println("Created bag-info.txt file")
	}

	// add additional headers to bag-info.txt
	if len(mapheader) != 0 {
		for k, v := range mapheader {
			_, err = fi.WriteString(k + ": " + v.(string) + "\n")
		}
		if verbose {
			log.Println("Added additional headers to bag-info.txt")
		}
	}

	// ToDo: write optional tagmanifest files

	return err
}

// getaddHeader gets additional headers from a json file
func getaddHeader(addHeader string) map[string]interface{} {
	jsonFile, err := os.Open(addHeader)
	e(err)
	defer jsonFile.Close()

	byteValue, _ := ioutil.ReadAll(jsonFile)

	var result map[string]interface{}
	json.Unmarshal([]byte(byteValue), &result)

	return result
}
