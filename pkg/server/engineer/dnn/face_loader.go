package dnn

import (
	log "github.com/sirupsen/logrus"
	"gocv.io/x/gocv"
	"gocv.io/x/gocv/contrib"
)

type InitFunc = func(ratio float64, mean gocv.Scalar, swapRGB bool, net gocv.Net, input gocv.Mat) gocv.Mat

type NameImage = map[string]string

var imgs = map[string]string{
	"renjie": "data/dnn/img/IMG_0773.HEIC.JPG",
}

type Computer struct {
	hashes   []contrib.ImgHashBase
	baseHash contrib.ImgHashBase
	stocks   map[string]gocv.Mat
}

func (c *Computer) compareAndReturnName(srcImg gocv.Mat) string {
	if c == nil || c.stocks == nil {
		return ""
	}
	for name, stock := range c.stocks {
		img := gocv.NewMat()
		srcImg.ConvertTo(&img, gocv.MatTypeCV8UC3)
		c.baseHash.Compute(img, &img)

		dist := c.baseHash.Compare(img, stock)
		similar := 1 - dist/64.0
		if similar > 0.6 {
			return name
		}
	}
	return ""
}

func LoaderInitImage(initData NameImage, dnn *Classifier) *Computer {
	var results = make(map[string]gocv.Mat)
	baseHash := contrib.AverageHash{}
	for key, path := range initData {
		img := gocv.IMRead(path, gocv.IMReadColor)
		if img.Empty() {
			log.Errorf("cannot read image %s\n", path)
			continue
		}
		img, err := dnn.RegionFace(img)
		if err != nil {
			log.Errorf("init database image %s error : %v\n", path, err)
			continue
		}
		img.ConvertTo(&img, gocv.MatTypeCV8U)
		result := gocv.NewMat()
		baseHash.Compute(img, &result)
		results[key] = result
		_ = img.Close()
	}
	return &Computer{
		hashes:   SetupHashes(),
		stocks:   results,
		baseHash: baseHash,
	}
}

func SetupHashes() []contrib.ImgHashBase {
	var hashes []contrib.ImgHashBase

	hashes = append(hashes, contrib.PHash{})
	hashes = append(hashes, contrib.AverageHash{})

	hashes = append(hashes, contrib.BlockMeanHash{})

	hashes = append(hashes, contrib.BlockMeanHash{Mode: contrib.BlockMeanHashMode1})

	hashes = append(hashes, contrib.ColorMomentHash{})

	hashes = append(hashes, contrib.NewMarrHildrethHash())

	hashes = append(hashes, contrib.NewRadialVarianceHash())

	return hashes
}
