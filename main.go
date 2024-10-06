package main

import (
	"archive/zip"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"runtime"
)

var Hash map[string]string= make(map[string]string)

func printMemStats() {
	var memStats runtime.MemStats
	runtime.ReadMemStats(&memStats)

	fmt.Printf("Alloc = %v MiB", bToMb(memStats.Alloc))
	fmt.Printf("\tTotalAlloc = %v MiB", bToMb(memStats.TotalAlloc))
	fmt.Printf("\tSys = %v MiB", bToMb(memStats.Sys))
	fmt.Printf("\tNumGC = %v\n", memStats.NumGC)
}

func bToMb(b uint64) uint64 {
	return b / 1024 / 1024
}


func HandleUpload(w http.ResponseWriter, r *http.Request){
  toBe := make([]string, 0)
  key := r.FormValue("key") 
  const maxMemory = 1 * 1024 * 1024 * 1024 // 1 GB
  err := r.ParseMultipartForm(maxMemory)
  if err != nil{
    http.Error(w, "Unable to parse form: "+err.Error(), http.StatusBadRequest)
    return
  }
  
  files := r.MultipartForm.File["files"] 

  zipFile ,err := os.Create("./uploads/"+key+".zip")
  if err != nil{
    fmt.Println("error creating a new file")
    return 
  }
  defer zipFile.Close()
  zipw := zip.NewWriter(zipFile)
  defer zipw.Close() 

  var response string

  for _, fileHeader := range files {
    file , err := fileHeader.Open()
    if err != nil{
      fmt.Println("1" , err)
    } 
    defer file.Close()
    addFile(file,fileHeader.Filename, zipw) 
    response += fmt.Sprintf("Uploaded : %s<br>" , fileHeader.Filename); 
    toBe= append(toBe, fileHeader.Filename)
  }
  fmt.Fprint(w ,response)
  Hash[key]=key+".zip"

}

func addFile(file multipart.File , fileName string , zipw *zip.Writer){
  

  wr ,err := zipw.Create(fileName)
  if err != nil{
    fmt.Println(err)
    return
  }
  if _ ,err := io.Copy(wr ,file ) ; err!= nil {
    fmt.Println(err)
    return 
  }

}
func HandleDownload(w http.ResponseWriter, r *http.Request){
  key := r.FormValue("key")
  fileName , ok := Hash[key]
  if !ok {
    http.Error(w, "Resource not found", http.StatusNotFound)
    return
  }
  file, err := os.Open("./uploads/"+fileName)
  if err != nil{
    fmt.Println(err)
    http.Error(w,"cannot open file" , 500)
    return
  }

  defer file.Close()
  fileStat , _ := file.Stat()

  w.Header().Set("Content-Disposition", "attachment; filename="+file.Name())
  w.Header().Set("Content-Type", "application/octet-stream")

  http.ServeContent(w ,r ,fileStat.Name() ,fileStat.ModTime() , file)

}
// func tarify(toBe []string, key string) {
//   // all , err := os.Create("./uploads/all.tar") 
//   // if err != nil{
//   //   fmt.Println("1" , err)
//   //   return 
//   // }
//   var buff bytes.Buffer
//   tw:= tar.NewWriter(&buff)
//   defer func() {
//         if err := tw.Close(); err != nil {
//             fmt.Println("Error closing tar writer:", err)
//         }
//     }()
//   for _ ,fileName := range toBe {
//     file , err := os.Open("./uploads/"+fileName)
//     if err != nil{
//       fmt.Println("2", err)
//       return 
//     }
//     defer func()  {
//       file.Close()
//       err:= os.Remove("./uploads/"+fileName)
//       if err != nil{
//         fmt.Println(err)
//       } 
//     }() 
//     stat , _:= file.Stat()
//     header , err := tar.FileInfoHeader(stat ,"")
//     header.Name = fileName
//     if err != nil{
//       fmt.Println("3" , err)
//       return  
//     }
//     if err:= tw.WriteHeader(header) ; err!= nil {
//       fmt.Println("4" , err)
//       return 
//     }
//
//     if _, err := io.Copy(tw , file); err!= nil {
//       fmt.Println("6" , err)
//       return 
//     }
//
//   }
//   var xx bytes.Buffer
//   gz , _ := os.Create("./uploads/all.tar.gz")
//   defer gz.Close()
//   gw := gzip.NewWriter(&xx)
//   gw.Write(buff.Bytes()) 
//   gz.Write(xx.Bytes())
//
// }
//
// func gzipify(tarfile *os.File)  {
//   x , _ := io.ReadAll(tarfile)
//   fmt.Println(len(x))
//   file , err := os.Create("./uploads/all.tar.gz")
//   if err != nil{
//     fmt.Println(err)
//   }
//   defer file.Close()
//   gw := gzip.NewWriter(file)
//   defer gw.Flush()
//   defer gw.Close()
//   if _, err := io.Copy(gw, tarfile); err != nil {
// 		fmt.Println("Error copying to gzip:", err)
// 	}
// }
//

func main() {
    // Set the directory from which to serve files
    fs := http.FileServer(http.Dir("./views"))

    // Serve the files with a URL prefix
    http.Handle("/static/", http.StripPrefix("/static/", fs))
    
    http.HandleFunc("/" ,func(w http.ResponseWriter, r *http.Request) {
      http.ServeFile(w ,r ,"./views/index.html")
    })
    http.HandleFunc("/wtvr", HandleUpload)

    http.HandleFunc("/download", HandleDownload)
    // Start the HTTP server
    http.ListenAndServe(":3000", nil)
}
