package main

import (
	"bytes"
	"crypto/tls"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/disintegration/imaging"
	"github.com/google/uuid"
)

const (
	MB5  = "https://kcalixto-firebolt-uploads-development.s3.sa-east-1.amazonaws.com/5mb.jpg"
	MB20 = "https://kcalixto-firebolt-uploads-development.s3.sa-east-1.amazonaws.com/20mb.jpg"
)

func main() {
	defer fmt.Println("finished")

	fmt.Println("started conversion 5mb image")
	image, err := GetImage(MB5)
	if err != nil {
		panic(err)
	}
	err = CompressImage(image, 1024)
	if err != nil {
		panic(err)
	}

	fmt.Println("started conversion 20mb image")
	image, err = GetImage(MB5)
	if err != nil {
		panic(err)
	}
	err = CompressImage(image, 1024)
	if err != nil {
		panic(err)
	}
}

func CompressImage(image []byte, newSize int) error {
	src, err := imaging.Decode(bytes.NewReader(image))
	if err != nil {
		panic(err)
	}

	out := new(bytes.Buffer)
	resizedImage := imaging.Resize(src, newSize, 0, imaging.Lanczos)

	err = imaging.Encode(out, resizedImage, imaging.PNG)
	if err != nil {
		panic(err)
	}

	filename := strings.Replace(uuid.New().String(), "-", "", -1) + ".png"
	return SaveInS3(filename, out.Bytes())
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
