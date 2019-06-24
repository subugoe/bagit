package main

import (
	"flag"
	"log"
	"os"
	"time"

	"github.com/steffenfritz/bagit"
)

const version = "0.2.0"

var starttime = time.Now().Format("2006-01-02T150405")

func main() {
	b := bagit.New()

	vers := flag.Bool("version", false, "Print version")
	validate := flag.String("validate", "", "Validate bag. Expects path to bag")
	b.SrcDir = flag.String("create", "", "Create bag. Expects path to source directory")
	b.OutDir = flag.String("output", "bag_"+starttime, "Output directory for bag. Used with create flag")
	tarit := flag.Bool("tar", false, "Create a tar archive when creating a bag")
	b.HashAlg = flag.String("hash", "sha512", "Hash algorithm used for manifest file when creating a bag [sha1, sha256, sha512, md5]")
	verbose := flag.Bool("v", false, "Verbose output")
	b.AddHeader = flag.String("header", "", "Additional headers for bag-info.txt. Expects path to json file")
	b.FetchFile = flag.String("fetch", "", "Adds optional fetch file to bag. Expects path to fetch.txt file and switch manifetch")
	b.FetchManifest = flag.String("manifetch", "", "Path to manifest file for optional fetch.txt file. Mandatory if fetch switch is used")

	flag.Parse()

	if *vers {
		log.Println("Version: " + version)

		return
	}

	if len(*validate) != 0 {
		b := bagit.New()
		b.Validate(*validate, *verbose)

		return
	}

	if len(*b.SrcDir) != 0 {
		_, err := os.Stat(*b.SrcDir)
		if err != nil {
			log.Println("Cannot read source directory")
			return
		}

		_, err = os.Stat(*b.OutDir)
		if err == nil {
			log.Println("Output directory already exists. Refusing to overwrite.")
			return
		}
		// validate fetch.txt file and exit if not valid
		if len(*b.FetchFile) != 0 {
			fetchStatus := bagit.ValidateFetchFile(*b.FetchFile)
			if !fetchStatus {
				log.Println("fetch.txt file not valid. Exiting creation process.")
				return
			}

			if len(*b.FetchManifest) != 0 {
				log.Println("The usage of a fetch.txt expects a manifest file. Quitting.")
				return
			}
		}

		b.Create(*verbose)

		if *tarit {
			b.Tarit(*b.OutDir, *b.OutDir+".tar.gz")
		}

		return
	}

}
