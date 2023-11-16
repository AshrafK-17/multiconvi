package main

import (
	"archive/zip"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"

	log "github.com/sirupsen/logrus"
)

func main() {
	http.HandleFunc("/", convertImageHandler)
	http.ListenAndServe(":8080", nil)
}

// Limit the content size of the request to 2GB
const (
	maxRequestSize = int64(2 << 30)
)

// Supported Formats
var supportedImageFormats = []string{
	".aai",
	".apng",
	".art",
	".arw",
	".avi",
	".avif",
	".avs",
	".bayer",
	".bmp",
	".bmp2",
	".bmp3",
	".bpg",
	".brf",
	".cals",
	".cin",
	".cip",
	".clipboard",
	".cmyk",
	".cmyka",
	".cr2",
	".crw",
	".cube",
	".cur",
	".cut",
	".dcm",
	".dcr",
	".dcx",
	".dds",
	".debug",
	".dib",
	".djvu",
	".dng",
	".dot",
	".dpx",
	".emf",
	".epdf",
	".epi",
	".eps",
	".eps2",
	".eps3",
	".epsf",
	".epsi",
	".ept",
	".exr",
	".farbfeld",
	".fax",
	".fits",
	".fl32",
	".flif",
	".fpx",
	".ftxt",
	".gif",
	".gplt",
	".gray",
	".graya",
	".hdr",
	".hdr",
	".heic",
	".hpgl",
	".hrz",
	".html",
	".ico",
	".info",
	".isobrl",
	".isobrl6",
	".j2c",
	".j2k",
	".jbig",
	".jng",
	".jp2",
	".jpeg",
	".jpg",
	".jpt",
	".json",
	".jxl",
	".jxr",
	".kernel",
	".m2v",
	".man",
	".mat",
	".miff",
	".mng",
	".mono",
	".mpc",
	".mpeg",
	".mpr",
	".mrsid",
	".mrw",
	".msl",
	".mtv",
	".mvg",
	".nef",
	".ora",
	".orf",
	".otb",
	".p7",
	".palm",
	".pam",
	".pbm",
	".pcd",
	".pcds",
	".pcl",
	".pcx",
	".pdb",
	".pdf",
	".pef",
	".pes",
	".pfa",
	".pfb",
	".pfm",
	".pgm",
	".phm",
	".picon",
	".pict",
	".pix",
	".png",
	".png00",
	".png24",
	".png32",
	".png48",
	".png64",
	".png8",
	".pnm",
	".pocketmod",
	".ppm",
	".ps",
	".ps2",
	".ps3",
	".psb",
	".psd",
	".ptif",
	".pwp",
	".qoi",
	".rad",
	".raf",
	".raw",
	".rgb",
	".rgb565",
	".rgba",
	".rgf",
	".rla",
	".rle",
	".sct",
	".sfw",
	".sgi",
	".shtml",
	".sid",
	".sparse-color",
	".strimg",
	".sun",
	".svg",
	".text",
	".tga",
	".tiff",
	".tim",
	".ttf",
	".txt",
	".ubrl",
	".ubrl6",
	".uil",
	".uyvy",
	".vicar",
	".video",
	".viff",
	".wbmp",
	".wdp",
	".webp",
	".wmf",
	".wpg",
	".x",
	".x3f",
	".xbm",
	".xcf",
	".xpm",
	".xwd",
	".yaml",
	".ycbcr",
	".ycbcra",
	".yuv",
}

var wg sync.WaitGroup

//outputChan := make(chan *os.File)

func convertImageHandler(w http.ResponseWriter, r *http.Request) {
	defer trackTime(time.Now())
	err := r.ParseMultipartForm(10 << 20)
	if err != nil {
		log.Error("Error", err)
		http.Error(w, "Error   parsing form", http.StatusBadRequest)
	}

	if r.Method != http.MethodPost {
		log.Errorf("Method   %s not   allowed", r.Method)
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	r.Body = http.MaxBytesReader(w, r.Body, maxRequestSize)

	if r.ContentLength > maxRequestSize {
		log.Error("File   size is too large")
		w.WriteHeader(http.StatusRequestEntityTooLarge)
		return
	}

	if len(r.MultipartForm.File) == 0 || r.FormValue("outputFormat") == "" {
		log.Error("Missing   Input Parameters")
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	files := r.MultipartForm.File["inputFile"]
	if len(files) == 0 {
		log.Error("No   files found in request")
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	_, fileheader, _ := r.FormFile("inputFile")

	outputFormat := r.FormValue("outputFormat")

	sort.Slice(files, func(i, j int) bool {
		return files[i].Filename < files[j].Filename
	})

	convertedFiles := make([]*os.File, 0)
	outputChan := make(chan *os.File, len(files))

	for i := range files {
		wg.Add(1)
		go convertAndSend(w, files[i], outputFormat, outputChan)

	}

	wg.Wait()

	close(outputChan)

	if len(files) == 1 {
		convertedFile := <-outputChan

		log.Info("single   file found")
		fileNametr := strings.ReplaceAll(fileheader.Filename, "   ", "")
		fileName := fmt.Sprintf(fileNametr[:len(fileNametr)-len(filepath.Ext(fileNametr))] + outputFormat)

		// Set the response header to indicate that the   response contains an image file
		w.Header().Set("Content-Type", fmt.Sprintf("image/%s", strings.TrimPrefix(outputFormat, ".")))
		w.Header().Set("Content-Disposition", fmt.Sprintf("attachment;   filename=%s", fileName))
		w.Header().Set("Access-Control-Expose-Headers", "Content-Disposition")

		// Copy the converted file to the response writer to   send it as a response
		_, err = io.Copy(w, convertedFile)
		if err != nil {
			log.Error("Error   copying converted file", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	} else if len(files) > 1 {
		for convertedFile := range outputChan {
			convertedFiles = append(convertedFiles, convertedFile)
		}

		output := "done.zip"
		log.Info("multiple   files found")

		//create a zip file and add all converted files
		file := []string{}
		for i := range convertedFiles {
			file = append(file, convertedFiles[i].Name())
		}

		if err := ZipFiles(output, file); err != nil {
			panic(err)
		}
		fmt.Println("Zipped   File:", output)

		convertedZipfile, err := os.Open(output)
		if err != nil {
			log.Error(err)
		}
		defer convertedZipfile.Close()

		zipfileName := "convertedimages.zip"

		w.Header().Set("Content-Type", "application/zip")
		w.Header().Set("Content-Disposition", fmt.Sprintf("attachment;   filename=%s", zipfileName))
		w.Header().Set("Access-Control-Expose-Headers", "Content-Disposition")

		_, err = io.Copy(w, convertedZipfile)
		if err != nil {
			log.Error("Error   copying converted file", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}
	log.Info("converted   successfully")
}

func inputSupportedFormat(filename string) bool {
	if filepath.Ext(filename) == "" {
		log.Error("Extension   not present")
		return false
	}
	ext := strings.ToLower(filepath.Ext(filename))
	index := sort.SearchStrings(supportedImageFormats, ext)
	return index < len(supportedImageFormats) && supportedImageFormats[index] == ext
}

func outputSupportedFormat(filename string, outputformat string) bool {
	if filepath.Ext(filename) == outputformat {
		log.Error("Input   format cannot be same as output format")
		return false
	}
	index := sort.SearchStrings(supportedImageFormats, outputformat)
	return index < len(supportedImageFormats) && supportedImageFormats[index] == outputformat
}

// https://golangcode.com/create-zip-files-in-go/
func ZipFiles(filename string, files []string) error {

	newZipFile, err := os.Create(filename)
	if err != nil {
		log.Error(err)
		return err
	}
	defer newZipFile.Close()

	zipWriter := zip.NewWriter(newZipFile)
	defer zipWriter.Close()

	// Add files to zip
	for _, file := range files {
		log.Info(file)
		if err = AddFileToZip(zipWriter, file); err != nil {
			log.Error("Failed   to add files in zip ", err)
			return err
		}
	}
	return nil
}

func AddFileToZip(zipWriter *zip.Writer, filename string) error {

	fileToZip, err := os.Open(filename)
	if err != nil {
		log.Error("Failed   to open file ", filename)
		return err
	}
	defer fileToZip.Close()

	// Get the file information
	info, err := fileToZip.Stat()
	if err != nil {
		log.Error("Failed   to get file info")
		return err
	}

	header, err := zip.FileInfoHeader(info)
	if err != nil {
		log.Error("Failed   to get file info header")
		return err
	}

	// Using FileInfoHeader() above only uses the basename   of the file. If we want
	// to preserve the folder structure we can overwrite   this with the full path.
	header.Name = filepath.Base(filename)
	log.Info(header.Name)

	// Change to deflate to gain better compression
	// see http://golang.org/pkg/archive/zip/#pkg-constants
	header.Method = zip.Deflate

	writer, err := zipWriter.CreateHeader(header)
	if err != nil {
		log.Error("Failed   to add file header into zipwriter")
		return err
	}
	log.Info("zip   archive success")
	_, err = io.Copy(writer, fileToZip)
	return err

}

func convertAndSend(w http.ResponseWriter, inputFile *multipart.FileHeader, outputFormat string, outputChan chan *os.File) {
	defer wg.Done()
	file, err := inputFile.Open()
	if err != nil {
		log.Error("Error   getting file from request", err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	defer file.Close()
	if !inputSupportedFormat(inputFile.Filename) {
		log.Error("Input   format not supported")
		http.Error(w, err.Error(), http.StatusBadRequest)
	}

	if !outputSupportedFormat(inputFile.Filename, outputFormat) {
		log.Error("Output   format not supported")
		http.Error(w, err.Error(), http.StatusBadRequest)
	}

	fileName := strings.ReplaceAll(inputFile.Filename, "   ", "")

	tempFile, err := os.CreateTemp("", fmt.Sprintf("%s-*%s", fileName[:len(fileName)-len(path.Ext(fileName))], filepath.Ext(fileName)))
	if err != nil {
		log.Error("Error   creating temp file", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer tempFile.Close()
	defer os.Remove(tempFile.Name())
	log.Info("tempfile   name: ", tempFile.Name())

	_, err = io.Copy(tempFile, file)
	if err != nil {
		log.Error("Error   writing uploade file to temp file", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	inputFilePath := tempFile.Name()

	// Create the output file path based on the input file   path and output format
	outputFileName := strings.TrimSuffix(filepath.Base(inputFilePath), filepath.Ext(inputFilePath)) + outputFormat
	outputFilePath := filepath.Join(filepath.Dir(inputFilePath), outputFileName)

	//Convert to ico format
	if outputFormat == ".ico" {
		cmd := exec.Command("magick", inputFilePath, "-alpha", "off", "-resize", "256x256", outputFilePath)
		err = cmd.Run()
		if err != nil {
			log.Error("Error   using convert", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	} else {
		// Convert the input file to the specified output   format using ImageMagick
		cmd := exec.Command("convert", inputFilePath, outputFilePath)
		err = cmd.Run()
		if err != nil {
			log.Error("Error   using convert", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}

	convertedFile, err := os.Open(outputFilePath)
	if err != nil {
		log.Error("Error   opening converted file", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer convertedFile.Close()
	outputChan <- convertedFile
}

func trackTime(pre time.Time) {
	elapsed := time.Since(pre)
	log.Println("elapsed:", elapsed)
}
