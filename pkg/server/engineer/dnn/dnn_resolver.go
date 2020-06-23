package dnn

import (
	"errors"
	"fmt"
	"gocv.io/x/gocv"
	"image"
	"image/color"
	"path/filepath"
)

type Config struct {
	Model             string
	Config            string
	ClassifyNameImage NameImage
}

type Classifier struct {
	ratio        float64
	mean         gocv.Scalar
	swapRGB      bool
	net          gocv.Net
	descriptions []string
	computer     *Computer
}

func CreateDnnWrapper(cfg *Config) (*Classifier, error) {

	model, config := cfg.Model, cfg.Config

	backend := gocv.NetBackendDefault
	target := gocv.NetTargetCPU

	// open DNN object tracking model
	net := gocv.ReadNet(model, config)
	if net.Empty() {
		return nil, errors.New(fmt.Sprintf("Error reading network model from : %v %v\n", model, config))
	}

	_ = net.SetPreferableBackend(backend)
	_ = net.SetPreferableTarget(target)

	var ratio float64
	var mean gocv.Scalar
	var swapRGB bool

	if filepath.Ext(model) == ".caffemodel" {
		ratio = 1.0
		mean = gocv.NewScalar(104, 177, 123, 0)
		swapRGB = false
	} else {
		ratio = 1.0 / 127.5
		mean = gocv.NewScalar(127.5, 127.5, 127.5, 0)
		swapRGB = true
	}

	dnn := &Classifier{
		ratio:   ratio,
		mean:    mean,
		swapRGB: swapRGB,
		net:     net,
	}

	computer := LoaderInitImage(cfg.ClassifyNameImage, dnn)

	dnn.computer = computer
	return dnn, nil

}

func (c *Classifier) RegionFace(input gocv.Mat) (gocv.Mat, error) {
	var ratio = c.ratio
	var mean = c.mean
	var swapRGB = c.swapRGB
	var net = c.net
	img := input
	img.ConvertTo(&img, gocv.MatTypeCV32F)
	// convert image Mat to 300x300 blob that the object detector can analyze
	blob := gocv.BlobFromImage(img, ratio, image.Pt(300, 300), mean, swapRGB, false)

	// feed the blob into the detector
	net.SetInput(blob, "")

	// run a forward pass thru the network
	prob := net.Forward("")

	return regionFace(&img, prob)
}

func (c *Classifier) Detach(input gocv.Mat) gocv.Mat {
	var ratio = c.ratio
	var mean = c.mean
	var swapRGB = c.swapRGB
	var net = c.net
	img := input
	img.ConvertTo(&img, gocv.MatTypeCV32F)
	// convert image Mat to 300x300 blob that the object detector can analyze
	blob := gocv.BlobFromImage(img, ratio, image.Pt(300, 300), mean, swapRGB, false)

	// feed the blob into the detector
	net.SetInput(blob, "")

	// run a forward pass thru the network
	prob := net.Forward("")

	c.performDetection(&img, prob)

	//c.computer.compare(img)

	prob.Close()
	blob.Close()

	img.ConvertTo(&img, gocv.MatTypeCV8U)
	return img
}

func (c *Classifier) close() {
	c.net.Close()
}

// performDetection analyzes the results from the detector network,
// which produces an output blob with a shape 1x1xNx7
// where N is the number of detections, and each detection
// is a vector of float values
// [batchId, classId, confidence, left, top, right, bottom]
func (c *Classifier) performDetection(frame *gocv.Mat, results gocv.Mat) {
	for i := 0; i < results.Total(); i += 7 {
		confidence := results.GetFloatAt(0, i+2)
		if confidence > 0.8 {
			left := int(results.GetFloatAt(0, i+3) * float32(frame.Cols()))
			top := int(results.GetFloatAt(0, i+4) * float32(frame.Rows()))
			right := int(results.GetFloatAt(0, i+5) * float32(frame.Cols()))
			bottom := int(results.GetFloatAt(0, i+6) * float32(frame.Rows()))
			mask := image.Rect(left, top, right, bottom)
			gocv.Rectangle(frame, mask, color.RGBA{0, 255, 0, 0}, 2)
			if mask.Size().X > 0 && mask.Size().Y > 0 && mask.Max.X <= frame.Cols() && mask.Max.Y <= frame.Rows() {
				face := frame.Region(mask)
				name := c.computer.compareAndReturnName(face)
				gocv.PutText(frame, name, image.Pt(left, top), gocv.FontHersheyPlain, 1, color.RGBA{0, 0, 0, 255}, 2)
			}

		}
	}
}

func regionFace(frame *gocv.Mat, results gocv.Mat) (gocv.Mat, error) {
	confidence := results.GetFloatAt(0, 2)
	if confidence > 0.8 {
		left := int(results.GetFloatAt(0, 3) * float32(frame.Cols()))
		top := int(results.GetFloatAt(0, 4) * float32(frame.Rows()))
		right := int(results.GetFloatAt(0, 5) * float32(frame.Cols()))
		bottom := int(results.GetFloatAt(0, 6) * float32(frame.Rows()))
		mask := image.Rect(left, top, right, bottom)
		gocv.Rectangle(frame, mask, color.RGBA{0, 255, 0, 0}, 2)
		if mask.Size().X > 0 && mask.Size().Y > 0 && mask.Max.X <= frame.Cols() && mask.Max.Y <= frame.Rows() {
			return frame.Region(mask), nil
		}
	}

	return gocv.Mat{}, errors.New("not face found int image")
}
