package controllers

import (
	"errors"
	"fmt"
	"image"
	"math"
	"os"
	"sort"

	"gocv.io/x/gocv"
)

var ErrNoImage = errors.New("no image found")

func grayscaleImageFile(file *os.File) error {
	img := gocv.IMRead(file.Name(), gocv.IMReadColor)
	defer img.Close()

	if img.Empty() {
		return ErrNoImage
	}

	grayscaleImg := gocv.NewMat()
	defer grayscaleImg.Close()

	if err := gocv.CvtColor(img, &grayscaleImg, gocv.ColorBGRToGray); err != nil {
		return err
	}

	if !gocv.IMWrite(file.Name(), grayscaleImg) {
		return gocv.LastExceptionError()
	}

	return nil
}

var (
	ErrNoContours = errors.New("unable to find contours")
	ErrCVWrite    = errors.New("unable to write file from opencv")
)

func deskewImageFile(file *os.File) error {
	img := gocv.IMRead(file.Name(), gocv.IMReadColor)
	defer img.Close()

	if img.Empty() {
		return fmt.Errorf("gocv cannot read image")
	}

	ocrReady, err := deskewMaterial(img)

	if err != nil {
		return err
	}

	defer ocrReady.Close()

	if !gocv.IMWrite(file.Name(), *ocrReady) {
		return gocv.LastExceptionError()
	}

	return nil
}

func deskewMaterial(img gocv.Mat) (*gocv.Mat, error) {
	// Copy the image to grayscale as a temporary copy for detection
	grayTmp := gocv.NewMat()
	defer grayTmp.Close()

	if err := gocv.CvtColor(img, &grayTmp, gocv.ColorBGRToGray); err != nil {
		return nil, err
	}

	blurredTmp := gocv.NewMat()
	defer blurredTmp.Close()

	if err := gocv.GaussianBlur(grayTmp, &blurredTmp, image.Point{X: 3, Y: 3}, 0, 0, gocv.BorderDefault); err != nil {
		return nil, err
	}

	threshTmp := gocv.NewMat()
	defer threshTmp.Close()

	gocv.Threshold(blurredTmp, &threshTmp, 0, 255, gocv.ThresholdBinaryInv+gocv.ThresholdOtsu)

	kernel := gocv.GetStructuringElement(gocv.MorphRect, image.Point{X: 15, Y: 5}) // Toned down kernel
	defer kernel.Close()

	dilatedTmp := gocv.NewMat()
	defer dilatedTmp.Close()

	if err := gocv.Dilate(threshTmp, &dilatedTmp, kernel); err != nil {
		return nil, err
	}

	contours := gocv.FindContours(dilatedTmp, gocv.RetrievalExternal, gocv.ChainApproxSimple)

	if contours.Size() == 0 {
		return nil, ErrNoContours
	}

	var largestContour gocv.PointVector

	maxArea := 0.0
	for i := 0; i < contours.Size(); i++ {
		c := contours.At(i)
		area := gocv.ContourArea(c)
		if area > maxArea {
			maxArea = area
			largestContour = c
		}
	}

	minRect := gocv.MinAreaRect(largestContour)
	boxPoints := minRect.Points

	warpedColor := warpPerspective(img, boxPoints)
	defer warpedColor.Close()

	ocrReady := gocv.NewMat()

	if err := gocv.CvtColor(warpedColor, &ocrReady, gocv.ColorBGRToGray); err != nil {
		return nil, err
	}

	return &ocrReady, nil
}

func warpPerspective(src gocv.Mat, pts []image.Point) gocv.Mat {
	// Order the points: top-left, top-right, bottom-right, bottom-left
	ordered := orderPoints(pts)

	tl := ordered[0]
	tr := ordered[1]
	br := ordered[2]
	bl := ordered[3]

	// Compute width of the target flattened text block
	widthA := math.Sqrt(math.Pow(float64(br.X-bl.X), 2) + math.Pow(float64(br.Y-bl.Y), 2))
	widthB := math.Sqrt(math.Pow(float64(tr.X-tl.X), 2) + math.Pow(float64(tr.Y-tl.Y), 2))
	maxWidth := int(math.Max(widthA, widthB))

	// Compute height
	heightA := math.Sqrt(math.Pow(float64(tr.X-br.X), 2) + math.Pow(float64(tr.Y-br.Y), 2))
	heightB := math.Sqrt(math.Pow(float64(tl.X-bl.X), 2) + math.Pow(float64(tl.Y-bl.Y), 2))
	maxHeight := int(math.Max(heightA, heightB))

	// Map source coordinates to destination flat-rectangle coordinates
	srcPts := gocv.NewPointVectorFromPoints(ordered)
	defer srcPts.Close()

	dstPoints := []image.Point{
		{X: 0, Y: 0},
		{X: maxWidth - 1, Y: 0},
		{X: maxWidth - 1, Y: maxHeight - 1},
		{X: 0, Y: maxHeight - 1},
	}
	dstPts := gocv.NewPointVectorFromPoints(dstPoints)
	defer dstPts.Close()

	// Compute transformation matrix and warp execution
	transformMatrix := gocv.GetPerspectiveTransform(srcPts, dstPts)
	defer transformMatrix.Close()

	warped := gocv.NewMat()
	gocv.WarpPerspective(src, &warped, transformMatrix, image.Point{X: maxWidth, Y: maxHeight})

	return warped
}

// orderPoints sorts corners consistently using sum and difference coordinates
func orderPoints(pts []image.Point) []image.Point {
	ordered := make([]image.Point, 4)

	// Sort based on X + Y sum to find Top-Left and Bottom-Right
	sort.Slice(pts, func(i, j int) bool {
		return (pts[i].X + pts[i].Y) < (pts[j].X + pts[j].Y)
	})
	ordered[0] = pts[0] // Smallest sum -> Top-Left
	ordered[2] = pts[3] // Largest sum  -> Bottom-Right

	// Sort based on Y - X difference to find Top-Right and Bottom-Left
	sort.Slice(pts, func(i, j int) bool {
		return (pts[i].Y - pts[i].X) < (pts[j].Y - pts[j].X)
	})
	ordered[1] = pts[0] // Smallest diff -> Top-Right
	ordered[3] = pts[3] // Largest diff  -> Bottom-Left

	return ordered
}
