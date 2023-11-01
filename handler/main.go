package main

import (
	"bytes"
	"crypto/tls"
	"fmt"
	"image"
	"image/jpeg"
	_ "image/png"
	"io"
	"net/http"
	"os"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	// "github.com/disintegration/imaging"
)

const (
	MB5  = "https://kcalixto-firebolt-uploads-development.s3.sa-east-1.amazonaws.com/5mb.jpg"
	MB20 = "https://kcalixto-firebolt-uploads-development.s3.sa-east-1.amazonaws.com/20mb.jpg"
)

func main() {
	defer fmt.Println("finished")

	// fileName := "brazil_id_image_front (1).png"
	// image, err := os.ReadFile(fileName)
	// if err != nil {
	// 	panic(err)
	// }

	image, err := GetImage(MB5)
	if err != nil {
		panic(err)
	}

	out, err := CompressImage(image)
	if err != nil {
		panic(err)
	}

	err = SaveLocal(fmt.Sprintf("compressed"), out)
	if err != nil {
		panic(err)
	}
}

// func main() {
// 	defer fmt.Println("finished")

// 	fileName := "brazil_id_image_front (1)"
// 	image, err := os.ReadFile(fileName)
// 	if err != nil {
// 		panic(err)
// 	}

// 	out, err := CompressImage(image, 1024)
// 	if err != nil {
// 		panic(err)
// 	}

// 	err = SaveLocal(fmt.Sprintf("zbox-9-nopng_%s", fileName), out)
// 	if err != nil {
// 		panic(err)
// 	}

// 	fileName = "brazil_id_image_back (1)"
// 	image, err = os.ReadFile(fileName)
// 	if err != nil {
// 		panic(err)
// 	}

// 	out, err = CompressImage(image, 1024)
// 	if err != nil {
// 		panic(err)
// 	}

// 	err = SaveLocal(fmt.Sprintf("zbox-9-nopng_%s", fileName), out)
// 	if err != nil {
// 		panic(err)
// 	}
// }

// func CompressImage(image []byte, newSize int) ([]byte, error) {
// 	src, err := imaging.Decode(bytes.NewReader(image), imaging.AutoOrientation(true))
// 	if err != nil {
// 		panic(err)
// 	}

// 	out := new(bytes.Buffer)
// 	resizedImage := imaging.Resize(src, newSize, 0, imaging.Box)

// 	err = imaging.Encode(out, resizedImage, imaging.PNG, imaging.PNGCompressionLevel(9))
// 	if err != nil {
// 		panic(err)
// 	}

// 	return out.Bytes(), nil
// }

func CompressImage(source []byte) ([]byte, error) {
	img, _, err := image.Decode(bytes.NewReader(source))
	if err != nil {
		panic(err)
	}

	var compressedImage bytes.Buffer
	err = jpeg.Encode(&compressedImage, img, &jpeg.Options{Quality: 65})
	if err != nil {
		panic(err)
	}

	return compressedImage.Bytes(), nil
}

func SaveInS3(filename string, imageBytes []byte) error {
	sess, err := session.NewSession(&aws.Config{
		Region: aws.String("sa-east-1"),
	})
	if err != nil {
		panic(err)
	}

	svc := s3.New(sess)

	// Upload the image to S3
	_, err = svc.PutObject(&s3.PutObjectInput{
		Bucket: aws.String("kcalixto-firebolt-uploads-development"),
		Body:   bytes.NewReader(imageBytes),
		Key:    aws.String(filename),
	})
	if err != nil {
		panic(err)
	}

	return nil
}

func SaveLocal(file string, imageBytes []byte) error {
	writer, err := os.Create(file)
	if err != nil {
		panic(err)
	}

	_, err = writer.Write(imageBytes)
	if err != nil {
		panic(err)
	}

	return nil
}

func GetImage(url string) (result []byte, err error) {
	response, err := Get(url, nil, nil)
	if err != nil {
		panic(err)
	}
	defer response.Body.Close()

	result, err = io.ReadAll(response.Body)
	if err != nil {
		panic(err)
	}

	if response.StatusCode != http.StatusOK {
		panic(err)
	}

	return result, nil
}

func Get(url string, headers *map[string]string, params *map[string]string) (response *http.Response, err error) {
	request, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return response, err
	}

	queryString := request.URL.Query()
	if params != nil {
		for k, v := range *params {
			queryString.Set(k, v)
		}

		request.URL.RawQuery = queryString.Encode()
	}

	if headers != nil {
		for k, v := range *headers {
			request.Header.Set(k, v)
		}
	}

	httpClient := http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: true,
				Renegotiation:      tls.RenegotiateOnceAsClient,
			}},
	}

	response, err = httpClient.Do(request)
	if err != nil {
		return response, err
	}

	return response, nil
}
