package engineer

import (
	"fmt"
	"github.com/hybridgroup/mjpeg"
	log "github.com/sirupsen/logrus"
	"gocv.io/x/gocv"
	"net/http"
	"org.penitence/face-service/pkg/server/engineer/dnn"
	"sync"
)

const (
	model  string = "asset/res10_300x300_ssd_iter_140000.caffemodel"
	config string = "asset/deploy.prototxt"
)

type FaceInstance struct {
	counter  int
	mutex    sync.Mutex
	resolver *dnn.Classifier
	stream   *mjpeg.Stream
	webcam   *gocv.VideoCapture
}

type errorHandle struct {
	err error
}

func (e errorHandle) ServeHTTP(writer http.ResponseWriter, request *http.Request) {
	writer.WriteHeader(403)
	writer.Write([]byte(fmt.Sprintf("open video capture fail :%v\n", e.err)))
}

var instance *FaceInstance

func (i *FaceInstance) HttpHandler(id int) http.Handler {
	i.mutex.Lock()
	defer i.mutex.Unlock()
	if i.counter != 0 {
		i.counter = i.counter + 1
		return i.stream
	} else {
		i.stream = mjpeg.NewStream()
		var err error
		i.webcam, err = gocv.OpenVideoCapture(id)
		if err != nil {
			log.Errorf("open video capture fail :%v\n", err)
			return errorHandle{
				err: err,
			}
		}
		i.counter = i.counter + 1
		go i.mjpegCapture()
		return i.stream
	}
}

func init() {
	resolver, err := dnn.CreateDnnWrapper(&dnn.Config{
		Config: config,
		Model:  model,
	})

	if err != nil {
		log.Fatalf("init engineer fail :%v", err)
	}

	instance = &FaceInstance{
		counter:  0,
		resolver: resolver,
	}
}

func GetInstance() *FaceInstance {
	return instance
}

func (i *FaceInstance) Release() {
	i.mutex.Lock()
	defer i.mutex.Unlock()
	i.counter = i.counter - 1
}

func (i *FaceInstance) mjpegCapture() {
	img := gocv.NewMat()
	defer img.Close()
	for i.counter > 0 {
		if ok := i.webcam.Read(&img); !ok {
			log.Errorf("Device closed: %v\n", 0)
			i.Release()
			return
		}
		if img.Empty() {
			continue
		}

		img = i.resolver.Detach(img)
		buf, _ := gocv.IMEncode(".jpg", img)
		i.stream.UpdateJPEG(buf)
	}
	i.mutex.Lock()
	defer i.mutex.Unlock()
	i.webcam.Close()
}
